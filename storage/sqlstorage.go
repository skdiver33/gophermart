package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/golang-migrate/migrate/v4"
	balance "github.com/skdiver33/gophermart/internal/balance"
	order "github.com/skdiver33/gophermart/internal/order"
	user "github.com/skdiver33/gophermart/internal/user"
	withdraw "github.com/skdiver33/gophermart/internal/withdraw"
)

type SQLStorage struct {
	config *SQLStorageConfig
	db     *sql.DB
}

type SQLStorageConfig struct {
	DBAddress string
}

func NewSQLStorageConfig(address string) *SQLStorageConfig {
	storageConfig := SQLStorageConfig{DBAddress: address}
	return &storageConfig
}

func NewSQLStorage(address string) (*SQLStorage, error) {
	newStorage := SQLStorage{}
	newStorage.config = NewSQLStorageConfig(address)
	err := newStorage.InitializeConnection()
	if err != nil {
		return nil, err
	}

	migrator, err := NewMigrator(newStorage.db)
	if err != nil {
		return nil, err
	}
	err = migrator.ApplyMigrations("up")
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}
	err = newStorage.InitializeConnection()
	if err != nil {
		return nil, err
	}
	return &newStorage, nil
}

func (storage *SQLStorage) InitializeConnection() error {
	db, err := sql.Open("pgx", storage.config.DBAddress)
	if err != nil {
		log.Println("error open connection to DB")
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	if err = db.PingContext(ctx); err != nil {
		cancel()
		return err
	}
	defer cancel()
	storage.db = db
	return nil
}

func (storage *SQLStorage) CloseAndClean() error {
	migrator, err := NewMigrator(storage.db)
	if err != nil {
		return err
	}
	err = migrator.ApplyMigrations("down")
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	storage.db.Close()
	return nil
}

func (storage *SQLStorage) CloseConnection() {
	storage.db.Close()
}

func (storage *SQLStorage) AddUser(ctx context.Context, user *user.User) (int, error) {
	id := -1
	err := storage.db.QueryRowContext(ctx, "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING user_id", user.Login, user.Password).Scan((&id))
	if err != nil {
		return -1, errors.New("error get inserted id")
	}
	return int(id), nil
}
func (storage *SQLStorage) GetUser(ctx context.Context, login string, password string) (*user.User, error) {

	user := user.User{}
	row := storage.db.QueryRowContext(ctx, "SELECT * FROM users WHERE login=$1", login)
	err := row.Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (storage *SQLStorage) AddOrder(ctx context.Context, order *order.Order) error {
	_, err := storage.db.ExecContext(ctx, "INSERT INTO orders (order_number,user_id, status,accrual,upload_data) VALUES ($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING", order.Number, order.UserID, order.Status, order.Accrual, order.UploadData.Format(time.RFC3339))
	if err != nil {
		return errors.New("error inserted new order to DB")
	}
	return nil
}

func (storage *SQLStorage) GetOrder(ctx context.Context, number string) (*order.Order, error) {
	order := order.Order{}
	row := storage.db.QueryRowContext(ctx, "SELECT * FROM orders WHERE order_number=$1", number)
	var timeStr string
	err := row.Scan(&order.Number, &order.UserID, &order.Status, &order.Accrual, &timeStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	order.UploadData, err = time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, errors.New("error parse data string from  DB")
	}
	return &order, nil
}

func (storage *SQLStorage) GetAllOrderForID(ctx context.Context, id int) ([]order.Order, error) {

	rows, err := storage.db.QueryContext(ctx, "SELECT order_number,status,accrual,upload_data FROM orders where user_id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("error get all orders for user %w", err)
	}
	defer rows.Close()

	result := make([]order.Order, 0)

	for rows.Next() {
		curOrder := order.Order{}
		var upload string
		if err := rows.Scan(&curOrder.Number, &curOrder.Status, &curOrder.Accrual, &upload); err != nil {

			return nil, fmt.Errorf("error parse result from DB %w", err)
		}
		curOrder.UploadData, err = time.Parse(time.RFC3339, upload)
		if err != nil {
			return nil, errors.New("error parse data string from  DB")
		}
		result = append(result, curOrder)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error load data from DB %s", err.Error())
	}

	return result, nil
}

func (storage *SQLStorage) GetUnprocOrders(ctx context.Context) ([]order.Order, error) {

	rows, err := storage.db.QueryContext(ctx, "SELECT order_number,user_id,status,accrual,upload_data FROM orders where status in ('NEW','PROCESSING')")
	if err != nil {
		return nil, fmt.Errorf("error get all orders for user %w", err)
	}
	defer rows.Close()

	result := make([]order.Order, 0)

	for rows.Next() {
		curOrder := order.Order{}
		var upload string
		if err := rows.Scan(&curOrder.Number, &curOrder.UserID, &curOrder.Status, &curOrder.Accrual, &upload); err != nil {

			return nil, fmt.Errorf("error parse result from DB %w", err)
		}
		curOrder.UploadData, err = time.Parse(time.RFC3339, upload)
		if err != nil {
			return nil, errors.New("error parse data string from  DB")
		}
		result = append(result, curOrder)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error get unprocessed order from DB %s", err.Error())
	}

	return result, nil
}

func (storage *SQLStorage) UpdateOrderStatus(ctx context.Context, orderNumber string, newStatus string, accrual float32) error {

	_, err := storage.db.ExecContext(ctx, "UPDATE orders SET status = $2, accrual = $3 WHERE order_number = $1", orderNumber, newStatus, accrual)
	if err != nil {
		return fmt.Errorf("error update order status %w", err)
	}
	return nil
}

func (storage *SQLStorage) AddWithdraw(ctx context.Context, withdraw *withdraw.Withdraw) error {
	_, err := storage.db.ExecContext(ctx, "INSERT INTO withdraws (order_number,user_id, sum,upload_data) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING", withdraw.OrderNumber, withdraw.UserID, withdraw.Sum, withdraw.UploadData.Format(time.RFC3339))
	if err != nil {
		return errors.New("error inserted new order to DB")
	}
	return nil
}

func (storage *SQLStorage) GetWithdraw(ctx context.Context, number string) (*withdraw.Withdraw, error) {
	withdraw := withdraw.Withdraw{OrderNumber: number}
	row := storage.db.QueryRowContext(ctx, "SELECT * FROM withdraws WHERE order_number=$1", number)
	var timeStr string
	err := row.Scan(&withdraw.OrderNumber, &withdraw.UserID, &withdraw.Sum, &timeStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	withdraw.UploadData, err = time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, errors.New("error parse data string from  DB")
	}
	return &withdraw, nil
}

func (storage *SQLStorage) GetAllWithdrawsForUser(ctx context.Context, id int) ([]withdraw.Withdraw, error) {

	rows, err := storage.db.QueryContext(ctx, "SELECT order_number,sum,upload_data FROM withdraws where user_id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("error get all orders for user %w", err)
	}
	defer rows.Close()

	result := make([]withdraw.Withdraw, 0)

	for rows.Next() {
		curWithdraw := withdraw.Withdraw{UserID: id}
		var upload string
		if err := rows.Scan(&curWithdraw.OrderNumber, &curWithdraw.Sum, &upload); err != nil {

			return nil, fmt.Errorf("error parse result from DB %w", err)
		}
		curWithdraw.UploadData, err = time.Parse(time.RFC3339, upload)
		if err != nil {
			return nil, errors.New("error parse data string from  DB")
		}
		result = append(result, curWithdraw)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error load data from DB %s", err.Error())
	}

	return result, nil
}

func (storage *SQLStorage) GetUserBalance(ctx context.Context, userID int) (*balance.Balance, error) {
	balance := balance.Balance{}
	row := storage.db.QueryRowContext(ctx, "SELECT accrual,withdraw FROM balances WHERE user_id=$1", userID)
	err := row.Scan(&balance.Amount, &balance.Withdraw)
	if err != nil {
		return nil, err
	}
	return &balance, nil
}
func (storage *SQLStorage) CreateUserBalance(ctx context.Context, userID int) error {
	_, err := storage.db.ExecContext(ctx, "INSERT INTO balances (user_id, accrual,withdraw) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING", userID, 0, 0)
	if err != nil {
		return errors.New("error inserted new order to DB")
	}
	return nil
}
func (storage *SQLStorage) ChangeBalanceAddAccrual(ctx context.Context, userID int, accrual float32) error {
	tx, err := storage.db.Begin()
	if err != nil {
		log.Print("error create transaction")
		return err
	}
	defer tx.Rollback()

	currentAccrual := float32(0)
	row := tx.QueryRowContext(ctx, "SELECT accrual FROM balances WHERE user_id = $1 FOR UPDATE", userID)
	err = row.Scan(&currentAccrual)
	if err != nil {
		log.Printf("error get current accrual for update for userid %d", userID)
		return err
	}
	_, err = tx.ExecContext(ctx, "UPDATE balances SET accrual = $1 WHERE user_id = $2", currentAccrual+accrual, userID)
	if err != nil {
		log.Printf("error update balance for userid %d, %s", userID, err.Error())
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Print("error commit update balance transaction")
		return err
	}

	return nil
}

func (storage *SQLStorage) ChangeBalanceAddWithdraw(ctx context.Context, userID int, sum float32, orderNumber string, uploadData time.Time) error {
	tx, err := storage.db.Begin()
	if err != nil {
		log.Print("error create transaction")
		return err
	}
	defer tx.Rollback()

	currentBalance := balance.Balance{}
	row := tx.QueryRowContext(ctx, "SELECT accrual,withdraw FROM balances WHERE user_id = $1 FOR UPDATE", userID)
	err = row.Scan(&currentBalance.Amount, &currentBalance.Withdraw)
	if err != nil {
		log.Print("error get current balance for withdraw")
		return err
	}
	if currentBalance.Amount < sum {
		return balance.ErrBalanceNoEnoughBals
	}

	_, err = tx.ExecContext(ctx, "UPDATE balances SET accrual = $1, withdraw = $2 WHERE user_id = $3", currentBalance.Amount-sum, currentBalance.Withdraw+sum, userID)
	if err != nil {
		log.Print("error update balance")
		return err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO withdraws (order_number,user_id, sum,upload_data) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING", orderNumber, userID, sum, uploadData.Format(time.RFC3339))

	if err != nil {
		return errors.New("error inserted new order to DB")
	}

	if err := tx.Commit(); err != nil {
		log.Print("error commit update balance transaction")
		return err
	}

	return nil
}

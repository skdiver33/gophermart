package balance

import (
	"context"
	"time"
)

type BalanceStorageInterface interface {
	GetUserBalance(ctx context.Context, userID int) (*Balance, error)
	CreateUserBalance(ctx context.Context, userID int) error
	ChangeBalanceAddAccrual(ctx context.Context, userID int, accrual float32) error
	ChangeBalanceAddWithdraw(ctx context.Context, userID int, sum float32, order string, uploadData time.Time) error
}

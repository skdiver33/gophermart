package balance

import "context"

type BalanceStorageInterface interface {
	GetUserBalance(ctx context.Context, userId int) (*Balance, error)
	CreateUserBalance(ctx context.Context, userId int) error
	ChangeBalanceAddAccrual(ctx context.Context, userId int, accrual int) error
	ChangeBalanceAddWithdraw(ctx context.Context, userId int, sum int) error
}

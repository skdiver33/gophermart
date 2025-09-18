package balance

import "context"

type BalanceStorageInterface interface {
	GetUserBalance(ctx context.Context, userId int) (*Balance, error)
	CreateUserBalance(ctx context.Context, userId int) error
	ChangeBalanceAddAccrual(ctx context.Context, userId int, accrual float32) error
	ChangeBalanceAddWithdraw(ctx context.Context, userId int, sum float32) error
}

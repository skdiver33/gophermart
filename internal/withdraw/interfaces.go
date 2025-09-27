package withdraw

import "context"

type WithdrawStorageInterface interface {
	AddWithdraw(ctx context.Context, withdraw *Withdraw) error
	GetWithdraw(ctx context.Context, orderNumber string) (*Withdraw, error)
	GetAllWithdrawsForUser(ctx context.Context, id int) ([]Withdraw, error)
}

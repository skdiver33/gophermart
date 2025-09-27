package order

import "context"

type OrderStorageInterface interface {
	AddOrder(ctx context.Context, order *Order) error
	GetOrder(ctx context.Context, number string) (*Order, error)
	GetAllOrderForID(ctx context.Context, id int) ([]Order, error)
	GetUnprocOrders(ctx context.Context) ([]Order, error)
	UpdateOrderStatus(ctx context.Context, orderNumber string, newStatus string, accrual float32) error
}

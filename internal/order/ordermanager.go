package order

import (
	"context"
	"errors"
	"fmt"
	"sort"
)

type OrderManager struct {
	OrderStorage OrderStorageInterface
	// AccrualClient loyalty.AccrualClientInterface
}

func NewOrderManager(storage OrderStorageInterface) *OrderManager {
	return &OrderManager{OrderStorage: storage}
}

var (
	ErrOrderLoadAnotherUser = errors.New("order load another user")
	ErrOrderAlreadyLoad     = errors.New("order already load")
)

func (om *OrderManager) LoadOrder(ctx context.Context, newOrder *Order) error {
	order, err := om.OrderStorage.GetOrder(ctx, newOrder.Number)
	if err != nil {
		return fmt.Errorf("internal error %w", err)
	}
	if order == nil {
		if err := om.OrderStorage.AddOrder(ctx, newOrder); err != nil {
			return fmt.Errorf("error add new order. %w", err)
		}
		return nil
	}
	if order.UserID != newOrder.UserID {
		return ErrOrderLoadAnotherUser
	}
	newOrder.Status = order.Status
	return ErrOrderAlreadyLoad
}

func (om *OrderManager) GetAllOrdersForUser(ctx context.Context, id int) (*[]Order, error) {
	allOrders, err := om.OrderStorage.GetAllOrderForID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error get all order %w", err)
	}
	sort.Slice(allOrders, func(i, j int) bool {
		return allOrders[i].UploadData.After(allOrders[j].UploadData)
	})
	return &allOrders, nil
}

func (om *OrderManager) GetAllUnprocOrders(ctx context.Context) (map[string]Order, error) {
	allOrders, err := om.OrderStorage.GetUnprocOrders(ctx)
	resultMap := make(map[string]Order, 0)
	if err != nil {
		return nil, fmt.Errorf("error get all unprocess order %w", err)
	}
	for _, order := range allOrders {
		err := om.UpdateOrderStatus(ctx, order.Number, OrderStautsProcessing, 0)
		if err != nil {
			return nil, fmt.Errorf("error update status for unproc order. %w", err)
		}
		resultMap[order.Number] = order
	}
	return resultMap, nil
}

func (om *OrderManager) UpdateOrderStatus(ctx context.Context, orderNumber string, newStatus string, accrual float32) error {
	err := om.OrderStorage.UpdateOrderStatus(ctx, orderNumber, newStatus, accrual)
	if err != nil {
		return fmt.Errorf("error update status for order. %w", err)
	}
	return nil
}

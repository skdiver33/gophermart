package withdraw

import (
	"context"
	"errors"
	"fmt"
	"sort"
)

type WithdrawManager struct {
	WithdrawStorage WithdrawStorageInterface
}

func NewWithdrawManager(storage WithdrawStorageInterface) *WithdrawManager {
	return &WithdrawManager{WithdrawStorage: storage}
}

func (wm *WithdrawManager) GetWithdraw(ctx context.Context, orderNumber string) (*Withdraw, error) {
	requestWithdraw, err := wm.WithdrawStorage.GetWithdraw(ctx, orderNumber)
	if err != nil {
		return nil, fmt.Errorf("internal error %w", err)
	}
	return requestWithdraw, nil
}

func (wm *WithdrawManager) AddWithdraw(ctx context.Context, withdraw *Withdraw) error {
	wd, err := wm.GetWithdraw(ctx, withdraw.OrderNumber)
	if err != nil {
		return fmt.Errorf("error get withdraft for order. %w", err)
	}
	if wd != nil {
		return errors.New("error. withdrawt for oder already exist")
	}
	err = wm.WithdrawStorage.AddWithdraw(ctx, withdraw)
	return err
}

func (wm *WithdrawManager) GetAllWithdrawsForUser(ctx context.Context, id int) (*[]Withdraw, error) {
	allWithdraw, err := wm.WithdrawStorage.GetAllWithdrawsForUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error get all withdraw for user %w", err)
	}
	sort.Slice(allWithdraw, func(i, j int) bool {
		return allWithdraw[i].UploadData.After(allWithdraw[j].UploadData)
	})
	return &allWithdraw, nil
}

package balance

import (
	"context"
	"errors"
)

type Balance struct {
	Amount   float32 `json:"current"`
	Withdraw float32 `json:"withdrawn"`
}

type BalanceManager struct {
	BalanceStorage BalanceStorageInterface
}

func NewBalanceManager(storage BalanceStorageInterface) *BalanceManager {
	return &BalanceManager{BalanceStorage: storage}
}

var (
	ErrBalanceNoEnoughBals = errors.New("no enought bals")
)

func (bm *BalanceManager) GetUserBalance(ctx context.Context, userId int) (*Balance, error) {
	balance, err := bm.BalanceStorage.GetUserBalance(ctx, userId)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (bm *BalanceManager) CreateUserBalance(ctx context.Context, userId int) error {
	err := bm.BalanceStorage.CreateUserBalance(ctx, userId)
	return err
}

func (bm *BalanceManager) WithdrawUserAccural(ctx context.Context, userId int, sum float32) error {
	err := bm.BalanceStorage.ChangeBalanceAddWithdraw(ctx, userId, sum)
	return err
}

func (bm *BalanceManager) AddAmount(ctx context.Context, userId int, accrual float32) error {
	err := bm.BalanceStorage.ChangeBalanceAddAccrual(ctx, userId, accrual)
	return err
}

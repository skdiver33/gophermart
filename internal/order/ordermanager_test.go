package order_test

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mocks "github.com/skdiver33/gophermart/internal/mocks"
	"github.com/skdiver33/gophermart/internal/order"
)

func TestOrderManager_LoadOrder(t *testing.T) {
	type testData struct {
		name      string
		testOrder order.Order
		err       error
		wantErr   bool
		getErr    error
	}
	orders := []testData{
		{
			name: "positive test #1",
			testOrder: order.Order{
				Number:     "123456",
				UserID:     1,
				Status:     order.OrderStatusNew,
				UploadData: time.Now(),
			},
			err:     nil,
			getErr:  nil,
			wantErr: false,
		},
		{
			name: "positive test #2",
			testOrder: order.Order{
				Number:     "7890123",
				UserID:     1,
				Status:     order.OrderStatusNew,
				UploadData: time.Now(),
			},
			err:     nil,
			getErr:  nil,
			wantErr: false,
		},
		{
			name: "negative test #1",
			testOrder: order.Order{
				Number:     "7890123",
				UserID:     2,
				Status:     order.OrderStatusNew,
				UploadData: time.Now(),
			},
			err:     order.ErrOrderLoadAnotherUser,
			getErr:  nil,
			wantErr: true,
		},
		{
			name: "negative test #2",
			testOrder: order.Order{
				Number:     "123456",
				UserID:     1,
				Status:     order.OrderStatusNew,
				UploadData: time.Now(),
			},
			err:     order.ErrOrderAlreadyLoad,
			getErr:  nil,
			wantErr: true,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockOrderStorageInterface(ctrl)

	for _, item := range orders {
		addOrder := &item.testOrder
		m.EXPECT().AddOrder(t.Context(), addOrder).Return(item.err).AnyTimes()
		var ret *order.Order
		if item.wantErr {
			ret = addOrder
		}
		m.EXPECT().GetOrder(t.Context(), addOrder.Number).Return(ret, item.getErr)
	}

	om := order.NewOrderManager(m)
	for _, tt := range orders {
		t.Run(tt.name, func(t *testing.T) {

			gotErr := om.LoadOrder(t.Context(), &tt.testOrder)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("LoadOrder() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("LoadOrder() succeeded unexpectedly")
			}
		})
	}
}

func TestOrderManager_UpdateOrderStatus(t *testing.T) {
	type testData struct {
		name      string
		testOrder order.Order
		err       error
		wantErr   bool
	}
	orders := []testData{
		{
			name: "update order positive test #1",
			testOrder: order.Order{
				Number:  "123456",
				Status:  order.OrderStatusProcessed,
				Accrual: 10.5,
			},
			err: nil,
		},
		{
			name: "update order negative test #1",
			testOrder: order.Order{
				Number:  "7890123",
				Status:  order.OrderStatusInvalid,
				Accrual: 0,
			},
			err:     errors.New("not found"),
			wantErr: true,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockOrderStorageInterface(ctrl)

	for _, item := range orders {
		addOrder := &item.testOrder
		m.EXPECT().UpdateOrderStatus(t.Context(), addOrder.Number, addOrder.Status, addOrder.Accrual).Return(item.err)
	}

	om := order.NewOrderManager(m)

	for _, tt := range orders {
		t.Run(tt.name, func(t *testing.T) {

			gotErr := om.UpdateOrderStatus(t.Context(), tt.testOrder.Number, tt.testOrder.Status, tt.testOrder.Accrual)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateOrderStatus() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateOrderStatus() succeeded unexpectedly")
			}
		})
	}
}

package withdraw_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/skdiver33/gophermart/internal/mocks"
	"github.com/skdiver33/gophermart/internal/withdraw"
)

func TestWithdrawManager_AddWithdraw(t *testing.T) {

	type testData struct {
		name         string
		testWithdraw *withdraw.Withdraw
		addErr       error
		getErr       error
		wantErr      bool
	}
	tests := []testData{
		{
			name: "positive test #1",
			testWithdraw: &withdraw.Withdraw{
				OrderNumber: "123456",
				UserID:      1,
				Sum:         100,
			},
			getErr:  nil,
			addErr:  nil,
			wantErr: false,
		},
		{
			name: "negative test #1",
			testWithdraw: &withdraw.Withdraw{
				OrderNumber: "123456",
				UserID:      1,
				Sum:         100,
			},
			getErr:  nil,
			addErr:  errors.New("already exist"),
			wantErr: true,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockWithdrawStorageInterface(ctrl)

	for _, data := range tests {
		m.EXPECT().AddWithdraw(t.Context(), data.testWithdraw).Return(data.addErr).AnyTimes()
		var ret *withdraw.Withdraw
		if data.wantErr {
			ret = data.testWithdraw
		}
		m.EXPECT().GetWithdraw(t.Context(), data.testWithdraw.OrderNumber).Return(ret, data.getErr)
	}

	wm := withdraw.NewWithdrawManager(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotErr := wm.AddWithdraw(t.Context(), tt.testWithdraw)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("AddWithdraw() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("AddWithdraw() succeeded unexpectedly")
			}
		})
	}
}

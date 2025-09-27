package balance_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/skdiver33/gophermart/internal/balance"
	"github.com/skdiver33/gophermart/internal/mocks"
	"github.com/stretchr/testify/require"
)

func TestBalanceManager_GetUserBalance(t *testing.T) {

	type testData struct {
		name        string
		testBalance *balance.Balance
		userID      int
		err         error
		wantErr     bool
	}
	tests := []testData{
		{
			name: "positive test #1",
			testBalance: &balance.Balance{
				Amount:   10,
				Withdraw: 10,
			},
			userID:  1,
			err:     nil,
			wantErr: false,
		},
		{
			name:        "negative test #1",
			testBalance: nil,
			userID:      5,
			err:         errors.New("wrong ID"),
			wantErr:     true,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockBalanceStorageInterface(ctrl)

	for _, item := range tests {
		m.EXPECT().GetUserBalance(t.Context(), item.userID).Return(item.testBalance, item.err)
	}

	bm := balance.NewBalanceManager(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, gotErr := bm.GetUserBalance(t.Context(), tt.userID)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}
			require.Nil(t, gotErr)
			require.NotNil(t, got)
		})
	}
}

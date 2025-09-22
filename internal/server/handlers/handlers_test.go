package handlers_test

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/jwtauth/v5"
	"github.com/skdiver33/gophermart/internal/auth"
	bm "github.com/skdiver33/gophermart/internal/balance"
	om "github.com/skdiver33/gophermart/internal/order"
	"github.com/skdiver33/gophermart/internal/server/handlers"
	um "github.com/skdiver33/gophermart/internal/user"
	wm "github.com/skdiver33/gophermart/internal/withdraw"
	"github.com/skdiver33/gophermart/storage"
	"github.com/stretchr/testify/assert"
)

var (
	authent         *auth.Auth
	sqlStorage      *storage.SQLStorage
	userManager     *um.UserManager
	orderManager    *om.OrderManager
	withdrawManager *wm.WithdrawManager
	balanceManager  *bm.BalanceManager
	handler         *handlers.ServerHandler
	token           []string
)

func init() {
	authent = auth.NewAuth()
	var err error
	sqlStorage, err = storage.NewSQLStorage("postgres://gophermart:secret@192.168.1.48:5432/gophermart?sslmode=disable")
	if err != nil {
		log.Printf("error create sql storage %s", err.Error())
	}
	userManager = um.NewUserManager(sqlStorage, authent)
	orderManager = om.NewOrderManager(sqlStorage)
	withdrawManager = wm.NewWithdrawManager(sqlStorage)
	balanceManager = bm.NewBalanceManager(sqlStorage)
	handler = handlers.NewServerHandler(userManager, orderManager, withdrawManager, balanceManager)
	token = make([]string, 0)
}

func TestServerHandler_UserRegisterHandler(t *testing.T) {

	type want struct {
		code        int
		wantCookies bool
	}
	tests := []struct {
		name        string
		requestData []byte
		want        want
	}{
		{
			name:        "positive register user",
			requestData: []byte(`{"login":"user","password":"user"}`),
			want: want{
				code:        200,
				wantCookies: true,
			},
		},
		{
			name:        "positive register user",
			requestData: []byte(`{"login":"admin","password":"admin"}`),
			want: want{
				code:        200,
				wantCookies: true,
			},
		},
		{
			name:        "user already exist",
			requestData: []byte(`{"login":"user","password":"user"}`),
			want: want{
				code:        409,
				wantCookies: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(tt.requestData))
			request.Header.Add("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.UserRegisterHandler(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			defer res.Body.Close()
			cookies := res.Cookies()
			isCookieExist := len(cookies) > 0
			assert.Equal(t, tt.want.wantCookies, isCookieExist)

		})

	}
	//sqlStorage.CloseAndClean().Error()
}

func TestServerHandler_UserLoginHandler(t *testing.T) {
	type want struct {
		code        int
		wantCookies bool
	}
	tests := []struct {
		name        string
		requestData []byte
		want        want
	}{
		{
			name:        "positive auth user",
			requestData: []byte(`{"login":"user","password":"user"}`),
			want: want{
				code:        200,
				wantCookies: true,
			},
		},
		{
			name:        "positive register user",
			requestData: []byte(`{"login":"admin","password":"admin"}`),
			want: want{
				code:        200,
				wantCookies: true,
			},
		},
		{
			name:        "wrong password",
			requestData: []byte(`{"login":"user","password":"123"}`),
			want: want{
				code:        401,
				wantCookies: false,
			},
		},
		{
			name:        "wrong request type",
			requestData: []byte(`{"name":"user","passd":"123}`),
			want: want{
				code:        400,
				wantCookies: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(tt.requestData))
			request.Header.Add("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.UserLoginHandler(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			defer res.Body.Close()
			cookies := res.Cookies()
			isCookieExist := len(cookies) > 0
			assert.Equal(t, tt.want.wantCookies, isCookieExist)
			if len(cookies) > 0 {
				for _, item := range cookies {
					if item.Name == "jwt" {
						token = append(token, item.Value)
					}
				}
			}
		})
	}

}

func TestServerHandler_LoadOrderHandler(t *testing.T) {
	tests := []struct {
		name        string
		requestData string
		userToken   string
		wantCode    int
	}{
		{
			name:        "positive order load",
			requestData: "5633425",
			userToken:   token[0],
			wantCode:    202,
		},
		{
			name:        "order already load ",
			requestData: "5633425",
			userToken:   token[0],
			wantCode:    200,
		},
		{
			name:        "order load another user",
			requestData: "5633425",
			userToken:   token[1],
			wantCode:    409,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tok, err := jwtauth.VerifyToken(authent.GetBaseToken(), tt.userToken)
			if err != nil {
				log.Print("error create tok")
			}
			ctx2 := context.WithValue(ctx, jwtauth.TokenCtxKey, tok)
			request := httptest.NewRequestWithContext(ctx2, http.MethodPost, "/api/user/login", bytes.NewReader([]byte(tt.requestData)))
			request.Header.Add("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			handler.LoadOrderHandler(w, request)
			res := w.Result()
			assert.Equal(t, tt.wantCode, res.StatusCode)
		})
	}
	sqlStorage.CloseAndClean().Error()
}

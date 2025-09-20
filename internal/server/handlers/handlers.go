package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	bm "github.com/skdiver33/gophermart/internal/balance"
	om "github.com/skdiver33/gophermart/internal/order"
	um "github.com/skdiver33/gophermart/internal/user"
	wm "github.com/skdiver33/gophermart/internal/withdraw"
)

type ServerHandler struct {
	userManager     *um.UserManager
	orderManager    *om.OrderManager
	withdrawManager *wm.WithdrawManager
	balanceManager  *bm.BalanceManager
}

func NewServerHandler(userManager *um.UserManager, orderManager *om.OrderManager, withdrawManager *wm.WithdrawManager, balanceManager *bm.BalanceManager) *ServerHandler {
	return &ServerHandler{userManager: userManager, orderManager: orderManager, withdrawManager: withdrawManager, balanceManager: balanceManager}
}

func (handler *ServerHandler) UserRegisterHandler(rw http.ResponseWriter, request *http.Request) {
	userData := um.User{}
	if err := json.NewDecoder(request.Body).Decode(&userData); err != nil {
		log.Printf("error user register. %s", err.Error())
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	userData.CryptPasswd()
	user, err := handler.userManager.UserRegister(request.Context(), &userData)
	if err != nil {
		returnStatus := http.StatusInternalServerError
		if errors.Is(err, um.ErrUserAlreadyExist) {
			returnStatus = http.StatusConflict
		}
		log.Printf("error user register. %s", err.Error())
		http.Error(rw, err.Error(), returnStatus)
		return
	}
	userAuthToken, err := handler.userManager.UserAuth(request.Context(), user)
	if err != nil {
		returnStatus := http.StatusInternalServerError
		if errors.Is(err, um.ErrUserWithCredNotFound) {
			returnStatus = http.StatusUnauthorized
		}
		log.Printf("error user register. auth error. %s", err.Error())
		http.Error(rw, err.Error(), returnStatus)
		return
	}
	err = handler.balanceManager.CreateUserBalance(request.Context(), user.ID)
	if err != nil {
		log.Printf("error user register. create user balance error. %s", err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{}
	cookie.Name = "jwt"
	cookie.Value = userAuthToken
	http.SetCookie(rw, &cookie)
	rw.Header().Set("Content-Type", "application/text-plain")
	rw.Write([]byte(userAuthToken))
}

func (handler *ServerHandler) UserLoginHandler(rw http.ResponseWriter, request *http.Request) {
	userData := um.User{}
	if err := json.NewDecoder(request.Body).Decode(&userData); err != nil {
		log.Printf("error user login.  %s", err.Error())
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	userData.CryptPasswd()
	userAuthToken, err := handler.userManager.UserAuth(request.Context(), &userData)
	if err != nil {
		returnStatus := http.StatusInternalServerError
		if errors.Is(err, um.ErrUserWithCredNotFound) {
			returnStatus = http.StatusUnauthorized
		}
		log.Printf("error user login. auth error. %s", err.Error())
		http.Error(rw, err.Error(), returnStatus)
		return
	}
	cookie := http.Cookie{}
	cookie.Name = "jwt"
	cookie.Value = userAuthToken
	http.SetCookie(rw, &cookie)
	rw.Header().Set("Content-Type", "application/text-plain")
	rw.Write([]byte(userAuthToken))
}

func (handler *ServerHandler) LoadOrderHandler(rw http.ResponseWriter, request *http.Request) {
	newOrder := om.Order{Status: om.OrderStatusNew, Accrual: 0, UploadData: time.Now()}
	if request.Header.Get("Content-Type") != "text/plain" {
		log.Printf("error upload order.  Request type wrong.")
		http.Error(rw, "bad requst format", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		log.Printf("error read upload body %s", err.Error())
		return
	}
	newOrder.Number = string(body)
	if !newOrder.LunaCheck() {
		log.Printf("error check order number %v", newOrder.Number)
		http.Error(rw, "wrong order number format", http.StatusUnprocessableEntity)
		return
	}
	newOrder.UserID, err = handler.userManager.Authenticator.GetUserIDFromClaims(request.Context())
	if err != nil {
		log.Printf("error get user ID from JWT %s", err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	err = handler.orderManager.LoadOrder(request.Context(), &newOrder)
	if err != nil {
		returnCode := http.StatusInternalServerError
		if errors.Is(err, om.ErrOrderLoadAnotherUser) {
			returnCode = http.StatusConflict
		}
		if errors.Is(err, om.ErrOrderAlreadyLoad) {
			returnCode = http.StatusOK
		}
		log.Printf("error upload order %s", err.Error())
		http.Error(rw, err.Error(), returnCode)
		return
	}
	rw.WriteHeader(http.StatusAccepted)
}

func (handler *ServerHandler) GetAllOrdersHandler(rw http.ResponseWriter, request *http.Request) {
	userID, err := handler.userManager.Authenticator.GetUserIDFromClaims(request.Context())
	if err != nil {
		log.Printf("error get user ID from JWT %s", err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	orders, err := handler.orderManager.GetAllOrdersForUser(request.Context(), userID)
	if err != nil {
		log.Printf("error get all orders for user %s", err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(*orders) == 0 {
		http.Error(rw, "no orders for user", http.StatusNoContent)
		return
	}
	resp, err := json.Marshal(orders)
	if err != nil {
		log.Printf("error marshall orders %s", err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(resp)
}

func (handler *ServerHandler) GetBalanceHandler(rw http.ResponseWriter, request *http.Request) {
	userID, err := handler.userManager.Authenticator.GetUserIDFromClaims(request.Context())
	if err != nil {
		log.Printf("error get user ID from JWT %s", err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	balance, err := handler.balanceManager.GetUserBalance(request.Context(), userID)
	if err != nil {
		log.Printf("error get user balance %s", err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(balance)
	if err != nil {
		log.Printf("error marshall balance %s", err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(resp)

}

func (handler *ServerHandler) GetWithdrawHandler(rw http.ResponseWriter, request *http.Request) {
	newWithdraw := wm.Withdraw{}
	if err := json.NewDecoder(request.Body).Decode(&newWithdraw); err != nil {
		log.Printf("error decode withdraw %s", err.Error())
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	newWithdraw.UploadData = time.Now()
	id, err := handler.userManager.Authenticator.GetUserIDFromClaims(request.Context())
	if err != nil {
		log.Printf("error get user ID from JWT %s", err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	newWithdraw.UserID = id
	if !newWithdraw.CheckNumber() {
		log.Printf("error withdraw. wrong order number format")
		http.Error(rw, "wrong order number format", http.StatusUnprocessableEntity)
		return
	}
	if wd, _ := handler.withdrawManager.GetWithdraw(request.Context(), newWithdraw.OrderNumber); wd != nil {
		log.Printf("error withdraw. withdraw for order allready exist")
		http.Error(rw, "withdraw for order already exist", http.StatusUnprocessableEntity)
		return
	}
	err = handler.balanceManager.WithdrawUserAccural(request.Context(), id, newWithdraw.Sum)
	if err != nil {
		retCode := http.StatusInternalServerError
		if errors.Is(err, bm.ErrBalanceNoEnoughBals) {
			retCode = http.StatusPaymentRequired
		}
		log.Printf("error withdraw. %s", err)
		http.Error(rw, err.Error(), retCode)
		return
	}
	err = handler.withdrawManager.AddWithdraw(request.Context(), &newWithdraw)
	if err != nil {
		log.Printf("error change withdraw %s", err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (handler *ServerHandler) GetWithdrawAllHandler(rw http.ResponseWriter, request *http.Request) {
	userID, err := handler.userManager.Authenticator.GetUserIDFromClaims(request.Context())
	if err != nil {
		log.Printf("error get user ID from JWT %s", err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	wdraws, err := handler.withdrawManager.GetAllWithdrawsForUser(request.Context(), userID)
	if err != nil {
		log.Printf("error get withdraw report %s", err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(*wdraws) == 0 {
		http.Error(rw, "no withdraws for user", http.StatusNoContent)
		return
	}
	resp, err := json.Marshal(wdraws)
	if err != nil {
		log.Printf("error marshal withdraws %s", err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(resp)
}

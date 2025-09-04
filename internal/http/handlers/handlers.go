package handlers

import (
	"encoding/json"
	"net/http"

	um "github.com/skdiver33/gophermart/internal/user"
)

type ServerHandler struct {
	userManager um.UserManager
}

func NewServerHandler(manager *um.UserManager) *ServerHandler {
	return &ServerHandler{userManager: *manager}
}

func (handler *ServerHandler) UserRegisterHandler(rw http.ResponseWriter, request *http.Request) {}
func (handler *ServerHandler) UserLoginHandler(rw http.ResponseWriter, request *http.Request) {
	userData := um.User{}
	if err := json.NewDecoder(request.Body).Decode(&userData); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

}
func (handler *ServerHandler) UploadOrderHandler(rw http.ResponseWriter, request *http.Request)    {}
func (handler *ServerHandler) DownloadOrdersHandler(rw http.ResponseWriter, request *http.Request) {}
func (handler *ServerHandler) GetBalanceHandler(rw http.ResponseWriter, request *http.Request)     {}
func (handler *ServerHandler) GetWithdrawHandler(rw http.ResponseWriter, request *http.Request)    {}
func (handler *ServerHandler) GetWithdrawAllHandler(rw http.ResponseWriter, request *http.Request) {}

func (handler *ServerHandler) DefaultHandler(rw http.ResponseWriter, request *http.Request) {}

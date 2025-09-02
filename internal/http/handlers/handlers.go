package handlers

import "net/http"

type ServerHandler struct{}

func NewServerHandler() *ServerHandler {
	return &ServerHandler{}
}

func (handler *ServerHandler) UserRegisterHandler(rw http.ResponseWriter, request *http.Request)   {}
func (handler *ServerHandler) UserLoginHandler(rw http.ResponseWriter, request *http.Request)      {}
func (handler *ServerHandler) UploadOrderHandler(rw http.ResponseWriter, request *http.Request)    {}
func (handler *ServerHandler) DownloadOrdersHandler(rw http.ResponseWriter, request *http.Request) {}
func (handler *ServerHandler) GetBalanceHandler(rw http.ResponseWriter, request *http.Request)     {}
func (handler *ServerHandler) GetWithdrawHandler(rw http.ResponseWriter, request *http.Request)    {}
func (handler *ServerHandler) GetWithdrawAllHandler(rw http.ResponseWriter, request *http.Request) {}

func (handler *ServerHandler) DefaultHandler(rw http.ResponseWriter, request *http.Request) {}

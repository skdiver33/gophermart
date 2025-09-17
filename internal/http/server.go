package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth/v5"
	auth "github.com/skdiver33/gophermart/internal/auth"
	bm "github.com/skdiver33/gophermart/internal/balance"
	handler "github.com/skdiver33/gophermart/internal/http/handlers"
	mid "github.com/skdiver33/gophermart/internal/http/middleware"
	loyalty "github.com/skdiver33/gophermart/internal/loyalty"
	om "github.com/skdiver33/gophermart/internal/order"
	um "github.com/skdiver33/gophermart/internal/user"
	wm "github.com/skdiver33/gophermart/internal/withdraw"
	storage "github.com/skdiver33/gophermart/storage"
)

type Server struct {
	Config         *ServerConfig
	HandlersRouter http.Handler
}

type ServerConfig struct {
	ListenAddress string
}

func newServerConfig() *ServerConfig {
	return &ServerConfig{ListenAddress: "localhost:8080"}
}

func NewServer() (*Server, error) {
	server := Server{}
	server.Config = newServerConfig()

	auth := auth.NewAuth()
	storage, err := storage.NewSQLStorage("postgres://gophermart:secret@192.168.1.48:5432/gophermart?sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("error creatr sql storage. %w", err)
	}

	userManager := um.NewUserManager(storage, auth)
	orderManager := om.NewOrderManager(storage)
	withdrawManager := wm.NewWithdrawManager(storage)
	balanceManager := bm.NewBalanceManager(storage)

	orderProcessor := loyalty.NewOrderProcessor(balanceManager, orderManager, loyalty.NewAccuralClientConfig())
	go orderProcessor.Start(context.TODO())

	serverHandler := handler.NewServerHandler(userManager, orderManager, withdrawManager, balanceManager)
	serverLoger := mid.NewServiceLogger()

	newRouter := chi.NewRouter()
	newRouter.Use(serverLoger.RequestLogger)

	newRouter.Group(func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("welcome anonymous"))
		})
		r.Route("/api/user", func(r chi.Router) {
			r.Post("/register", serverHandler.UserRegisterHandler)
			r.Post("/login", serverHandler.UserLoginHandler)
			r.Group(func(r chi.Router) {
				r.Use(jwtauth.Verifier(auth.GetBaseToken()))
				r.Use(jwtauth.Authenticator(auth.GetBaseToken()))
				r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
					_, claims, _ := jwtauth.FromContext(r.Context())
					w.Write([]byte(fmt.Sprintf("protected area. hi %v", claims["user_id"])))
				})
				r.Post("/orders", serverHandler.LoadOrderHandler)
				r.Get("/orders", serverHandler.GetAllOrdersHandler)
				r.Get("/balance", serverHandler.GetBalanceHandler)
				r.Post("/balance/withdraw", serverHandler.GetWithdrawHandler)
				r.Get("/withdrawals", serverHandler.GetWithdrawAllHandler)
			})
		})

	})

	server.HandlersRouter = newRouter

	return &server, nil
}

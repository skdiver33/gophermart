package http

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

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
	ListenAddress  string
	SQLDBAddress   string
	AccrualAddress string
}

func newServerConfig() *ServerConfig {

	serverConfig := ServerConfig{}
	serverFlags := flag.NewFlagSet("Server config flags", flag.ContinueOnError)
	serverFlags.StringVar(&serverConfig.ListenAddress, "a", "localhost:8000", "address for start server in form ip:port. default localhost:8080")
	serverFlags.StringVar(&serverConfig.SQLDBAddress, "d", "postgres://gophermart:secret@192.168.1.48:5432/gophermart?sslmode=disable", "DB connection string. Default - empty and disable.")
	serverFlags.StringVar(&serverConfig.AccrualAddress, "r", "localhost:8080", "accrual system address")
	serverFlags.Parse(os.Args[1:])

	envServerAddr, ok := os.LookupEnv("RUN_ADDRESS")
	if ok {
		serverConfig.ListenAddress = envServerAddr
	}

	envDBAddr, ok := os.LookupEnv("DATABASE_URI")
	if ok {
		serverConfig.SQLDBAddress = envDBAddr
	}

	envAccrualAddr, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if ok {
		serverConfig.AccrualAddress = envAccrualAddr
	}

	return &serverConfig
}

func NewServer() (*Server, error) {
	server := Server{}
	server.Config = newServerConfig()

	auth := auth.NewAuth()
	storage, err := storage.NewSQLStorage(server.Config.SQLDBAddress)
	if err != nil {
		log.Printf("error create new server. error create sql storage %s", err.Error())
		return nil, fmt.Errorf("error create sql storage. %w", err)
	}

	userManager := um.NewUserManager(storage, auth)
	orderManager := om.NewOrderManager(storage)
	withdrawManager := wm.NewWithdrawManager(storage)
	balanceManager := bm.NewBalanceManager(storage)

	log.Printf("pre start accrual client")
	orderProcessor := loyalty.NewOrderProcessor(balanceManager, orderManager, loyalty.NewAccuralClientConfig(server.Config.AccrualAddress))
	go orderProcessor.Start(context.TODO())

	serverHandler := handler.NewServerHandler(userManager, orderManager, withdrawManager, balanceManager)
	serverLoger := mid.NewServiceLogger()

	newRouter := chi.NewRouter()
	newRouter.Use(serverLoger.RequestLogger)

	newRouter.Group(func(r chi.Router) {
		r.Route("/api/user", func(r chi.Router) {
			r.Post("/register", serverHandler.UserRegisterHandler)
			r.Post("/login", serverHandler.UserLoginHandler)
			r.Group(func(r chi.Router) {
				r.Use(jwtauth.Verifier(auth.GetBaseToken()))
				r.Use(jwtauth.Authenticator(auth.GetBaseToken()))
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

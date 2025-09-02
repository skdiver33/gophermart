package http

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth/v5"
	handler "github.com/skdiver33/gophermart/internal/http/handlers"
)

func init() {
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	// For debugging/example purposes, we generate and print
	// a sample jwt token with claims `user_id:123` here:
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"user_id": 123})
	fmt.Printf("DEBUG: a sample jwt is %s\n\n", tokenString)
}

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

	serverHandler := handler.NewServerHandler()
	newRouter := chi.NewRouter()
	//	newRouter.Use(serverHandler.RequestLogger)
	//	newRouter.Use(serverHandler.SigningHandle)
	//	newRouter.Use(newHserverHandlerandler.GzipHandle)

	// Public routes
	newRouter.Group(func(router chi.Router) {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("welcome anonymous"))
		})
		router.Post("/register", serverHandler.UserRegisterHandler)
		router.Post("/login", serverHandler.UserLoginHandler)
	})

	server.HandlersRouter = newRouter

	return &server, nil

	// newRouter.Route("/api/user", func(r chi.Router) {

	// 	r.Route("/value", func(r chi.Router) {
	// 		r.Post("/", newHandler.GetJSONMetrics)
	// 		r.Get("/{metricsType}/{metricsName}", newHandler.GetMetrics)
	// 	})
	// 	r.Route("/update", func(r chi.Router) {
	// 		r.Post("/", newHandler.SetJSONMetrics)
	// 		r.Post("/{metricsType}/{metricsName}/{metricsValue}", newHandler.SetMetrics)
	// 	})
	// })
}

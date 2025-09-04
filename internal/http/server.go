package http

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth/v5"
	auth "github.com/skdiver33/gophermart/internal/auth"
	handler "github.com/skdiver33/gophermart/internal/http/handlers"
	um "github.com/skdiver33/gophermart/internal/user"
	memstorage "github.com/skdiver33/gophermart/storage"
)

// func init() {
// 	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

// 	// For debugging/example purposes, we generate and print
// 	// a sample jwt token with claims `user_id:123` here:
// 	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"user_id": 123, "exp": time.Now().Add(20 * time.Second)})
// 	fmt.Printf("DEBUG: a sample jwt is %s\n\n", tokenString)
// }

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
	storage := memstorage.NewUserMemStorage()
	manager := um.NewUserManager(storage, auth)

	serverHandler := handler.NewServerHandler(manager)

	newRouter := chi.NewRouter()
	//	newRouter.Use(serverHandler.RequestLogger)
	//	newRouter.Use(serverHandler.SigningHandle)
	//	newRouter.Use(newHserverHandlerandler.GzipHandle)

	// Public routes
	newRouter.Group(func(router chi.Router) {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("welcome anonymous"))
		})
		router.Route("/api/user", func(r chi.Router) {
			r.Post("/register", serverHandler.UserRegisterHandler)
			r.Post("/login", serverHandler.UserLoginHandler)
		})

	})

	// Protected route
	newRouter.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(auth.GetBaseToken()))
		r.Use(jwtauth.Authenticator(auth.GetBaseToken()))
		r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			_, claims, _ := jwtauth.FromContext(r.Context())
			w.Write([]byte(fmt.Sprintf("protected area. hi %v", claims["user_id"])))
		})
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

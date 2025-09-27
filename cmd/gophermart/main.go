package main

import (
	"net/http"

	server "github.com/skdiver33/gophermart/internal/server"
)

func main() {
	server, err := server.NewServer()
	if err != nil {
		panic(err.Error())
	}
	if err := http.ListenAndServe(server.Config.ListenAddress, server.HandlersRouter); err != nil {
		panic(err.Error())
	}
}

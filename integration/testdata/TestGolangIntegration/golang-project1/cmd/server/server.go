package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/example/golang-app/pkg/http/handlers"
	"github.com/example/golang-app/pkg/http/middlewares"
	"github.com/gorilla/mux"
)

const (
	defaultServerAddress = "0.0.0.0"
	defaultServerPort    = "8080"
)

func main() {
	serverAddress := flag.String("server.address", defaultServerAddress, "IP address of the HTTP server")
	serverPort := flag.String("server.port", defaultServerPort, "Port of the HTTP server")
	flag.Parse()

	router := mux.NewRouter()

	router.HandleFunc("/hello/{who}", middlewares.MultipleMiddleware(handlers.HelloHandler,
		middlewares.LoggerMiddleware, middlewares.RecoverMiddleware)).Methods("GET", "POST")
	router.HandleFunc("/panic", middlewares.MultipleMiddleware(handlers.ThrowPanicHandler,
		middlewares.LoggerMiddleware, middlewares.RecoverMiddleware)).Methods("GET", "POST")
	http.Handle("/", router)

	fmt.Printf("Server address: %s:%s\n", *serverAddress, *serverPort)
	fmt.Println("Server is listening...")
	http.ListenAndServe(fmt.Sprintf("%s:%s", *serverAddress, *serverPort), nil)
}

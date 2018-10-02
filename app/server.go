package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/spf13/viper"
)

// Server provides an http.Server.
type Server struct {
	*http.Server
}

// NewServer creates and configures an APIServer serving all application routes.
func NewServer() (*Server, error) {
	log.Println("Configuring server...")
	app := InitRouter()

	port := viper.GetString("port")

	srv := http.Server{
		Addr:    ":" + port,
		Handler: app,
	}

	return &Server{&srv}, nil
}

// Start runs ListenAndServe on the http.Server with graceful shutdown.
func (srv *Server) Start() {
	log.Println("Starting server...")
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()
	log.Printf("Listening on %s\n", srv.Addr)

	// dealing with termination
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit
	log.Println("Shutting down server... Reason:", sig)
	if err := srv.Shutdown(context.Background()); err != nil {
		panic(err)
	}
	log.Println("Server gracefully stopped")
}

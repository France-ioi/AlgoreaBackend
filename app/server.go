package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// Server provides an http.Server.
type Server struct {
	*http.Server
}

// NewServer creates and configures an APIServer serving all application routes.
func NewServer(app *Application) (*Server, error) {
	log.Println("Configuring server...")
	serverConfig := app.Config.Sub(serverConfigKey)

	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", serverConfig.GetInt("Port")),
		ReadTimeout:  time.Duration(serverConfig.GetInt64("ReadTimeout")) * time.Second,
		WriteTimeout: time.Duration(serverConfig.GetInt64("WriteTimeout")) * time.Second,
		Handler:      app.HTTPHandler,
	}

	return &Server{&srv}, nil
}

// Start runs ListenAndServe on the http.Server with graceful shutdown.
func (srv *Server) Start() chan bool {
	log.Println("Starting server...")
	doneChannel := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal("Server returned an error: ", err)
		}
	}()
	log.Printf("Listening on %s\n", srv.Addr)

	// dealing with termination
	go func() {
		sig := <-quit
		log.Println("Shutting down server... Reason:", sig)
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatal("Can't shut down the server: ", err)
		}
		close(doneChannel)
		log.Println("Server gracefully stopped")
	}()
	return doneChannel
}

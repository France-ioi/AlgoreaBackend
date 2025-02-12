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
	serverConfig := ServerConfig(app.Config)
	serverConfig.SetDefault("port", 8080)
	serverConfig.SetDefault("readTimeout", 60)
	serverConfig.SetDefault("writeTimeout", 60)

	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", serverConfig.GetInt("Port")),
		ReadTimeout:  time.Duration(serverConfig.GetInt64("ReadTimeout")) * time.Second,
		WriteTimeout: time.Duration(serverConfig.GetInt64("WriteTimeout")) * time.Second,
		Handler:      app.HTTPHandler,
	}

	return &Server{&srv}, nil
}

// Start runs ListenAndServe on the http.Server with graceful shutdown.
// The caller should close the done channel upon error or when the server has stopped.
func (srv *Server) Start() chan error {
	log.Println("Starting server...")
	doneChannel := make(chan error)
	serverErrChannel := make(chan error)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			serverErrChannel <- err
		} else {
			serverErrChannel <- nil
		}
	}()
	log.Printf("Listening on %s\n", srv.Addr)

	// dealing with termination
	go func() {
		select {
		case err := <-serverErrChannel:
			if err != nil {
				doneChannel <- fmt.Errorf("server returned an error: %v", err)
			} else {
				doneChannel <- nil
			}
		case sig := <-quit:
			log.Println("Shutting down server... Reason:", sig)
			shutdownErr := srv.Shutdown(context.Background())
			if serverErr := <-serverErrChannel; serverErr != nil {
				doneChannel <- fmt.Errorf("server returned an error: %v", serverErr)
			} else if shutdownErr != nil {
				doneChannel <- fmt.Errorf("can't shut down the server: %v", shutdownErr)
			} else {
				doneChannel <- nil
			}
			log.Println("Server gracefully stopped")
		}
		close(serverErrChannel)
		signal.Stop(quit)
		close(quit)
	}()
	return doneChannel
}

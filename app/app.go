package app

import (
	"time"

	"github.com/France-ioi/AlgoreaBackend/app/service/api"

	"github.com/France-ioi/AlgoreaBackend/app/database"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// New configures application resources and routes.
func New() (*chi.Mux, error) {

	if err := Config.Load(); err != nil {
		return nil, err
	}

	logger := NewLogger()

	db, err := database.DBConn(Config.Database)
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return nil, err
	}

	apiCtx := api.NewCtx(db)

	// Set up middlewares
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.DefaultCompress)
	router.Use(middleware.Timeout(time.Duration(Config.Timeout) * time.Second))

	router.Use(NewStructuredLogger(logger))
	router.Use(render.SetContentType(render.ContentTypeJSON))

	router.Use(corsConfig().Handler) // no need for CORS if served through the same domain

	router.Mount("/", apiCtx.Router())

	return router, nil
}

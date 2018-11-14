package app

import (
	"math/rand"
	"time"

	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service/api"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// Application is the core state of the app
type Application struct {
	HTTPHandler *chi.Mux
	Config      *config.Root
	Database    *database.DB
}

// New configures application resources and routes.
func New() (*Application, error) {

	config, err := config.Load()
	if err != nil {
		return nil, err
	}

	// Init the PRNG with current time
	rand.Seed(time.Now().UTC().UnixNano())

	logger := NewLogger()

	db, err := database.DBConn(config.Database)
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return nil, err
	}

	apiCtx, err := api.NewCtx(config, db)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// Set up middlewares
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.DefaultCompress)
	router.Use(middleware.Timeout(time.Duration(config.Timeout) * time.Second))

	router.Use(NewStructuredLogger(logger))
	router.Use(render.SetContentType(render.ContentTypeJSON))

	router.Use(corsConfig().Handler) // no need for CORS if served through the same domain

	router.Mount("/", apiCtx.Router())

	return &Application{router, config, db}, nil
}

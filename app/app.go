package app

import (
	"log"
	"math/rand"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/api"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

// Application is the core state of the app
type Application struct {
	HTTPHandler *chi.Mux
	Config      *config.Root
	Database    *database.DB
}

// New configures application resources and routes.
func New() (*Application, error) {
	var err error

	var conf *config.Root
	if conf, err = config.Load(); err != nil {
		return nil, err
	}

	// Init the PRNG with current time
	rand.Seed(time.Now().UTC().UnixNano())

	logger := logging.New(conf.Logging)
	log.SetOutput(logger.Writer()) // redirect the stdlib's log to our logger

	var db *database.DB
	if db, err = database.DBConn(conf.Database); err != nil {
		logger.WithField("module", "database").Error(err)
		return nil, err
	}

	var apiCtx *api.Ctx
	if apiCtx, err = api.NewCtx(conf, db); err != nil {
		logger.Error(err)
		return nil, err
	}

	// Set up middlewares
	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.DefaultCompress)
	router.Use(middleware.Timeout(time.Duration(conf.Timeout) * time.Second))

	router.Use(logging.NewStructuredLogger(logger))
	router.Use(render.SetContentType(render.ContentTypeJSON))

	router.Use(corsConfig().Handler) // no need for CORS if served through the same domain

	router.Mount(conf.Server.RootPath, apiCtx.Router())

	return &Application{router, conf, db}, nil
}

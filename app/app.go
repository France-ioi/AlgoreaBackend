package app

import (
	"math/rand"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	log "github.com/sirupsen/logrus"

	"github.com/France-ioi/AlgoreaBackend/app/api"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
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

	// Load config
	var conf *config.Root
	if conf, err = config.Load(); err != nil {
		log.Error(err)
		return nil, err
	}

	// Set up logger
	setupLogger(log.StandardLogger(), conf.Logging)

	// Init the PRNG with current time
	rand.Seed(time.Now().UTC().UnixNano())

	// Connect to DB
	var db *database.DB
	dbConfig := conf.Database.Connection.FormatDSN()
	if db, err = database.Open(dbConfig); err != nil {
		log.Error(err)
	}
	db.SetLogger(log.StandardLogger())

	// Set up API
	var apiCtx *api.Ctx
	if apiCtx, err = api.NewCtx(conf, db); err != nil {
		log.Error(err)
		return nil, err
	}

	// Set up middlewares
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(service.NewStructuredLogger(log.StandardLogger()))
	router.Use(middleware.Recoverer)
	router.Use(middleware.RealIP)
	router.Use(middleware.DefaultCompress)
	router.Use(middleware.Timeout(time.Duration(conf.Timeout) * time.Second))

	router.Use(render.SetContentType(render.ContentTypeJSON))

	router.Use(corsConfig().Handler) // no need for CORS if served through the same domain

	router.Mount(conf.Server.RootPath, apiCtx.Router())

	return &Application{router, conf, db}, nil
}

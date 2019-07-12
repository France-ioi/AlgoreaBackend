package app

import (
	"math/rand"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/France-ioi/AlgoreaBackend/app/api"
	"github.com/France-ioi/AlgoreaBackend/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/app/config"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	_ "github.com/France-ioi/AlgoreaBackend/app/doc" // for doc generation
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

// Application is the core state of the app
type Application struct {
	HTTPHandler *chi.Mux
	Config      *config.Root
	Database    *database.DB
	TokenConfig *token.Config
}

// New configures application resources and routes.
func New() (*Application, error) {
	var err error

	conf := config.Load() // exits on errors

	// Apply the config to the global logger
	logging.SharedLogger.Configure(conf.Logging)

	// Init the PRNG with current time
	rand.Seed(time.Now().UTC().UnixNano())

	var db *database.DB
	dbConfig := conf.Database.Connection.FormatDSN()
	if db, err = database.Open(dbConfig); err != nil {
		logging.WithField("module", "database").Error(err)
		return nil, err
	}

	tokenConfig, err := token.Initialize(&conf.Token)
	if err != nil {
		logging.Error(err)
		return nil, err
	}

	var apiCtx *api.Ctx
	if apiCtx, err = api.NewCtx(conf, db, tokenConfig); err != nil {
		logging.Error(err)
		return nil, err
	}

	// Set up middlewares
	router := chi.NewRouter()

	router.Use(middleware.RealIP)             // must be before logger or any middleware using remote IP
	router.Use(middleware.DefaultCompress)    // apply last on response
	router.Use(middleware.RequestID)          // must be before any middleware using the request id (the logger and the recoverer do)
	router.Use(logging.NewStructuredLogger()) //
	router.Use(middleware.Recoverer)          // must be before logger so that it an log panics

	router.Use(corsConfig().Handler) // no need for CORS if served through the same domain

	if appenv.IsEnvDev() {
		router.Mount("/debug", middleware.Profiler())
	}
	router.Mount(conf.Server.RootPath, apiCtx.Router())

	return &Application{
		HTTPHandler: router,
		Config:      conf,
		Database:    db,
		TokenConfig: tokenConfig,
	}, nil
}

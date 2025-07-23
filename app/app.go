// Package app contains the app server.
package app

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/spf13/viper"

	"github.com/France-ioi/AlgoreaBackend/v2/app/api"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	_ "github.com/France-ioi/AlgoreaBackend/v2/app/doc" // for doc generation
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/rand"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/version"
)

// Application is the core state of the app.
type Application struct {
	HTTPHandler *chi.Mux
	Config      *viper.Viper
	Database    *database.DB
	apiCtx      *api.Ctx
}

// New configures application resources and routes.
// loggerOptional is an optional logger to use, if not provided, a new logger will be created from the config.
func New(loggerOptional ...*logging.Logger) (*Application, error) {
	// Getting all configs, they will be used to init components and to be passed
	config := LoadConfig()
	application := &Application{}

	var randomBytes [8]byte
	_, err := crand.Read(randomBytes[:])
	if err != nil {
		panic("cannot seed the randomizer")
	}
	// Init the PRNG with a random value
	rand.Seed(int64(binary.LittleEndian.Uint64(randomBytes[:]))) //nolint:gosec // G115: we don't care if a big number becomes negative

	if err := application.Reset(config, loggerOptional...); err != nil {
		return nil, err
	}
	return application, nil
}

// Reset reinitializes the application with the given config.
// loggerOptional is an optional logger to use, if not provided, a new logger will be created from the config.
func (app *Application) Reset(config *viper.Viper, loggerOptional ...*logging.Logger) error {
	dbConfig, err := DBConfig(config)
	if err != nil {
		return fmt.Errorf("unable to load the 'database' configuration: %w", err)
	}
	authConfig := AuthConfig(config)
	loggingConfig := LoggingConfig(config)
	domainsConfig, err := DomainsConfig(config)
	if err != nil {
		return fmt.Errorf("unable to load the 'domain' configuration: %w", err)
	}
	tokenConfig, err := TokenConfig(config)
	if err != nil {
		return fmt.Errorf("unable to load the 'token' configuration: %w", err)
	}
	serverConfig := ServerConfig(config)

	logger := resolveOrCreateLogger(loggingConfig, loggerOptional)

	// Init DB
	if dbConfig.Params == nil {
		dbConfig.Params = make(map[string]string, 1)
	}
	dbConfig.Params["charset"] = "utf8mb4"
	ctx := logging.ContextWithLogger(context.Background(), logger)
	db, err := database.Open(ctx, dbConfig.FormatDSN())
	if err != nil {
		logger.WithContext(ctx).WithField("module", "database").Error(err)
		return err
	}

	if serverConfig.GetBool("disableResultsPropagation") {
		database.ProhibitResultsPropagation(db)
	}

	// Set up responder.
	render.Respond = service.AppResponder

	// Set up middlewares
	router := chi.NewRouter()

	router.Use(logging.ContextWithLoggerMiddleware(logger))
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(database.NewDataStore(app.Database).MergeContext(r.Context())))
		})
	})

	router.Use(version.AddVersionHeader)

	router.Use(middleware.RealIP) // must be before logger or any middleware using remote IP
	if serverConfig.GetBool("compress") {
		router.Use(middleware.DefaultCompress) // apply last on response
	}
	router.Use(middleware.RequestID)          // must be before any middleware using the request id (the logger and the recoverer do)
	router.Use(logging.NewStructuredLogger()) //
	router.Use(middleware.Recoverer)          // must be before logger so that it an log panics

	router.Use(corsConfig().Handler) // no need for CORS if served through the same domain
	router.Use(domain.Middleware(domainsConfig, serverConfig.GetString("domainOverride")))

	if appenv.IsEnvDev() {
		router.Mount("/debug", middleware.Profiler())
	}

	serverConfig.SetDefault("rootPath", "/")
	apiCtx, apiRouter := api.Router(db, serverConfig, authConfig, domainsConfig, tokenConfig)
	router.Mount(serverConfig.GetString("rootPath"), apiRouter)

	app.HTTPHandler = router
	app.Config = config
	if app.Database != nil {
		_ = app.Database.Close()
	}
	app.Database = db
	app.apiCtx = apiCtx
	return nil
}

func resolveOrCreateLogger(loggingConfig *viper.Viper, loggerOptional []*logging.Logger) (logger *logging.Logger) {
	if len(loggerOptional) > 0 && loggerOptional[0] != nil {
		logger = loggerOptional[0]
	} else {
		logger = logging.NewLoggerFromConfig(loggingConfig)
	}
	return logger
}

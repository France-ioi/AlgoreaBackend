// Package app contains the app server.
package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"

	"github.com/France-ioi/AlgoreaBackend/v2/app/api"
	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	_ "github.com/France-ioi/AlgoreaBackend/v2/app/doc" // for doc generation
	"github.com/France-ioi/AlgoreaBackend/v2/app/domain"
	"github.com/France-ioi/AlgoreaBackend/v2/app/event"
	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
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

	if err := application.Reset(config, loggerOptional...); err != nil {
		return nil, err
	}
	return application, nil
}

type appConfigs struct {
	db      *mysql.Config
	auth    *viper.Viper
	logging *viper.Viper
	domains []domain.ConfigItem
	token   *token.Config
	server  *viper.Viper
	event   *viper.Viper
}

func loadAppConfigs(config *viper.Viper) (*appConfigs, error) {
	dbConfig, err := DBConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to load the 'database' configuration: %w", err)
	}
	domainsConfig, err := DomainsConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to load the 'domain' configuration: %w", err)
	}
	tokenConfig, err := TokenConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to load the 'token' configuration: %w", err)
	}
	return &appConfigs{
		db:      dbConfig,
		auth:    AuthConfig(config),
		logging: LoggingConfig(config),
		domains: domainsConfig,
		token:   tokenConfig,
		server:  ServerConfig(config),
		event:   EventConfig(config),
	}, nil
}

// Reset reinitializes the application with the given config.
// loggerOptional is an optional logger to use, if not provided, a new logger will be created from the config.
func (app *Application) Reset(config *viper.Viper, loggerOptional ...*logging.Logger) error {
	configs, err := loadAppConfigs(config)
	if err != nil {
		return err
	}

	logger := resolveOrCreateLogger(configs.logging, loggerOptional)

	ctx := logging.ContextWithLogger(context.Background(), logger)

	// Init event dispatcher (nil if not configured)
	eventDispatcher, err := event.NewDispatcherFromConfig(ctx, configs.event)
	if err != nil {
		logger.WithContext(ctx).WithField("module", "event").Error(err)
		return fmt.Errorf("unable to initialize event dispatcher: %w", err)
	}
	eventInstance := event.GetInstance(configs.event)

	// Init DB
	if configs.db.Params == nil {
		configs.db.Params = make(map[string]string, 1)
	}
	configs.db.Params["charset"] = "utf8mb4"
	db, err := database.Open(ctx, configs.db.FormatDSN())
	if err != nil {
		logger.WithContext(ctx).WithField("module", "database").Error(err)
		return err
	}

	if configs.server.GetBool("disableResultsPropagation") {
		database.ProhibitResultsPropagation(db)
	}

	// Set up responder.
	render.Respond = service.AppResponder

	// Set up middlewares
	router := chi.NewRouter()

	router.Use(logging.ContextWithLoggerMiddleware(logger))
	router.Use(event.ContextWithDispatcherMiddleware(eventDispatcher))
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := database.NewDataStore(app.Database).MergeContext(r.Context())
			ctx = event.ContextWithConfig(ctx, eventInstance)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	router.Use(version.AddVersionHeader)

	router.Use(middleware.RealIP) // must be before logger or any middleware using remote IP
	if configs.server.GetBool("compress") {
		router.Use(middleware.DefaultCompress) // apply last on response
	}
	router.Use(middleware.RequestID)          // must be before any middleware using the request id (the logger and the recoverer do)
	router.Use(logging.NewStructuredLogger()) //
	router.Use(middleware.Recoverer)          // must be before logger so that it an log panics

	router.Use(corsConfig().Handler) // no need for CORS if served through the same domain
	router.Use(domain.Middleware(configs.domains, configs.server.GetString("domainOverride")))

	if appenv.IsEnvDev() {
		router.Mount("/debug", middleware.Profiler())
	}

	configs.server.SetDefault("rootPath", "/")
	apiCtx, apiRouter := api.Router(db, configs.server, configs.auth, configs.domains, configs.token)
	router.Mount(configs.server.GetString("rootPath"), apiRouter)

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
		logger.Configure(loggingConfig)
	} else {
		logger = logging.NewLoggerFromConfig(loggingConfig)
	}
	return logger
}

package app

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/spf13/viper"

	"github.com/France-ioi/AlgoreaBackend/app/api"
	"github.com/France-ioi/AlgoreaBackend/app/appenv"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	_ "github.com/France-ioi/AlgoreaBackend/app/doc" // for doc generation
	"github.com/France-ioi/AlgoreaBackend/app/domain"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

// Application is the core state of the app
type Application struct {
	HTTPHandler *chi.Mux
	Config      *viper.Viper
	Database    *database.DB
	apiCtx      *api.Ctx
}

// New configures application resources and routes.
func New() (*Application, error) {

	// Getting all configs, they will be used to init components and to be passed
	config := LoadConfig()
	dbConfig, err := DBConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to load the 'database' configuration: %w", err)
	}
	authConfig := AuthConfig(config)
	loggingConfig := LoggingConfig(config)
	domainsConfig, err := DomainsConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to load the 'domain' configuration: %w", err)
	}
	tokenConfig, err := TokenConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to load the 'token' configuration: %w", err)
	}
	serverConfig := ServerConfig(config)

	// Apply the config to the global logger
	logging.SharedLogger.Configure(loggingConfig)

	// Init the PRNG with current time
	rand.Seed(time.Now().UTC().UnixNano())

	// Init DB
	db, err := database.Open(dbConfig.FormatDSN())
	if err != nil {
		logging.WithField("module", "database").Error(err)
		return nil, err
	}

	// Set up middlewares
	router := chi.NewRouter()

	router.Use(middleware.RealIP) // must be before logger or any middleware using remote IP
	if serverConfig.GetBool("compress") {
		router.Use(middleware.DefaultCompress) // apply last on response
	}
	router.Use(middleware.RequestID)          // must be before any middleware using the request id (the logger and the recoverer do)
	router.Use(logging.NewStructuredLogger()) //
	router.Use(middleware.Recoverer)          // must be before logger so that it an log panics

	router.Use(corsConfig().Handler) // no need for CORS if served through the same domain
	router.Use(domain.Middleware(domainsConfig))

	if appenv.IsEnvDev() {
		router.Mount("/debug", middleware.Profiler())
	}

	serverConfig.SetDefault("rootpath", "/")
	apiCtx, apiRouter := api.Router(db, serverConfig, authConfig, domainsConfig, tokenConfig)
	router.Mount(serverConfig.GetString("RootPath"), apiRouter)

	return &Application{
		HTTPHandler: router,
		Config:      config,
		Database:    db,
		apiCtx:      apiCtx,
	}, nil
}

// CheckConfig checks that the database contains all the data needed by the config
func (app *Application) CheckConfig() error {
	groupStore := database.NewDataStore(app.Database).Groups()
	groupGroupStore := groupStore.ActiveGroupGroups()
	domainsConfig, err := DomainsConfig(app.Config)
	if err != nil {
		return fmt.Errorf("unable to unmarshal domains config: %w", err)
	}
	for _, domainConfig := range domainsConfig {
		for _, spec := range []struct {
			name string
			id   int64
		}{
			{name: "Root", id: domainConfig.RootGroup},
			{name: "RootSelf", id: domainConfig.RootSelfGroup},
			{name: "RootTemp", id: domainConfig.RootTempGroup},
		} {
			hasRows, err := groupStore.ByID(spec.id).HasRows()
			if err != nil {
				return err
			}
			if !hasRows {
				return fmt.Errorf("no %s group for domain %q", spec.name, domainConfig.Domains[0])
			}
		}

		for _, spec := range []struct {
			parentName string
			childName  string
			parentID   int64
			childID    int64
		}{
			{parentName: "Root", childName: "RootSelf", parentID: domainConfig.RootGroup, childID: domainConfig.RootSelfGroup},
			{parentName: "RootSelf", childName: "RootTemp", parentID: domainConfig.RootSelfGroup, childID: domainConfig.RootTempGroup},
		} {
			hasRows, err := groupGroupStore.
				Where("parent_group_id = ?", spec.parentID).
				Where("child_group_id = ?", spec.childID).
				Select("1").Limit(1).HasRows()
			if err != nil {
				return err
			}
			if !hasRows {
				return fmt.Errorf("no %s -> %s link in groups_groups for domain %q",
					spec.parentName, spec.childName, domainConfig.Domains[0])
			}
		}
	}
	return nil
}

// CreateMissingData fills the database with required data (if missing)
func (app *Application) CreateMissingData() error {
	return database.NewDataStore(app.Database).InTransaction(app.insertRootGroupsAndRelations)
}

func (app *Application) insertRootGroupsAndRelations(store *database.DataStore) error {

	groupStore := store.Groups()
	groupGroupStore := store.GroupGroups()
	var relationsToCreate []map[string]interface{}
	var inserted bool
	domainsConfig, err := DomainsConfig(app.Config)
	if err != nil {
		return fmt.Errorf("unable to unmarshal domains config: %w", err)
	}
	for _, domainConfig := range domainsConfig {
		domainConfig := domainConfig
		insertedForDomain, err := insertRootGroups(groupStore, &domainConfig)
		if err != nil {
			return err
		}
		inserted = inserted || insertedForDomain
		for _, spec := range []map[string]interface{}{
			{"parent_group_id": domainConfig.RootGroup, "child_group_id": domainConfig.RootSelfGroup},
			{"parent_group_id": domainConfig.RootSelfGroup, "child_group_id": domainConfig.RootTempGroup},
		} {
			found, err := groupGroupStore.
				Where("parent_group_id = ?", spec["parent_group_id"]).
				Where("child_group_id = ?", spec["child_group_id"]).
				Limit(1).HasRows()
			if err != nil {
				return err
			}
			if !found {
				relationsToCreate = append(relationsToCreate, spec)
			}
		}
		if len(relationsToCreate) > 0 || inserted {
			return groupStore.GroupGroups().CreateRelationsWithoutChecking(relationsToCreate)
		}
	}
	return nil
}

func insertRootGroups(groupStore *database.GroupStore, domainConfig *domain.ConfigItem) (bool, error) {
	var inserted bool
	for _, spec := range []struct {
		name string
		id   int64
	}{
		{name: "Root", id: domainConfig.RootGroup},
		{name: "RootSelf", id: domainConfig.RootSelfGroup},
		{name: "RootTemp", id: domainConfig.RootTempGroup},
	} {
		found, err := groupStore.ByID(spec.id).Where("type = 'Base'").
			Where("name = ?", spec.name).
			Where("text_id = ?", spec.name).Limit(1).HasRows()
		if err != nil {
			return false, err
		}
		if !found {
			if err := groupStore.InsertMap(map[string]interface{}{
				"id": spec.id, "type": "Base", "name": spec.name, "text_id": spec.name,
			}); err != nil {
				return false, err
			}
			inserted = true
		}
	}
	return inserted, nil
}

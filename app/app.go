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
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

// Application is the core state of the app
type Application struct {
	HTTPHandler *chi.Mux
	Config      *viper.Viper
	Database    *database.DB
}

type ServerConfig struct {
	Port         int32
	ReadTimeout  int32
	WriteTimeout int32
	Domain       string
	RootPath     string
}

// New configures application resources and routes.
func New() (*Application, error) {
	var err error

	config := LoadConfig()

	// Apply the config to the global logger
	logging.SharedLogger.Configure(config.Sub(loggingConfigKey))

	// Init the PRNG with current time
	rand.Seed(time.Now().UTC().UnixNano())

	// Init DB
	var db *database.DB
	if db, err = database.OpenFromConfig(config.Sub(databaseConfigKey)); err != nil {
		logging.WithField("module", "database").Error(err)
		return nil, err
	}

	// Load token config
	tokenConfig, err := token.Initialize(config.Sub(tokenConfigKey))
	if err != nil {
		logging.Error(err)
		return nil, err
	}

	// "Server" config
	serverConfig := config.Sub(serverConfigKey)
	serverConfig.SetDefault("rootpath", "/")
	serverConfig.SetDefault("port", 8080)
	serverConfig.SetDefault("readTimeout", 60)
	serverConfig.SetDefault("writeTimeout", 60)

	apiCtx := api.NewCtx(db, serverConfig, config.Sub(domainsConfigKey), config.Sub(authConfigKey), tokenConfig)

	// Set up middlewares
	router := chi.NewRouter()

	router.Use(middleware.RealIP)             // must be before logger or any middleware using remote IP
	router.Use(middleware.DefaultCompress)    // apply last on response
	router.Use(middleware.RequestID)          // must be before any middleware using the request id (the logger and the recoverer do)
	router.Use(logging.NewStructuredLogger()) //
	router.Use(middleware.Recoverer)          // must be before logger so that it an log panics

	router.Use(corsConfig().Handler) // no need for CORS if served through the same domain
	router.Use(domain.Middleware(domain.ParseConfig(config.Sub(domainsConfigKey))))

	if appenv.IsEnvDev() {
		router.Mount("/debug", middleware.Profiler())
	}

	router.Mount(serverConfig.GetString("RootPath"), apiCtx.Router())

	return &Application{
		HTTPHandler: router,
		Config:      config,
		Database:    db,
	}, nil
}

// CheckConfig checks that the database contains all the data needed by the config
func (app *Application) CheckConfig() error {
	groupStore := database.NewDataStore(app.Database).Groups()
	groupGroupStore := groupStore.ActiveGroupGroups()
	for _, domainConfig := range domain.ParseConfig(app.Config.Sub(domainsConfigKey)) {
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

	appDomainConfig := domain.ParseConfig(app.Config.Sub(domainsConfigKey))
	groupStore := store.Groups()
	groupGroupStore := store.GroupGroups()
	var relationsToCreate []map[string]interface{}
	var inserted bool
	for _, domainConfig := range appDomainConfig {
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

func insertRootGroups(groupStore *database.GroupStore, domainConfig *domain.AppConfigItem) (bool, error) {
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

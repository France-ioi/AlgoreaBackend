package database

import (
	"context"
	"database/sql"
	"sync/atomic"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database/mysqldb"
	"github.com/France-ioi/AlgoreaBackend/v2/app/rand"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

// DataStore gather all stores for database operations on business data.
type DataStore struct {
	*DB
	tableName string
}

// NewDataStore returns a DataStore.
func NewDataStore(conn *DB) *DataStore {
	return &DataStore{DB: conn}
}

// NewDataStoreWithContext returns a new DataStore with the given context.
func NewDataStoreWithContext(ctx context.Context, db *DB) *DataStore {
	return &DataStore{DB: cloneDBWithNewContext(ctx, db)}
}

// NewDataStoreWithTable returns a specialized DataStore.
func NewDataStoreWithTable(conn *DB, tableName string) *DataStore {
	return &DataStore{conn.Table(tableName), tableName}
}

// ProhibitResultsPropagation marks the context inside the DB connection as prohibiting the propagation of results.
func ProhibitResultsPropagation(conn *DB) {
	prohibitedPropagations := getProhibitedPropagationsFromContext(conn.ctx)
	prohibitedPropagations.Results = true
	conn.ctx = context.WithValue(conn.ctx, prohibitedPropagationsContextKey, prohibitedPropagations)
}

// IsResultsPropagationProhibited returns true if the propagation of results is prohibited in the context of the current DB connection.
func (s *DataStore) IsResultsPropagationProhibited() bool {
	return getProhibitedPropagationsFromContext(s.DB.ctx).Results
}

func getProhibitedPropagationsFromContext(ctx context.Context) propagationsBitField {
	prohibitedPropagations := ctx.Value(prohibitedPropagationsContextKey)
	if prohibitedPropagations == nil {
		return propagationsBitField{}
	}
	return prohibitedPropagations.(propagationsBitField)
}

// MergeContext returns a new context based on the given one, with DB-related values copied
// from the context of the current DB connection.
func (s *DataStore) MergeContext(ctx context.Context) context.Context {
	prohibitedPropagations := getProhibitedPropagationsFromContext(s.DB.ctx)
	return context.WithValue(ctx, prohibitedPropagationsContextKey, prohibitedPropagations)
}

// ActiveGroupGroups returns a GroupGroupStore working with the `groups_groups_active` view.
func (s *DataStore) ActiveGroupGroups() *GroupGroupStore {
	return &GroupGroupStore{NewDataStoreWithTable(s.DB, "groups_groups_active")}
}

// Answers returns a AnswerStore.
func (s *DataStore) Answers() *AnswerStore {
	return &AnswerStore{NewDataStoreWithTable(s.DB, "answers")}
}

// Attempts returns a AttemptStore.
func (s *DataStore) Attempts() *AttemptStore {
	return &AttemptStore{NewDataStoreWithTable(s.DB, "attempts")}
}

// Gradings returns a GradingStore.
func (s *DataStore) Gradings() *GradingStore {
	return &GradingStore{NewDataStoreWithTable(s.DB, "gradings")}
}

// Users returns a UserStore.
func (s *DataStore) Users() *UserStore {
	return &UserStore{NewDataStoreWithTable(s.DB, "users")}
}

// Items returns a ItemStore.
func (s *DataStore) Items() *ItemStore {
	return &ItemStore{NewDataStoreWithTable(s.DB, "items")}
}

// Groups returns a GroupStore.
func (s *DataStore) Groups() *GroupStore {
	return &GroupStore{NewDataStoreWithTable(s.DB, "groups")}
}

// GroupAncestors returns a GroupAncestorStore.
func (s *DataStore) GroupAncestors() *GroupAncestorStore {
	return &GroupAncestorStore{NewDataStoreWithTable(s.DB, "groups_ancestors")}
}

// ActiveGroupAncestors returns a GroupAncestorStore working with the `groups_ancestors_active` view.
func (s *DataStore) ActiveGroupAncestors() *GroupAncestorStore {
	return &GroupAncestorStore{NewDataStoreWithTable(s.DB, "groups_ancestors_active")}
}

// GroupGroups returns a GroupGroupStore.
func (s *DataStore) GroupGroups() *GroupGroupStore {
	return &GroupGroupStore{NewDataStoreWithTable(s.DB, "groups_groups")}
}

// GroupMembershipChanges returns a GroupMembershipChangeStore.
func (s *DataStore) GroupMembershipChanges() *GroupMembershipChangeStore {
	return &GroupMembershipChangeStore{NewDataStoreWithTable(s.DB, "group_membership_changes")}
}

// GroupPendingRequests returns a GroupPendingRequestStore.
func (s *DataStore) GroupPendingRequests() *GroupPendingRequestStore {
	return &GroupPendingRequestStore{NewDataStoreWithTable(s.DB, "group_pending_requests")}
}

// GroupItemAdditionalTimes returns a GroupItemAdditionalTimeStore.
func (s *DataStore) GroupItemAdditionalTimes() *GroupItemAdditionalTimeStore {
	return &GroupItemAdditionalTimeStore{NewDataStoreWithTable(s.DB, "group_item_additional_times")}
}

// GroupManagers returns a GroupManagerStore.
func (s *DataStore) GroupManagers() *GroupManagerStore {
	return &GroupManagerStore{NewDataStoreWithTable(s.DB, "group_managers")}
}

// Permissions returns a PermissionGeneratedStore.
func (s *DataStore) Permissions() *PermissionGeneratedStore {
	return &PermissionGeneratedStore{NewDataStoreWithTable(s.DB, "permissions_generated AS permissions")}
}

// PermissionsGranted returns a PermissionGrantedStore.
func (s *DataStore) PermissionsGranted() *PermissionGrantedStore {
	return &PermissionGrantedStore{NewDataStoreWithTable(s.DB, "permissions_granted")}
}

// ItemAncestors returns an ItemAncestorStore.
func (s *DataStore) ItemAncestors() *ItemAncestorStore {
	return &ItemAncestorStore{NewDataStoreWithTable(s.DB, "items_ancestors")}
}

// ItemStrings returns an ItemStringStore.
func (s *DataStore) ItemStrings() *ItemStringStore {
	return &ItemStringStore{NewDataStoreWithTable(s.DB, "items_strings")}
}

// ItemItems returns an ItemItemStore.
func (s *DataStore) ItemItems() *ItemItemStore {
	return &ItemItemStore{NewDataStoreWithTable(s.DB, "items_items")}
}

// ItemDependencies returns an ItemDependencyStore.
func (s *DataStore) ItemDependencies() *ItemDependencyStore {
	return &ItemDependencyStore{NewDataStoreWithTable(s.DB, "item_dependencies")}
}

// Languages returns a LanguageStore.
func (s *DataStore) Languages() *LanguageStore {
	return &LanguageStore{NewDataStoreWithTable(s.DB, "languages")}
}

// Platforms returns a PlatformStore.
func (s *DataStore) Platforms() *PlatformStore {
	return &PlatformStore{NewDataStoreWithTable(s.DB, "platforms")}
}

// Results returns a ResultStore.
func (s *DataStore) Results() *ResultStore {
	return &ResultStore{NewDataStoreWithTable(s.DB, "results")}
}

// Sessions returns a SessionStore.
func (s *DataStore) Sessions() *SessionStore {
	return &SessionStore{NewDataStoreWithTable(s.DB, "sessions")}
}

// AccessTokens returns a AccessTokenStore.
func (s *DataStore) AccessTokens() *AccessTokenStore {
	return &AccessTokenStore{NewDataStoreWithTable(s.DB, "access_tokens")}
}

// Threads returns a ThreadStore.
func (s *DataStore) Threads() *ThreadStore {
	return &ThreadStore{NewDataStoreWithTable(s.DB, "threads")}
}

// UserBatches returns a UserBatchStore.
func (s *DataStore) UserBatches() *UserBatchStore {
	return &UserBatchStore{NewDataStoreWithTable(s.DB, "user_batches_v2")}
}

// UserBatchPrefixes returns a UserBatchPrefixStore.
func (s *DataStore) UserBatchPrefixes() *UserBatchPrefixStore {
	return &UserBatchPrefixStore{NewDataStoreWithTable(s.DB, "user_batch_prefixes")}
}

// NewID generates a positive random int64 to be used as id
// !!! To be safe, the insertion should be retried if the id conflicts with an existing entry.
func (s *DataStore) NewID() int64 {
	// gen a 63-bits number as we want unsigned number stored in a 64-bits signed DB attribute
	return rand.Int63()
}

type propagationsBitField struct {
	Permissions bool
	Results     bool
}

type dbContextKey string

const (
	awaitingPropagationsContextKey   = dbContextKey("awaitingPropagations")
	prohibitedPropagationsContextKey = dbContextKey("prohibitedPropagations")
	retryEachTransactionContextKey   = dbContextKey("retryEachTransaction")
	propagationsAreSyncContextKey    = dbContextKey("propagationsAreSync")
	logErrorAsInfoFuncContextKey     = dbContextKey("logErrorAsInfoFunc")
)

var (
	onStartOfTransactionToBeRetriedForcefullyHook atomic.Value
	onForcefulRetryOfTransactionHook              atomic.Value
)

func init() { //nolint:gochecknoinits // this is an initialization function to store the default hooks
	onStartOfTransactionToBeRetriedForcefullyHook.Store(func() {})
	onForcefulRetryOfTransactionHook.Store(func() {})
}

// InTransaction executes the given function in a transaction and commits.
// If a propagation is scheduled, it will be run after the transaction commit,
// so we can run each step of the propagation in a separate transaction.
//
// For testing purposes, it is possible to force this method to retry the transaction once
// by providing a context created with ContextWithTransactionRetrying.
func (s *DataStore) InTransaction(txFunc func(*DataStore) error, txOptions ...*sql.TxOptions) error {
	s.DB.ctx = context.WithValue(s.DB.ctx, awaitingPropagationsContextKey, &propagationsBitField{})
	var retried bool

	err := s.inTransaction(func(db *DB) error {
		shouldForceTransactionRetry := s.DB.ctx.Value(retryEachTransactionContextKey) != nil && !retried

		dataStore := NewDataStoreWithTable(db, s.tableName)
		if shouldForceTransactionRetry {
			onStartOfTransactionToBeRetriedForcefullyHook.Load().(func())()
		}

		err := txFunc(dataStore)

		if err == nil && shouldForceTransactionRetry {
			retried = true
			onForcefulRetryOfTransactionHook.Load().(func())()
			return &mysql.MySQLError{
				Number: uint16(mysqldb.DeadlockError),
			}
		}

		return err
	}, txOptions...)
	if err != nil {
		return err
	}

	propagationsToRun := s.ctx.Value(awaitingPropagationsContextKey).(*propagationsBitField)
	prohibitedPropagations := getProhibitedPropagationsFromContext(s.ctx)

	if propagationsToRun.Permissions && !prohibitedPropagations.Permissions {
		propagationsToRun.Permissions = false
		s.PermissionsGranted().computeAllAccess()
	}
	if propagationsToRun.Results && !prohibitedPropagations.Results {
		propagationsToRun.Results = false
		err = s.Results().processResultsRecomputeForItemsAndPropagate()
	}

	return err
}

// EnsureTransaction executes the given function in a transaction and commits. If a transaction is already started,
// it will execute the function in the current transaction.
func (s *DataStore) EnsureTransaction(txFunc func(*DataStore) error, txOptions ...*sql.TxOptions) error {
	if s.IsInTransaction() {
		return txFunc(s)
	}
	return s.InTransaction(txFunc, txOptions...)
}

// SetOnStartOfTransactionToBeRetriedForcefullyHook sets a hook to be called on the start
// of a transaction that will be forcefully retried.
// For testing purposes only.
func SetOnStartOfTransactionToBeRetriedForcefullyHook(hook func()) {
	onStartOfTransactionToBeRetriedForcefullyHook.Store(hook)
}

// SetOnForcefulRetryOfTransactionHook sets a hook to be called on the retry
// of a forcefully retried transaction.
// For testing purposes only.
func SetOnForcefulRetryOfTransactionHook(hook func()) {
	onForcefulRetryOfTransactionHook.Store(hook)
}

// SetPropagationsModeToSync sets the mode of propagations to synchronous.
// In this mode, the propagation of permissions and results will be done synchronously
// before the transaction commit.
func (s *DataStore) SetPropagationsModeToSync() (err error) {
	s.mustBeInTransaction()

	defer recoverPanics(&err)

	mustNotBeError(s.Exec("SET @synchronous_propagations_connection_id = CONNECTION_ID()").Error())

	s.DB = cloneDBWithNewContext(context.WithValue(s.DB.ctx, propagationsAreSyncContextKey, true), s.DB)
	return nil
}

// ScheduleResultsPropagation schedules a run of ResultStore::processResultsRecomputeForItemsAndPropagate() after the transaction commit.
func (s *DataStore) ScheduleResultsPropagation() {
	s.mustBeInTransaction()

	propagationsToRun := s.DB.ctx.Value(awaitingPropagationsContextKey).(*propagationsBitField)
	propagationsToRun.Results = true
}

// SchedulePermissionsPropagation schedules a run of the groups ancestors propagation after the transaction commit.
func (s *DataStore) SchedulePermissionsPropagation() {
	s.mustBeInTransaction()

	propagationsToRun := s.DB.ctx.Value(awaitingPropagationsContextKey).(*propagationsBitField)
	propagationsToRun.Permissions = true
}

// WithForeignKeyChecksDisabled executes the given function with foreign keys checking disabled
// (wraps it up in a transaction if no transaction started).
func (s *DataStore) WithForeignKeyChecksDisabled(blockFunc func(*DataStore) error, txOptions ...*sql.TxOptions) error {
	return s.withForeignKeyChecksDisabled(func(db *DB) error {
		return blockFunc(NewDataStoreWithTable(db, s.tableName))
	}, txOptions...)
}

// IsInTransaction returns true if the store operates in a DB transaction at the moment.
func (s *DataStore) IsInTransaction() bool {
	return s.DB.isInTransaction()
}

// WithNamedLock wraps the given function in GET_LOCK/RELEASE_LOCK.
func (s *DataStore) WithNamedLock(lockName string, timeout time.Duration, funcToCall func(*DataStore) error) error {
	return s.withNamedLock(lockName, timeout, func(db *DB) error {
		return funcToCall(NewDataStoreWithTable(db, s.tableName))
	})
}

// WithExclusiveWriteLock converts "SELECT ..." statement into "SELECT ... FOR UPDATE" statement.
// For existing rows, it will read the latest committed data (instead of the data from the repeatable-read snapshot)
// and acquire an exclusive lock on them, preventing other transactions from modifying them and
// even from getting exclusive/shared locks on them. For non-existing rows, it works similarly to a shared lock (FOR SHARE).
func (s *DataStore) WithExclusiveWriteLock() *DataStore {
	return NewDataStore(s.DB.WithExclusiveWriteLock())
}

// WithSharedWriteLock converts "SELECT ..." statement into "SELECT ... FOR SHARE" statement.
// For existing rows, it will read the latest committed data (instead of the data from the repeatable-read snapshot)
// and acquire a shared lock on them, preventing other transactions from modifying them.
func (s *DataStore) WithSharedWriteLock() *DataStore {
	return NewDataStore(s.DB.WithSharedWriteLock())
}

// WithCustomWriteLocks converts "SELECT ..." statement into "SELECT ... FOR SHARE OF ... FOR UPDATE ..." statement.
// For existing rows, it will read the latest committed data for the listed tables
// (instead of the data from the repeatable-read snapshot) and acquire shared/exclusive locks on them,
// preventing other transactions from modifying them.
func (s *DataStore) WithCustomWriteLocks(shared, exclusive *golang.Set[string]) *DataStore {
	return NewDataStore(s.DB.WithCustomWriteLocks(shared, exclusive))
}

// ByID returns a composable query for filtering by _table_.id.
func (s *DataStore) ByID(id int64) *DB {
	if s.tableName == "" {
		panic("method ByID() called for abstract DataStore")
	}
	return s.Where(s.tableName+".id = ?", id)
}

// RetryOnDuplicatePrimaryKeyError will retry the given function on getting duplicate entry errors
// for primary keys.
func (s *DataStore) RetryOnDuplicatePrimaryKeyError(tableName string, f func(store *DataStore) error) error {
	return s.DB.retryOnDuplicatePrimaryKeyError(tableName, func(db *DB) error {
		return f(NewDataStore(db))
	})
}

// RetryOnDuplicateKeyError will retry the given function on getting duplicate entry errors
// for the given key.
func (s *DataStore) RetryOnDuplicateKeyError(tableName, keyName, nameInError string, f func(store *DataStore) error) error {
	return s.DB.retryOnDuplicateKeyError(tableName, keyName, nameInError, func(db *DB) error {
		return f(NewDataStore(db))
	})
}

// InsertMap reads fields from the given map and inserts the values which have been set
// into the store's table.
func (s *DataStore) InsertMap(dataMap map[string]interface{}) error {
	return s.DB.insertMaps(s.tableName, []map[string]interface{}{dataMap})
}

// InsertMaps reads fields from the given map and inserts the values set in the first row (so all the rows should have the same keys)
// into the store's table.
func (s *DataStore) InsertMaps(dataMaps []map[string]interface{}) error {
	return s.DB.insertMaps(s.tableName, dataMaps)
}

// InsertOrUpdateMap reads fields from the given map and inserts the values which have been set
// into the store's table (like InsertMap does). If it is a duplicate, the listed columns will be updated.
func (s *DataStore) InsertOrUpdateMap(dataMap map[string]interface{}, updateColumns []string) error {
	return s.DB.insertOrUpdateMaps(s.tableName, []map[string]interface{}{dataMap}, updateColumns)
}

// InsertOrUpdateMaps reads fields from the given maps and inserts the values set in the first row
// (so all the maps should have the same keys)
// into the store's table (like InsertMaps does). If it is a duplicate, the listed columns will be updated.
// If updateColumns is nil, all the columns in dataMaps will be updated.
func (s *DataStore) InsertOrUpdateMaps(dataMap []map[string]interface{}, updateColumns []string) error {
	return s.DB.insertOrUpdateMaps(s.tableName, dataMap, updateColumns)
}

// ContextWithTransactionRetrying wraps the given context with a flag to retry each transaction once.
// Use it for testing purposes only.
func ContextWithTransactionRetrying(ctx context.Context) context.Context {
	return context.WithValue(ctx, retryEachTransactionContextKey, true)
}

func (s *DataStore) arePropagationsSync() bool {
	propagationsAreSync, _ := s.ctx.Value(propagationsAreSyncContextKey).(bool)
	return propagationsAreSync
}

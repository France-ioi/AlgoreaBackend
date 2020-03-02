package database

import (
	"math/rand"
	"time"
)

// DataStore gather all stores for database operations on business data
type DataStore struct {
	*DB
	tableName string
}

// NewDataStore returns a DataStore
func NewDataStore(conn *DB) *DataStore {
	return &DataStore{DB: conn}
}

// NewDataStoreWithTable returns a specialized DataStore
func NewDataStoreWithTable(conn *DB, tableName string) *DataStore {
	return &DataStore{conn.Table(tableName), tableName}
}

// Answers returns a AnswerStore
func (s *DataStore) Answers() *AnswerStore {
	return &AnswerStore{NewDataStoreWithTable(s.DB, "answers")}
}

// Attempts returns a AttemptStore
func (s *DataStore) Attempts() *AttemptStore {
	return &AttemptStore{NewDataStoreWithTable(s.DB, "attempts")}
}

// Gradings returns a GradingStore
func (s *DataStore) Gradings() *GradingStore {
	return &GradingStore{NewDataStoreWithTable(s.DB, "gradings")}
}

// Users returns a UserStore
func (s *DataStore) Users() *UserStore {
	return &UserStore{NewDataStoreWithTable(s.DB, "users")}
}

// Items returns a ItemStore
func (s *DataStore) Items() *ItemStore {
	return &ItemStore{NewDataStoreWithTable(s.DB, "items")}
}

// Groups returns a GroupStore
func (s *DataStore) Groups() *GroupStore {
	return &GroupStore{NewDataStoreWithTable(s.DB, "groups")}
}

// GroupAncestors returns a GroupAncestorStore
func (s *DataStore) GroupAncestors() *GroupAncestorStore {
	return &GroupAncestorStore{NewDataStoreWithTable(s.DB, "groups_ancestors")}
}

// ActiveGroupAncestors returns a GroupAncestorStore working with the `groups_ancestors_active` view
func (s *DataStore) ActiveGroupAncestors() *GroupAncestorStore {
	return &GroupAncestorStore{NewDataStoreWithTable(s.DB, "groups_ancestors_active")}
}

// GroupGroups returns a GroupGroupStore
func (s *DataStore) GroupGroups() *GroupGroupStore {
	return &GroupGroupStore{NewDataStoreWithTable(s.DB, "groups_groups")}
}

// ActiveGroupGroups returns a GroupGroupStore working with the `groups_groups_active` view
func (s *DataStore) ActiveGroupGroups() *GroupGroupStore {
	return &GroupGroupStore{NewDataStoreWithTable(s.DB, "groups_groups_active")}
}

// GroupMembershipChanges returns a GroupMembershipChangeStore
func (s *DataStore) GroupMembershipChanges() *GroupMembershipChangeStore {
	return &GroupMembershipChangeStore{NewDataStoreWithTable(s.DB, "group_membership_changes")}
}

// GroupPendingRequests returns a GroupPendingRequestStore
func (s *DataStore) GroupPendingRequests() *GroupPendingRequestStore {
	return &GroupPendingRequestStore{NewDataStoreWithTable(s.DB, "group_pending_requests")}
}

// GroupContestItems returns a GroupContestItemStore
func (s *DataStore) GroupContestItems() *GroupContestItemStore {
	return &GroupContestItemStore{NewDataStoreWithTable(s.DB, "groups_contest_items")}
}

// GroupManagers returns a GroupManagerStore
func (s *DataStore) GroupManagers() *GroupManagerStore {
	return &GroupManagerStore{NewDataStoreWithTable(s.DB, "group_managers")}
}

// Permissions returns a PermissionGeneratedStore
func (s *DataStore) Permissions() *PermissionGeneratedStore {
	return &PermissionGeneratedStore{NewDataStoreWithTable(s.DB, "permissions_generated AS permissions")}
}

// PermissionsGranted returns a PermissionGrantedStore
func (s *DataStore) PermissionsGranted() *PermissionGrantedStore {
	return &PermissionGrantedStore{NewDataStoreWithTable(s.DB, "permissions_granted")}
}

// ItemAncestors returns an ItemAncestorStore
func (s *DataStore) ItemAncestors() *ItemAncestorStore {
	return &ItemAncestorStore{NewDataStoreWithTable(s.DB, "items_ancestors")}
}

// ItemStrings returns an ItemStringStore
func (s *DataStore) ItemStrings() *ItemStringStore {
	return &ItemStringStore{NewDataStoreWithTable(s.DB, "items_strings")}
}

// ItemItems returns an ItemItemStore
func (s *DataStore) ItemItems() *ItemItemStore {
	return &ItemItemStore{NewDataStoreWithTable(s.DB, "items_items")}
}

// ItemUnlockingRules returns an ItemUnlockingRuleStore
func (s *DataStore) ItemUnlockingRules() *ItemUnlockingRuleStore {
	return &ItemUnlockingRuleStore{NewDataStoreWithTable(s.DB, "item_unlocking_rules")}
}

// Languages returns a LanguageStore
func (s *DataStore) Languages() *LanguageStore {
	return &LanguageStore{NewDataStoreWithTable(s.DB, "languages")}
}

// LoginStates returns a LoginStateStore
func (s *DataStore) LoginStates() *LoginStateStore {
	return &LoginStateStore{NewDataStoreWithTable(s.DB, "login_states")}
}

// Platforms returns a PlatformStore
func (s *DataStore) Platforms() *PlatformStore {
	return &PlatformStore{NewDataStoreWithTable(s.DB, "platforms")}
}

// RefreshTokens returns a RefreshTokenStore
func (s *DataStore) RefreshTokens() *RefreshTokenStore {
	return &RefreshTokenStore{NewDataStoreWithTable(s.DB, "refresh_tokens")}
}

// Sessions returns a SessionStore
func (s *DataStore) Sessions() *SessionStore {
	return &SessionStore{NewDataStoreWithTable(s.DB, "sessions")}
}

// UserBatches returns a UserBatchStore
func (s *DataStore) UserBatches() *UserBatchStore {
	return &UserBatchStore{NewDataStoreWithTable(s.DB, "user_batches")}
}

// UserBatchPrefixes returns a UserBatchPrefixStore
func (s *DataStore) UserBatchPrefixes() *UserBatchPrefixStore {
	return &UserBatchPrefixStore{NewDataStoreWithTable(s.DB, "user_batch_prefixes")}
}

// NewID generates a positive random int64 to be used as id
// !!! To be safe, the insertion should be be retried if the id conflicts with an existing entry
func (s *DataStore) NewID() int64 {
	// gen a 63-bits number as we want unsigned number stored in a 64-bits signed DB attribute
	return rand.Int63()
}

// InTransaction executes the given function in a transaction and commits
func (s *DataStore) InTransaction(txFunc func(*DataStore) error) error {
	return s.inTransaction(func(db *DB) error {
		return txFunc(NewDataStoreWithTable(db, s.tableName))
	})
}

// WithForeignKeyChecksDisabled executes the given function with foreign keys checking disabled
// (wraps it up in a transaction if no transaction started)
func (s *DataStore) WithForeignKeyChecksDisabled(blockFunc func(*DataStore) error) error {
	return s.withForeignKeyChecksDisabled(func(db *DB) error {
		return blockFunc(NewDataStoreWithTable(db, s.tableName))
	})
}

// WithNamedLock wraps the given function in GET_LOCK/RELEASE_LOCK
func (s *DataStore) WithNamedLock(lockName string, timeout time.Duration, txFunc func(*DataStore) error) error {
	return s.withNamedLock(lockName, timeout, func(db *DB) error {
		return txFunc(NewDataStoreWithTable(db, s.tableName))
	})
}

// ByID returns a composable query for filtering by _table_.id
func (s *DataStore) ByID(id int64) *DB {
	if s.tableName == "" {
		panic("method ByID() called for abstract DataStore")
	}
	return s.Where(s.tableName+".id = ?", id)
}

// RetryOnDuplicatePrimaryKeyError will retry the given function on getting duplicate entry errors
// for primary keys
func (s *DataStore) RetryOnDuplicatePrimaryKeyError(f func(store *DataStore) error) error {
	return s.DB.retryOnDuplicatePrimaryKeyError(func(db *DB) error {
		return f(NewDataStore(db))
	})
}

// RetryOnDuplicateKeyError will retry the given function on getting duplicate entry errors
// for the given key
func (s *DataStore) RetryOnDuplicateKeyError(keyName, nameInError string, f func(store *DataStore) error) error {
	return s.DB.retryOnDuplicateKeyError(keyName, nameInError, func(db *DB) error {
		return f(NewDataStore(db))
	})
}

// InsertMap reads fields from the given map and inserts the values which have been set
// into the store's table
func (s *DataStore) InsertMap(dataMap map[string]interface{}) error {
	return s.DB.insertMaps(s.tableName, []map[string]interface{}{dataMap})
}

// InsertMaps reads fields from the given map and inserts the values set in the first row (so all the rows should have the same keys)
// into the store's table
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
func (s *DataStore) InsertOrUpdateMaps(dataMap []map[string]interface{}, updateColumns []string) error {
	return s.DB.insertOrUpdateMaps(s.tableName, dataMap, updateColumns)
}

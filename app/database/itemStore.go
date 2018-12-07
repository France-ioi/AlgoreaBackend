package database

import (
	"fmt"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/logging"
	t "github.com/France-ioi/AlgoreaBackend/app/types"
)

// ItemStore implements database operations on items
type ItemStore struct {
	*DataStore
}

type itemAccessDetails struct {
	ItemID        int64 `sql:"column:idItem"`
	FullAccess    bool  `sql:"column:fullAccess"`
	PartialAccess bool  `sql:"column:partialAccess"`
	GrayedAccess  bool  `sql:"column:grayedAccess"`
}

// Item matches the content the `items` table
type Item struct {
	ID            t.Int64  `db:"ID"`
	Type          t.String `db:"sType"`
	TeamsEditable bool     `db:"bTeamsEditable"` // use Go default in DB (to be fixed)
	NoScore       bool     `db:"bNoScore"`       // use Go default in DB (to be fixed)
	Version       int64    `db:"iVersion"`       // use Go default in DB (to be fixed)
}

// Create insert an Item row in the database and associted values in related tables if needed
func (s *ItemStore) Create(item *Item, languageID t.Int64, title t.String, parentID t.Int64, order t.Int64) (int64, error) {

	if !item.ID.Set { // set it here as it will be returned
		item.ID = *t.NewInt64(generateID())
	}
	itemID := item.ID

	return itemID.Value, s.db.inTransaction(func(db *DB) error {
		var err error
		dataStore := &DataStore{db} // transaction datastore

		if _, err = dataStore.Items().createRaw(item); err != nil {
			return err
		}
		if _, err = dataStore.GroupItems().createRaw(&GroupItem{ItemID: itemID}); err != nil {
			return err
		}
		if _, err = dataStore.ItemStrings().createRaw(&ItemString{ItemID: itemID, LanguageID: languageID, Title: title}); err != nil {
			return err
		}
		if _, err = dataStore.ItemItems().createRaw(&ItemItem{ChildItemID: itemID, Order: order}); err != nil {
			return err
		}
		return nil
	})
}

// createRaw insert a row in the transaction and returns the
func (s *ItemStore) createRaw(entry *Item) (int64, error) {
	if !entry.ID.Set {
		entry.ID = *t.NewInt64(generateID())
	}
	err := s.db.insert("items", entry)
	return entry.ID.Value, err
}

// IsValidHierarchy gets an ordered set of item ids and returns whether they forms a valid item hierarchy path from a root
func (s *ItemStore) IsValidHierarchy(ids []int64) (bool, error) {
	return false, nil
}

// ValidateUserAccess gets a set of item ids and returns whether the given user is authorized to see them all
func (s *ItemStore) ValidateUserAccess(user *auth.User, itemIDs []int64) (bool, error) {

	var accDets []itemAccessDetails
	db := s.GroupItems().MatchingUserAncestors(user).
		Select("idItem, MAX(bCachedFullAccess) AS fullAccess, MAX(bCachedPartialAccess) AS partialAccess, MAX(bCachedGrayedAccess) AS grayedAccess").
		Where("groups_items.idItem IN (?)", itemIDs).
		Group("idItem").Scan(&accDets)
	if db.Error != nil {
		return false, db.Error
	}

	if err := checkAccess(itemIDs, accDets); err != nil {
		logging.Logger.Infof("User access validation failed: %v", err)
		return false, nil
	}
	return true, nil
}

// checkAccess checks if the user has access to all items:
// - user has to have full access to all items
// OR
// - user has to have full access to all but last, and grayed access to that last item.
func checkAccess(itemIDs []int64, accDets []itemAccessDetails) error {
	for i, id := range itemIDs {
		last := (i == len(itemIDs)-1)
		if err := checkAccessForID(id, last, accDets); err != nil {
			return err
		}
	}
	return nil
}

func checkAccessForID(id int64, last bool, accDets []itemAccessDetails) error {
	for _, res := range accDets {
		if res.ItemID != id {
			continue
		}
		if res.FullAccess || res.PartialAccess {
			// OK, user has full access.
			return nil
		}
		if res.GrayedAccess && last {
			// OK, user has grayed access on the last item.
			return nil
		}
		return fmt.Errorf("not enough perm on item_id %d", id)
	}

	// no row matching this item_id
	return fmt.Errorf("not visible item_id %d", id)
}

package database

import (
  "github.com/France-ioi/AlgoreaBackend/app/auth"
  "github.com/France-ioi/AlgoreaBackend/app/logging"
  t "github.com/France-ioi/AlgoreaBackend/app/types"
)

// ItemStore implements database operations on items
type ItemStore struct {
  *DataStore
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

// GetList returns all items with the given ids
func (s *ItemStore) GetList(itemIDs []int64, dest interface{}) error {
  query := s.db.
    Table("items_strings").
    Where("idItem IN (?)", itemIDs)
  query.Scan(dest)

  errors := query.GetErrors()
  if len(errors) > 0 {
    return errors[0]
  }
  return nil
}

// IsValidHierarchy gets an ordered set of item ids and returns whether they forms a valid item hierarchy path from a root
func (s *ItemStore) IsValidHierarchy(ids []int64) (bool, error) {
  return false, nil
}

// ValidateUserAccess gets a set of item ids and returns whether the given user is authorized to see them all
func (s *ItemStore) ValidateUserAccess(user *auth.User, itemIDs []int64) (bool, error) {

  accessResult := []struct {
    ItemID       int64 `sql:"column:idItem"`
    FullAccess   bool  `sql:"column:fullAccess"`
    GrayedAccess bool  `sql:"column:grayedAccess"`
  }{}

  query := s.db.
    Table("groups_items").
    Select("idItem, MAX(bCachedFullAccess) AS fullAccess, MAX(bCachedGrayedAccess) AS grayedAccess").
    Joins("JOIN groups_ancestors ON groups_items.idGroup = groups_ancestors.idGroupAncestor").
    Where("groups_ancestors.idGroupChild = ?", user.SelfGroupID()).
    Where("groups_items.idItem IN (?)", itemIDs).
    Group("idItem")
  query.Scan(&accessResult)

  errors := query.GetErrors()
  if len(errors) > 0 {
    return false, errors[0]
  }

  // check for each id whether it has access
  // must have full or partial access, or for the last one, grayed access
  for i, id := range itemIDs {
    found := false
    last := i == len(itemIDs)-1 // find the last id in the hierarchy
    for _, result := range accessResult {
      if result.ItemID == id {
        found = true
        if result.FullAccess || (last && result.GrayedAccess) {
          // ok
        } else {
          logging.Logger.Infof("User access validation failed: not enough perm on item_id %d", id)
          return false, nil // not enough access
        }
      }
    }
    if !found {
      logging.Logger.Infof("User access validation failed: not visible item_id %d", id)
      return false, nil // no row matching this item_id
    }
  }

  return true, nil
}

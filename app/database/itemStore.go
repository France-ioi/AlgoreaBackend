package database

import (
  t "github.com/France-ioi/AlgoreaBackend/app/types"
)

// ItemStore implements database operations on items
type ItemStore struct {
  db *DB
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

  groupItemStore := &GroupItemStore{s.db}
  itemItemStore := &ItemItemStore{s.db}
  itemStringStore := &ItemStringStore{s.db}

  if !item.ID.Set { // set it here as it will be returned
    item.ID = *t.NewInt64(generateID())
  }
  itemID := item.ID

  return itemID.Value, s.db.inTransaction(func(tx Tx) error {
    var err error

    if _, err = s.createRaw(tx, item); err != nil {
      return err
    }
    if _, err = groupItemStore.createRaw(tx, &GroupItem{ItemID: itemID}); err != nil {
      return err
    }
    if _, err = itemStringStore.createRaw(tx, &ItemString{ItemID: itemID, LanguageID: languageID, Title: title}); err != nil {
      return err
    }
    if _, err = itemItemStore.createRaw(tx, &ItemItem{ChildItemID: itemID, Order: order}); err != nil {
      return err
    }
    return nil
  })
}

// createRaw insert a row in the transaction and returns the
func (s *ItemStore) createRaw(tx Tx, entry *Item) (int64, error) {
  if !entry.ID.Set {
    entry.ID = *t.NewInt64(generateID())
  }
  err := tx.insert("items", entry)
  return entry.ID.Value, err
}

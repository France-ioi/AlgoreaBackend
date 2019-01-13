package items

import "github.com/France-ioi/AlgoreaBackend/app/database"

// Item .
type Item struct {
	ItemID   int64   `json:"item_id"`
	Title    string  `json:"title,omitempty"`
	Order    int64   `json:"order,omitempty"`
	Children []*Item `json:"children,omitempty"`
}

func treeItemFromDB(dbIt *database.TreeItem) *Item {
	if dbIt == nil {
		return nil
	}
	return &Item{
		ItemID: dbIt.ID.Value,
		Title:  dbIt.Title.Value,
		Order:  dbIt.Order.Value,
	}
}

func (it *Item) fillChildren(chItt []*database.TreeItem) {
	for _, chIt := range chItt {
		chItem := treeItemFromDB(chIt)
		it.Children = append(it.Children, chItem)
	}
}

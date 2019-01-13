package items

import "github.com/France-ioi/AlgoreaBackend/app/database"

type Item struct {
	ItemID   int64   `json:"item_id"`
	Title    string  `json:"title,omitempty"`
	Order    int64   `json:"order,omitempty"`
	Children []*Item `json:"children,omitempty"`
}

func itemFromDB(dbIt *database.Item) *Item {
	if dbIt == nil {
		return nil
	}
	return &Item{
		ItemID: dbIt.ID.Value,
		Title:  dbIt.Title.Value,
		Order:  dbIt.Order.Value,
	}
}

func (it *Item) fillChildren(chItt []*database.Item) {
	for _, chIt := range chItt {
		chItem := itemFromDB(chIt)
		it.Children = append(it.Children, chItem)
	}
}

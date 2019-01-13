package items

import "github.com/France-ioi/AlgoreaBackend/app/database"

type Item struct {
	ItemID   int64   `json:"item_id"`
	Title    string  `json:"title,omitempty"`
	Order    int64   `json:"order,omitempty"`
	Children []*Item `json:"children,omitempty"`
}

func (it *Item) fillItemData(dbIt *database.Item) {
	if dbIt == nil {
		return
	}
	it.ItemID = dbIt.ID.Value
	it.Title = dbIt.Title.Value
}

func (it *Item) fillItemItemData(dbItIt *database.ItemItem) {
	if dbItIt == nil {
		return
	}
	it.Order = dbItIt.Order.Value
}

func (it *Item) fillChildren(chItt []*database.Item, chItIts []*database.ItemItem) {
	for _, chIt := range chItt {
		chItIt := findItemItem(chItIts, chIt.ID.Value)

		chItem := &Item{}
		chItem.fillItemData(chIt)
		chItem.fillItemItemData(chItIt)
		it.Children = append(it.Children, chItem)
	}
}

func findItemItem(itt []*database.ItemItem, id int64) *database.ItemItem {
	for _, it := range itt {
		if it.ChildItemID.Value == id {
			return it
		}
	}
	return nil
}

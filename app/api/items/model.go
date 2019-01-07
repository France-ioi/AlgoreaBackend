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
}

func (it *Item) fillItemStringData(dbItStr *database.ItemString) {
	if dbItStr == nil {
		return
	}
	it.Title = dbItStr.Title.Value
}

func (it *Item) fillItemItemData(dbItIt *database.ItemItem) {
	if dbItIt == nil {
		return
	}
	it.Order = dbItIt.Order.Value
}

func (it *Item) fillChildren(chItt []*database.Item, chItIts []*database.ItemItem, chItStrs []*database.ItemString) {
	for _, chIt := range chItt {
		chItIt := findItemItem(chItIts, chIt.ID.Value)
		chItStr := findItemString(chItStrs, chIt.ID.Value)

		chItem := &Item{}
		chItem.fillItemData(chIt)
		chItem.fillItemItemData(chItIt)
		chItem.fillItemStringData(chItStr)
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

func findItemString(itt []*database.ItemString, id int64) *database.ItemString {
	for _, it := range itt {
		if it.ItemID.Value == id {
			return it
		}
	}
	return nil
}

package database

import (
	"fmt"
	"github.com/France-ioi/AlgoreaBackend/app/types"
	"testing"
)

func TestCheckAccess(t *testing.T) {
	testCases := []struct {
		desc               string
		itemIDs            []int64
		itemAccessDetailss []itemAccessDetails
		err                error
	}{
		{
			desc:               "empty IDs",
			itemIDs:            nil,
			itemAccessDetailss: nil,
			err:                nil,
		},
		{
			desc:               "empty access results",
			itemIDs:            []int64{21, 22, 23},
			itemAccessDetailss: nil,
			err:                fmt.Errorf("not visible item_id 21"),
		},
		{
			desc:    "missing access result on one of the items",
			itemIDs: []int64{21, 22, 23},
			itemAccessDetailss: []itemAccessDetails{
				{ItemID: 21, FullAccess: true},
				{ItemID: 22, FullAccess: true},
			},
			err: fmt.Errorf("not visible item_id 23"),
		},
		{
			desc:    "no access on one of the items",
			itemIDs: []int64{21, 22, 23},
			itemAccessDetailss: []itemAccessDetails{
				{ItemID: 21, FullAccess: true},
				{ItemID: 22},
				{ItemID: 23, FullAccess: true},
			},
			err: fmt.Errorf("not enough perm on item_id 22"),
		},
		{
			desc:    "full access on all items",
			itemIDs: []int64{21, 22, 23},
			itemAccessDetailss: []itemAccessDetails{
				{ItemID: 21, FullAccess: true},
				{ItemID: 22, FullAccess: true},
				{ItemID: 23, FullAccess: true},
			},
			err: nil,
		},
		{
			desc:    "full access on all but last, last with greyed",
			itemIDs: []int64{21, 22, 23},
			itemAccessDetailss: []itemAccessDetails{
				{ItemID: 21, PartialAccess: true},
				{ItemID: 22, PartialAccess: true},
				{ItemID: 23, GrayedAccess: true},
			},
			err: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := checkAccess(tC.itemIDs, tC.itemAccessDetailss)
			if err != nil {
				if tC.err != nil {
					if want, got := tC.err.Error(), err.Error(); want != got {
						t.Fatalf("Expected error to be %v, got: %v", want, got)
					}
					return
				}
				t.Fatalf("Unexpected error: %v", err)
			}
			if tC.err != nil {
				t.Fatalf("Expected error %v", tC.err)
			}
		})
	}
}


func TestCheckHierarchy(t *testing.T) {
	var rootID int64 = 21
	root := itemAncestorDetails{
		ID: rootID,
		Type: *types.NewString(ItemTypeRoot),
		IdItemChild: 22,
	}

	var taskID int64 = 23
	task := itemAncestorDetails{
		ID: taskID,
		Type: *types.NewString(ItemTypeTask),
		IdItemChild: 0,
	}

	var categoryID int64 = 22
	category := itemAncestorDetails{
		ID: categoryID,
		Type: *types.NewString(ItemTypeCategory),
		IdItemChild: 23,
	}

	var chapterID int64 = 25
	chapter := itemAncestorDetails{
		ID: chapterID,
		Type: *types.NewString(ItemTypeChapter),
		IdItemChild: 0,
	}

	items := []itemAncestorDetails{
		root,
		task,
		category,
	}

	err := checkHierarchy(items)
	if err != nil {
		t.Errorf("checkHierarchy must return nil on items")
	}

	itemsIncorrect := []itemAncestorDetails{
		root,
		chapter,
		task,
		category,
	}

	err = checkHierarchy(itemsIncorrect)
	if err == nil {
		t.Errorf("checkHierarchy must return error on itemsIncorrect")
	}
}

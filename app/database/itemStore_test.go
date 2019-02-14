package database

import (
	"fmt"
	"testing"
)

func TestCheckAccess(t *testing.T) {
	testCases := []struct {
		desc              string
		itemIDs           []int64
		itemAccessDetails []ItemAccessDetailsWithID
		err               error
	}{
		{
			desc:              "empty IDs",
			itemIDs:           nil,
			itemAccessDetails: nil,
			err:               nil,
		},
		{
			desc:              "empty access results",
			itemIDs:           []int64{21, 22, 23},
			itemAccessDetails: nil,
			err:               fmt.Errorf("not visible item_id 21"),
		},
		{
			desc:    "missing access result on one of the items",
			itemIDs: []int64{21, 22, 23},
			itemAccessDetails: []ItemAccessDetailsWithID{
				{ItemID: 21, ItemAccessDetails: ItemAccessDetails{FullAccess: true}},
				{ItemID: 22, ItemAccessDetails: ItemAccessDetails{FullAccess: true}},
			},
			err: fmt.Errorf("not visible item_id 23"),
		},
		{
			desc:    "no access on one of the items",
			itemIDs: []int64{21, 22, 23},
			itemAccessDetails: []ItemAccessDetailsWithID{
				{ItemID: 21, ItemAccessDetails: ItemAccessDetails{FullAccess: true}},
				{ItemID: 22},
				{ItemID: 23, ItemAccessDetails: ItemAccessDetails{FullAccess: true}},
			},
			err: fmt.Errorf("not enough perm on item_id 22"),
		},
		{
			desc:    "full access on all items",
			itemIDs: []int64{21, 22, 23},
			itemAccessDetails: []ItemAccessDetailsWithID{
				{ItemID: 21, ItemAccessDetails: ItemAccessDetails{FullAccess: true}},
				{ItemID: 22, ItemAccessDetails: ItemAccessDetails{FullAccess: true}},
				{ItemID: 23, ItemAccessDetails: ItemAccessDetails{FullAccess: true}},
			},
			err: nil,
		},
		{
			desc:    "full access on all but last, last with greyed",
			itemIDs: []int64{21, 22, 23},
			itemAccessDetails: []ItemAccessDetailsWithID{
				{ItemID: 21, ItemAccessDetails: ItemAccessDetails{PartialAccess: true}},
				{ItemID: 22, ItemAccessDetails: ItemAccessDetails{PartialAccess: true}},
				{ItemID: 23, ItemAccessDetails: ItemAccessDetails{GrayedAccess: true}},
			},
			err: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := checkAccess(tC.itemIDs, tC.itemAccessDetails)
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

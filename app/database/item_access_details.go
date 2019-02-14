package database

// ItemAccessDetails represents access rights for an item
type ItemAccessDetails struct {
	// MAX(groups_items.bCachedFullAccess)
	FullAccess bool `sql:"column:fullAccess" json:"full_access"`
	// MAX(groups_items.bCachedPartialAccess)
	PartialAccess bool `sql:"column:partialAccess" json:"partial_access"`
	// MAX(groups_items.bCachedGrayAccess)
	GrayedAccess bool `sql:"column:grayedAccess" json:"grayed_access"`
	// MAX(groups_items.bCachedAccessSolutions)
	AccessSolutions bool `sql:"column:accessSolutions" json:"access_solutions"`
}

// ItemAccessDetailsWithID represents access rights for an item + ItemID
type ItemAccessDetailsWithID struct {
	ItemID int64 `sql:"column:idItem"`
	ItemAccessDetails
}

// IsGrayed returns true when GrayedAccess is on, but FullAccess and PartialAccess are off
func (accessDetails *ItemAccessDetails) IsGrayed() bool {
	return !accessDetails.FullAccess && !accessDetails.PartialAccess && accessDetails.GrayedAccess
}

// IsForbidden returns true when FullAccess, PartialAccess, GrayedAccess are off
func (accessDetails *ItemAccessDetails) IsForbidden() bool {
	return !accessDetails.FullAccess && !accessDetails.PartialAccess && !accessDetails.GrayedAccess
}

package database

// ItemAccessDetails represents access rights for an item
type ItemAccessDetails struct {
	// MIN(groups_items.sCachedFullAccessDate) <= NOW()
	FullAccess bool `sql:"column:fullAccess" json:"full_access"`
	// MIN(groups_items.sCachedPartialAccessDate) <= NOW()
	PartialAccess bool `sql:"column:partialAccess" json:"partial_access"`
	// MIN(groups_items.sCachedGrayAccessDate) <= NOW()
	GrayedAccess bool `sql:"column:grayedAccess" json:"grayed_access"`
	// MIN(groups_items.sCachedAccessSolutionsDate) <= NOW()
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

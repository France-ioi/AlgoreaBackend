package database

// ItemAccessDetails represents access rights for an item
type ItemAccessDetails struct {
	// MIN(groups_items.cached_full_access_date) <= NOW()
	FullAccess bool `json:"full_access"`
	// MIN(groups_items.cached_partial_access_date) <= NOW()
	PartialAccess bool `json:"partial_access"`
	// MIN(groups_items.cached_grayed_access_date) <= NOW()
	GrayedAccess bool `json:"grayed_access"`
	// MIN(groups_items.cached_access_solutions_date) <= NOW()
	AccessSolutions bool `json:"access_solutions"`
}

// ItemAccessDetailsWithID represents access rights for an item + ItemID
type ItemAccessDetailsWithID struct {
	ItemID int64
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

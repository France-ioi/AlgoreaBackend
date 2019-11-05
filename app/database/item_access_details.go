package database

const canViewInfo = "info"

// ItemAccessDetails represents access rights for an item
type ItemAccessDetails struct {
	// MAX(permissions_generated.can_view_generated_value) converted back into the string representation
	CanView string `json:"can_view"`
}

// ItemAccessDetailsWithID represents access rights for an item + ItemID
type ItemAccessDetailsWithID struct {
	ItemID int64
	ItemAccessDetails
}

// IsGrayed returns true when can_view_generated = 'info'
func (accessDetails *ItemAccessDetails) IsGrayed() bool {
	return accessDetails.CanView == canViewInfo
}

// IsForbidden returns true when can_view_generated = 'none'
func (accessDetails *ItemAccessDetails) IsForbidden() bool {
	return accessDetails.CanView == "none"
}

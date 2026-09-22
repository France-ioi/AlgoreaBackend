package payloads

// GroupResultsToken represents data inside a group-results JWS token.
// This token authorizes obtaining a group's progress-with-answers export.
// swagger:model GroupResultsToken
type GroupResultsToken struct {
	// Format dd-mm-yyyy (auto-added by token.Generate)
	// required:true
	Date string `json:"date" validate:"dmy-date"`
	// The authenticated user's ID
	// required:true
	UserID string `json:"user_id"`
	// The group whose results may be exported
	// required:true
	GroupID string `json:"group_id"`
	// Validated, de-duplicated parent item IDs (may be empty)
	// required:true
	ItemIDs []string `json:"item_ids"`
	// Expiry date in the number of seconds since 01/01/1970 UTC.
	// required:true
	Exp int64 `json:"exp"`
}

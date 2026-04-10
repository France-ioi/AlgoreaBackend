package payloads

// PermissionsToken represents data inside a permissions JWS token.
// This token allows external services to verify a user's permissions on an item.
// swagger:model PermissionsToken
type PermissionsToken struct {
	// Format dd-mm-yyyy (auto-added by token.Generate)
	// required:true
	Date string `json:"date" validate:"dmy-date"`
	// The authenticated user's ID
	// required:true
	UserID string `json:"user_id"`
	// The item ID
	// required:true
	ItemID string `json:"item_id"`
	// required:true
	// enum: none,info,content,content_with_descendants,solution
	CanView string `json:"can_view"`
	// required:true
	// enum: none,enter,content,content_with_descendants,solution,solution_with_grant
	CanGrantView string `json:"can_grant_view"`
	// required:true
	// enum: none,result,answer,answer_with_grant
	CanWatch string `json:"can_watch"`
	// required:true
	// enum: none,children,all,all_with_grant
	CanEdit string `json:"can_edit"`
	// required:true
	IsOwner bool `json:"is_owner"`
	// Expiry date in the number of seconds since 01/01/1970 UTC.
	// required:true
	Exp int64 `json:"exp"`
}

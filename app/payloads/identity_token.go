package payloads

// IdentityToken represents data inside an identity token.
// This token allows external services to verify a user's identity
// without exposing session information.
// swagger:model IdentityToken
type IdentityToken struct {
	// Format dd-mm-yyyy (auto-added by token.Generate)
	// required:true
	Date string `json:"date" validate:"dmy-date"`
	// The authenticated user's ID
	// required:true
	UserID string `json:"user_id"`
	// Expiry date in seconds since 01/01/1970 UTC
	// required:true
	Exp int64 `json:"exp"`
}

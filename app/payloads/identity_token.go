package payloads

// IdentityToken represents data inside an identity JWS token.
// This token allows external services to verify a user's identity.
// swagger:model IdentityToken
type IdentityToken struct {
	// Format dd-mm-yyyy (auto-added by token.Generate)
	// required:true
	Date string `json:"date" validate:"dmy-date"`
	// The authenticated user's ID
	// required:true
	UserID string `json:"user_id"`
	// Whether the user is a temporary user
	// required:true
	IsTempUser bool `json:"is_temp_user"`
	// Expiry date in the number of seconds since 01/01/1970 UTC.
	// required:true
	Exp int64 `json:"exp"`
}

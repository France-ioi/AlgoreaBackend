package doc

// swagger:model userCreateTmpResponse
type userCreateTmpResponse struct {
	// description
	// swagger:allOf
	CreatedResponse
	// required:true
	Data struct {
		// required:true
		AccessToken string `json:"access_token"`
		// Number of seconds until the token's expiration
		// (when received by the UI, must be converted to actual time)
		// required:true
		ExpiresIn int32 `json:"expires_in"`
	} `json:"data"`
}

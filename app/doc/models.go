package doc

import "time"

// swagger:model createdResponse
type CreatedResponse struct {
	// "created"
	// enum: created
	// required: true
	Message string `json:"message"`
	// true
	// required: true
	Success bool `json:"success"`
}

// swagger:model userCreateTmpResponse
type userCreateTmpResponse struct {
	// description
	// swagger:allOf
	CreatedResponse
	// required:true
	Data struct {
		// Only if the cookie is not enabled
		AccessToken string `json:"access_token"`
		// Number of seconds until the token's expiration
		// (when received by the UI, must be converted to actual time)
		// required:true
		ExpiresIn int32 `json:"expires_in"`
	} `json:"data"`
}

// swagger:model groupsMembershipHistoryResponseRow
type groupsMembershipHistoryResponseRow struct {
	// `group_membership_changes.at`
	// required: true
	At time.Time `json:"at"`
	// `group_membership_changes.action`
	// required: true
	// enum: invitation_created,join_request_created,invitation_accepted,join_request_accepted,invitation_refused,joined_by_badge,joined_by_code,join_request_refused,join_request_withdrawn,invitation_withdrawn,removed,left,expired
	Action string `json:"action"`

	// required: true
	Group struct {
		// `groups.id`
		// required: true
		ID int64 `json:"id"`
		// required: true
		Name string `json:"name"`
		// required: true
		// enum: Class,Team,Club,Friends,Other,Session,Base
		Type string `json:"type"`
	} `json:"group"`
}

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
		// required:true
		AccessToken string `json:"access_token"`
		// Number of seconds until the token's expiration
		// (when received by the UI, must be converted to actual time)
		// required:true
		ExpiresIn int32 `json:"expires_in"`
	} `json:"data"`
}

// swagger:model groupsMembershipHistoryResponseRow
type groupsMembershipHistoryResponseRow struct {
	// `groups_groups.id`
	// required: true
	ID int64 `json:"id"`
	// `groups_groups.status_changed_at`
	// required: true
	StatusChangedAt time.Time `json:"status_changed_at"`
	// `groups_groups.type`
	// required: true
	// enum: invitationSent,requestSent,invitationAccepted,requestAccepted,invitationRefused,joinedByCode,requestRefused,removed,left
	Type string `json:"type"`

	// required: true
	Group struct {
		// required: true
		Name string `json:"name"`
		// required: true
		// enum: Class,Team,Club,Friends,Other
		Type string `json:"type"`
	} `json:"group"`
}

// swagger:model invitationsViewResponseRow
type invitationsViewResponseRow struct {
	// `groups_groups.id`
	// required: true
	ID int64 `json:"id"`
	// `groups_groups.status_changed_at`
	// required: true
	StatusChangedAt time.Time `json:"status_changed_at"`
	// `groups_groups.type`
	// required: true
	// enum: invitationSent,requestSent,requestRefused
	Type string `json:"type"`

	// the user that invited (Nullable: only for invitations)
	// required: true
	InvitingUser *struct {
		// `users.id`
		// required: true
		ID int64 `json:"id"`
		// required: true
		Login string `json:"login"`
		// Nullable
		// required: true
		FirstName string `json:"first_name"`
		// Nullable
		// required: true
		LastName string `json:"last_name"`
	} `json:"inviting_user"`

	// required: true
	Group struct {
		// `groups.id`
		// required: true
		ID int64 `json:"id"`
		// required: true
		Name string `json:"name"`
		// Nullable
		// required: true
		Description *string `json:"description"`
		// required: true
		// enum: Class,Team,Club,Friends,Other
		Type string `json:"type"`
	} `json:"group"`
}

// swagger:model membershipsViewResponseRow
type membershipsViewResponseRow struct {
	// `groups_groups.id`
	// required: true
	ID int64 `json:"id"`
	// `groups_groups.status_changed_at`
	// required: true
	StatusChangedAt time.Time `json:"status_changed_at"`
	// `groups_groups.type`
	// required: true
	// enum: invitationAccepted,requestAccepted,joinedByCode,direct
	Type string `json:"type"`

	// required: true
	Group struct {
		// `groups.id`
		// required: true
		ID int64 `json:"id"`
		// required: true
		Name string `json:"name"`
		// Nullable
		// required: true
		Description *string `json:"description"`
		// required: true
		// enum: Class,Team,Club,Friends,Other,Base
		Type string `json:"type"`
	} `json:"group"`
}

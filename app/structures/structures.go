// Package structures provides item-related structs.
package structures

import (
	"time"
)

// ItemPermissions represents all the permissions that a group can have on an item
type ItemPermissions struct {
	// required: true
	// enum: none,info,content,content_with_descendants,solution
	CanView string `json:"can_view"`
	// required: true
	// enum: none,enter,content,content_with_descendants,solution,solution_with_grant
	CanGrantView string `json:"can_grant_view"`
	// required: true
	// enum: none,result,answer,answer_with_grant
	CanWatch string `json:"can_watch"`
	// required: true
	// enum: none,children,all,all_with_grant
	CanEdit string `json:"can_edit"`
	// required: true
	IsOwner bool `json:"is_owner"`
}

// ItemString represents a title with a related language tag for an item
type ItemString struct {
	// [Nullable] title (from `items_strings`) in the user’s default language or (if not available) default language of the item
	// required: true
	Title *string `json:"title"`
	// language_tag (from `items_strings`) to which the title is related
	// required: true
	LanguageTag string `json:"language_tag"`
}

// ItemCommonFields represents item fields common for different services (id, type, string, permissions)
type ItemCommonFields struct {
	// required: true
	ID int64 `json:"id,string"`
	// required: true
	// enum: Chapter,Task,Skill
	Type string `json:"type"`

	// required: true
	String ItemString `json:"string"`

	// required: true
	Permissions ItemPermissions `json:"permissions"`
}

// ItemResult represents item result info
type ItemResult struct {
	// required:true
	AttemptID int64 `json:"attempt_id,string"`

	// required:true
	ScoreComputed float32 `json:"score_computed"`
	// required:true
	Validated bool `json:"validated"`
	// Nullable
	// required:true
	StartedAt *time.Time `json:"started_at"`
	// required:true
	LatestActivityAt time.Time `json:"latest_activity_at"`
	// Nullable
	// required:true
	EndedAt *time.Time `json:"ended_at"`
	// required:true
	// attempts.allows_submissions_until
	AttemptAllowsSubmissionsUntil time.Time `json:"attempt_allows_submissions_until"`
}

// GroupShortInfo contains group id & name
type GroupShortInfo struct {
	// group's `id`
	// required:true
	ID int64 `json:"id,string"`
	// required:true
	Name string `json:"name"`
}

// UserPersonalInfo contains first_name and last_name
type UserPersonalInfo struct {
	// Nullable
	FirstName *string `json:"first_name"`
	// Nullable
	LastName *string `json:"last_name"`
}

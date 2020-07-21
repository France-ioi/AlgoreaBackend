package items

import (
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// RawCommonItemFields represents DB data fields that are common for itemView & itemChildrenView
type RawCommonItemFields struct {
	// items
	ID                     int64
	Type                   string
	DisplayDetailsInParent bool
	ValidationType         string
	EntryParticipantType   string
	EnteringTimeMin        database.Time
	EnteringTimeMax        database.Time
	AllowsMultipleAttempts bool
	Duration               *string
	NoScore                bool
	DefaultLanguageTag     string
	RequiresExplicitEntry  bool

	*database.RawGeneratedPermissionFields
}

func (raw *RawCommonItemFields) asItemCommonFields(permissionGrantedStore *database.PermissionGrantedStore) *commonItemFields {
	return &commonItemFields{
		ID:                     raw.ID,
		Type:                   raw.Type,
		DisplayDetailsInParent: raw.DisplayDetailsInParent,
		ValidationType:         raw.ValidationType,
		RequiresExplicitEntry:  raw.RequiresExplicitEntry,
		AllowsMultipleAttempts: raw.AllowsMultipleAttempts,
		EntryParticipantType:   raw.EntryParticipantType,
		Duration:               raw.Duration,
		NoScore:                raw.NoScore,
		DefaultLanguageTag:     raw.DefaultLanguageTag,
		Permissions:            *raw.AsItemPermissions(permissionGrantedStore),
	}
}

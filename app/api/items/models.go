package items

import "github.com/France-ioi/AlgoreaBackend/app/types"

// NewItemRequest is the expected input for new created item
type NewItemRequest struct {
	ID      types.OptionalInt64  `json:"id"`
	Type    types.RequiredString `json:"type"`
	Strings []LocalizedTitle     `json:"strings"`
	Parents []ParentRef          `json:"parents"`
}

type NewItemResponse struct {
	ItemID int64 `json:"ID"`
}

type LocalizedTitle struct {
	LanguageID types.RequiredInt64  `json:"language_id"`
	Title      types.RequiredString `json:"title"`
}

type ParentRef struct {
	ID    types.RequiredInt64 `json:"id"`
	Order types.RequiredInt64 `json:"order"`
}

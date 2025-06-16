package items

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

type parentTitle struct {
	// required: true
	Title *string `json:"title"`
	// required: true
	LanguageTag string `json:"language_tag"`
}

// swagger:model
type itemTimeLimitedAdminList struct {
	// required: true
	ItemID int64 `json:"id,string"`
	// required: true
	Title *string `gorm:"column:title_translation" json:"title"`
	// required: true
	LanguageTag string `gorm:"column:title_language_tag" json:"language_tag"`
	// required: true
	// enum: User,Team
	EntryParticipantType string `json:"entry_participant_type"`
	// required: true
	AllowsMultipleAttempts bool `json:"allows_multiple_attempts"`
	// required: true
	Parents []parentTitle `json:"parents"`
}

// swagger:operation GET /items/time-limited/administered items itemTimeLimitedAdminList
//
//	---
//	summary: List administered time-limited items
//	description: Get time-limited items that the user has administration rights on.
//
//
//							 For all explicit-entry items that are time-limited items (with duration <> NULL) the user can administer
//							 (has `can_view` >= 'content', `can_grant_view` >= 'enter', and `can_watch` >= 'result'),
//							 returns the item info including items' parents.
//							 Only parents visible to the user are listed.
//
//
//							 Each title is returned in the user's default language if exists,
//							 otherwise the item's default language is used.
//	parameters:
//		- name: from.id
//			description: Start the page from the item next to the item with `id`=`{from.id}`
//			in: query
//			type: integer
//			format: int64
//		- name: sort
//			in: query
//			default: [title,id]
//			type: array
//			items:
//				type: string
//				enum: [title,-title,id,-id]
//		- name: limit
//			description: Display the first N items
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//			description: OK. Success response with items info
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/itemTimeLimitedAdminList"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getAdministeredList(w http.ResponseWriter, r *http.Request) error {
	user := srv.GetUser(r)
	store := srv.GetStore(r)

	var rows []itemTimeLimitedAdminList
	query := store.Items().Select(`
			items.id AS item_id,
			items.allows_multiple_attempts,
			items.entry_participant_type,
			COALESCE(user_strings.title, default_strings.title) AS title_translation,
			COALESCE(user_strings.language_tag, default_strings.language_tag) AS title_language_tag`).
		JoinsPermissionsForGroupToItemsWherePermissionAtLeast(user.GroupID, "view", "content").
		WherePermissionIsAtLeast("grant_view", "enter").
		WherePermissionIsAtLeast("watch", "result").
		JoinsUserAndDefaultItemStrings(user).
		Where("items.duration IS NOT NULL").
		Where("items.requires_explicit_entry")

	query, err := service.ApplySortingAndPaging(
		r, query,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"title": {ColumnName: "IFNULL(COALESCE(user_strings.title, default_strings.title), '')"},
				"id":    {ColumnName: "items.id"},
			},
			DefaultRules: "title,id",
			TieBreakers:  service.SortingAndPagingTieBreakers{"id": service.FieldTypeInt64},
		})
	service.MustNotBeError(err)
	query = service.NewQueryLimiter().Apply(r, query)

	service.MustNotBeError(query.Scan(&rows).Error())

	if len(rows) > 0 {
		itemIDs := make([]int64, len(rows))
		for index := range rows {
			itemIDs[index] = rows[index].ItemID
		}
		var parents []struct {
			ChildID           int64
			ParentTitle       *string
			ParentLanguageTag string
		}
		service.MustNotBeError(store.Items().
			Joins("JOIN items_items ON items_items.parent_item_id = items.id AND items_items.child_item_id IN (?)", itemIDs).
			WhereItemsAreVisible(user.GroupID).
			JoinsUserAndDefaultItemStrings(user).
			Order("COALESCE(user_strings.title, default_strings.title)").
			Select(`
				items_items.child_item_id as child_id,
				COALESCE(user_strings.title, default_strings.title) AS parent_title,
				COALESCE(user_strings.language_tag, default_strings.language_tag) AS parent_language_tag`).
			Scan(&parents).Error())

		parentTitlesMap := make(map[int64][]parentTitle, len(rows))
		for index := range parents {
			if _, ok := parentTitlesMap[parents[index].ChildID]; !ok {
				parentTitlesMap[parents[index].ChildID] = make([]parentTitle, 0, 1)
			}
			parentTitlesMap[parents[index].ChildID] = append(parentTitlesMap[parents[index].ChildID], parentTitle{
				Title:       parents[index].ParentTitle,
				LanguageTag: parents[index].ParentLanguageTag,
			})
		}
		for index := range rows {
			rows[index].Parents = parentTitlesMap[rows[index].ItemID]
			if rows[index].Parents == nil {
				rows[index].Parents = make([]parentTitle, 0)
			}
		}
	}

	render.Respond(w, r, rows)
	return nil
}

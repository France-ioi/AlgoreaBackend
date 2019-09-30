package contests

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type parentTitle struct {
	// Nullable
	// required: true
	Title *string `json:"title"`
	// Nullable
	// required: true
	LanguageID *int64 `json:"language_id,string"`
}

// swagger:model
type contestAdminListRow struct {
	// required: true
	ItemID int64 `json:"id,string"`
	// Nullable
	// required: true
	Title *string `gorm:"column:title_translation" json:"title"`
	// Nullable
	// required: true
	LanguageID *int64 `gorm:"column:title_language_id" json:"language_id,string"`
	// required: true
	TeamOnlyContest bool `json:"team_only_contest"`
	// required: true
	Parents []parentTitle `json:"parents"`
}

// swagger:operation GET /contests/administered contests groups contestAdminList
// ---
// summary: Get the contests that the user has administration rights on
// description: >
//                For all items that are timed contests and for that the user is a contest admin
//                (has `solutions` or `full` access), returns item info (`id`, `title`, `team_only_contest`, parents' `title`-s).
//                Only parents visible by the user (`full`, `partial`, `gray`) are listed.
//
//
//                Each title is returned in the user's default language if exists,
//                otherwise the item's default language is used.
// parameters:
// - name: from.title
//   description: Start the page from the contest next to the contest with `title` = `from.title` and `id` = `from.id`
//                (`from.id` is required when `from.title` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the contest next to the contest with `title` = `from.title` and `id`=`from.id`
//                (`from.title` is required when from.id is present)
//   in: query
//   type: integer
// - name: sort
//   in: query
//   default: [title,id]
//   type: array
//   items:
//     type: string
//     enum: [title,-title,id,-id]
// - name: limit
//   description: Display the first N contests
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. Success response with contests info
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/contestAdminListRow"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getAdministeredList(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	var rows []contestAdminListRow
	query := srv.Store.Items().Select(`
			items.id AS item_id,
			items.has_attempts AS team_only_contest,
			COALESCE(MAX(user_strings.title), MAX(default_strings.title)) AS title_translation,
			COALESCE(MAX(user_strings.language_id), MAX(default_strings.language_id)) AS title_language_id`).
		Joins("JOIN groups_items ON groups_items.item_id = items.id").
		Joins("JOIN groups_ancestors ON groups_ancestors.ancestor_group_id = groups_items.group_id").
		JoinsUserAndDefaultItemStrings(user).
		Where("groups_items.cached_full_access_since <= NOW() OR groups_items.cached_solutions_access_since <= NOW()").
		Where("groups_ancestors.child_group_id = ?", user.SelfGroupID).
		Where("items.duration IS NOT NULL").
		Group("items.id")

	query, apiError := service.ApplySortingAndPaging(r, query, map[string]*service.FieldSortingParams{
		"title": {
			ColumnName:            "IFNULL(COALESCE(user_strings.title, default_strings.title), '')",
			ColumnNameForOrdering: "IFNULL(COALESCE(MAX(user_strings.title), MAX(default_strings.title)), '')",
			FieldType:             "string",
		},
		"id": {ColumnName: "items.id", FieldType: "int64"},
	}, "title,id")
	if apiError != service.NoError {
		return apiError
	}
	query = service.NewQueryLimiter().Apply(r, query)

	service.MustNotBeError(query.Scan(&rows).Error())

	if len(rows) > 0 {
		itemIDs := make([]int64, len(rows))
		for index := range rows {
			itemIDs[index] = rows[index].ItemID
		}
		var parents []struct {
			ChildID          int64
			ParentTitle      *string
			ParentLanguageID *int64
		}
		service.MustNotBeError(srv.Store.Items().
			Joins("JOIN items_items ON items_items.parent_item_id = items.id AND items_items.child_item_id IN (?)", itemIDs).
			Joins(`
				JOIN groups_items AS parent_groups_items
					ON parent_groups_items.item_id = items.id AND (
						parent_groups_items.cached_full_access_since <= NOW() OR
						parent_groups_items.cached_partial_access_since <= NOW() OR
						parent_groups_items.cached_grayed_access_since <= NOW()
				)`).
			Joins(`
				JOIN groups_ancestors AS parent_groups_ancestors
					ON parent_groups_ancestors.ancestor_group_id = parent_groups_items.group_id AND
						parent_groups_ancestors.child_group_id = ?`, user.SelfGroupID).
			JoinsUserAndDefaultItemStrings(user).
			Group("items_items.parent_item_id, items_items.child_item_id").
			Order("COALESCE(MAX(user_strings.title), MAX(default_strings.title))").
			Select(`
				items_items.child_item_id as child_id,
				COALESCE(MAX(user_strings.title), MAX(default_strings.title)) AS parent_title,
				COALESCE(MAX(user_strings.language_id), MAX(default_strings.language_id)) AS parent_language_id`).
			Scan(&parents).Error())

		parentTitlesMap := make(map[int64][]parentTitle, len(rows))
		for index := range parents {
			if _, ok := parentTitlesMap[parents[index].ChildID]; !ok {
				parentTitlesMap[parents[index].ChildID] = make([]parentTitle, 0, 1)
			}
			parentTitlesMap[parents[index].ChildID] =
				append(parentTitlesMap[parents[index].ChildID], parentTitle{
					Title:      parents[index].ParentTitle,
					LanguageID: parents[index].ParentLanguageID,
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
	return service.NoError
}

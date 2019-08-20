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
	ItemID int64 `gorm:"column:idItem" json:"id,string"`
	// Nullable
	// required: true
	Title *string `gorm:"column:sTitleTranslation" json:"title"`
	// Nullable
	// required: true
	LanguageID *int64 `gorm:"column:idTitleLanguage" json:"language_id,string"`
	// required: true
	TeamOnlyContest bool `gorm:"column:bTeamOnlyContest" json:"team_only_contest"`
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
//   description: Start the page from the contest next to the contest with `title` = `from.title` and `ID` = `from.id`
//                (`from.id` is required when `from.title` is present)
//   in: query
//   type: string
// - name: from.id
//   description: Start the page from the contest next to the contest with `title` = `from.title` and `ID`=`from.id`
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
			items.ID AS idItem,
			items.bHasAttempts AS bTeamOnlyContest,
			COALESCE(user_strings.sTitle, default_strings.sTitle) AS sTitleTranslation,
			COALESCE(user_strings.idLanguage, default_strings.idLanguage) AS idTitleLanguage`).
		Joins("JOIN groups_items ON groups_items.idItem = items.ID").
		Joins("JOIN groups_ancestors ON groups_ancestors.idGroupAncestor = groups_items.idGroup").
		JoinsUserAndDefaultItemStrings(user).
		Where("groups_items.sCachedFullAccessDate <= NOW() OR groups_items.sCachedAccessSolutionsDate <= NOW()").
		Where("groups_ancestors.idGroupChild = ?", user.SelfGroupID).
		Where("items.sDuration IS NOT NULL").
		Group("items.ID")

	query, apiError := service.ApplySortingAndPaging(r, query, map[string]*service.FieldSortingParams{
		"title": {ColumnName: "IFNULL(COALESCE(user_strings.sTitle, default_strings.sTitle), '')", FieldType: "string"},
		"id":    {ColumnName: "items.ID", FieldType: "int64"},
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
			ChildID          int64   `gorm:"column:idChild"`
			ParentTitle      *string `gorm:"column:sTitleParent"`
			ParentLanguageID *int64  `gorm:"column:idLanguageParent"`
		}
		service.MustNotBeError(srv.Store.Items().
			Joins("JOIN items_items ON items_items.idItemParent = items.ID AND items_items.idItemChild IN (?)", itemIDs).
			Joins(`
				JOIN groups_items AS parent_groups_items
					ON parent_groups_items.idItem = items.ID AND (
						parent_groups_items.sCachedFullAccessDate <= NOW() OR
						parent_groups_items.sCachedPartialAccessDate <= NOW() OR
						parent_groups_items.sCachedGrayedAccessDate <= NOW()
				)`).
			Joins(`
				JOIN groups_ancestors AS parent_groups_ancestors
					ON parent_groups_ancestors.idGroupAncestor = parent_groups_items.idGroup AND
						parent_groups_ancestors.idGroupChild = ?`, user.SelfGroupID).
			JoinsUserAndDefaultItemStrings(user).
			Group("items_items.idItemParent, items_items.idItemChild").
			Order("COALESCE(user_strings.sTitle, default_strings.sTitle)").
			Select(`
				items_items.idItemChild as idChild,
				COALESCE(user_strings.sTitle, default_strings.sTitle) AS sTitleParent,
				COALESCE(user_strings.idLanguage, default_strings.idLanguage) AS idLanguageParent`).
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

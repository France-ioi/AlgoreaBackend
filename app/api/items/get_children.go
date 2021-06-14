package items

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/structures"
)

type listItemStringNotInfo struct {
	// Nullable; only if `can_view` >= 'content'
	Subtitle *string `json:"subtitle"`
}

type listItemString struct {
	// required: true
	LanguageTag string `json:"language_tag"`
	// Nullable
	// required: true
	Title *string `json:"title"`

	*listItemStringNotInfo
}

// swagger:model childItem
type childItem struct {
	*commonItemFields

	// required: true
	String listItemString `json:"string"`

	// `items_items.order`
	// required: true
	Order int32 `json:"order"`
	// `items_items.category`
	// required: true
	// enum: Undefined,Discovery,Application,Validation,Challenge
	Category string `json:"category"`
	// `items_items.score_weight`
	// required: true
	ScoreWeight int8 `json:"score_weight"`
	// `items_items.content_view_propagation`
	// required: true
	// enum: none,as_info,as_content
	ContentViewPropagation string `json:"content_view_propagation"`
	// `items_items.upper_view_levels_propagation`
	// required: true
	// enum: use_content_view_propagation,as_content_with_descendants,as_is
	UpperViewLevelsPropagation string `json:"upper_view_levels_propagation"`
	// `items_items.grant_view_propagation`
	// required: true
	GrantViewPropagation bool `json:"grant_view_propagation"`
	// `items_items.watch_propagation`
	// required: true
	WatchPropagation bool `json:"watch_propagation"`
	// `items_items.edit_propagation`
	// required: true
	EditPropagation bool `json:"edit_propagation"`

	// max among all attempts of the user (or of the team given in `{as_team_id}`)
	// required: true
	BestScore float32 `json:"best_score"`
	// required:true
	Results []structures.ItemResult `json:"results"`

	WatchedGroup *itemWatchedGroupStat `json:"watched_group,omitempty"`

	// whether solving this item grants access to some items (visible or not)
	// required: true
	GrantsAccessToItems bool `json:"grants_access_to_items"`
}

// RawListItem contains raw fields common for itemChildrenView & itemParentsView
type RawListItem struct {
	*RawCommonItemFields

	// from items_strings: in the user’s default language or (if not available) default language of the item
	StringLanguageTag string  `sql:"column:language_tag"`
	StringTitle       *string `sql:"column:title"`
	StringSubtitle    *string `sql:"column:subtitle"`

	// items_items
	Category string
	Order    int32 `sql:"column:child_order"`

	// max from results of the current participant
	BestScore float32

	*RawItemResultFields
	*RawWatchedGroupStatFields
}

type rawListChildItem struct {
	*RawListItem

	// items_items
	ScoreWeight                int8
	ContentViewPropagation     string
	UpperViewLevelsPropagation string
	GrantViewPropagation       bool
	WatchPropagation           bool
	EditPropagation            bool

	// item_dependencies
	GrantsAccessToItems bool
}

// swagger:operation GET /items/{item_id}/children items itemChildrenView
// ---
// summary: Get item children
// description: Lists children of the specified item
//              and the current user's (or the team's given in `as_team_id`) interactions with them
//              (from tables `items`, `items_items`, `items_string`, `results`, `permissions_generated`)
//              within the context of the given `{attempt_id}`.
//              Only items visible to the current user (or to the `{as_team_id}` team) are shown.
//              If `{watched_group_id}` is given, some additional info about the given group's results on the items is shown.
//
//
//              * The current user (or the team given in `as_team_id`) should have at least 'content' permissions on the specified item
//                and a started result for it, otherwise the 'forbidden' response is returned.
//
//              * If `as_team_id` is given, it should be a user's parent team group,
//                otherwise the "forbidden" error is returned.
//
//              * If `{watched_group_id}` is given, the user should ba a manager of the group with the 'can_watch_members' permission,
//                otherwise the "forbidden" error is returned.
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   format: int64
//   required: true
// - name: attempt_id
//   description: "`id` of an attempt for the item."
//   in: query
//   type: integer
//   required: true
// - name: as_team_id
//   in: query
//   type: integer
// - name: watched_group_id
//   in: query
//   type: integer
// responses:
//   "200":
//     description: OK. Success response with item children data
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/childItem"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getItemChildren(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	itemID, attemptID, participantID, user, watchedGroupID, watchedGroupIDSet, apiError :=
		srv.resolveGetParentsOrChildrenServiceParams(httpReq)
	if apiError != service.NoError {
		return apiError
	}

	found, err := srv.Store.Permissions().
		MatchingGroupAncestors(participantID).
		WherePermissionIsAtLeast("view", "content").
		Joins("JOIN results ON results.participant_id = ? AND results.item_id = permissions.item_id", participantID).
		Where("permissions.item_id = ?", itemID).
		Where("results.attempt_id = ?", attemptID).
		Where("results.started_at IS NOT NULL").
		HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	var rawData []rawListChildItem
	service.MustNotBeError(
		constructItemChildrenQuery(srv.Store, itemID, participantID, attemptID, watchedGroupIDSet, watchedGroupID,
			`items.allows_multiple_attempts, category, score_weight, content_view_propagation,
				upper_view_levels_propagation, grant_view_propagation, watch_propagation, edit_propagation,
				items.id, items.type, items.default_language_tag,
				validation_type, display_details_in_parent, duration, entry_participant_type, no_score,
				can_view_generated_value, can_grant_view_generated_value, can_watch_generated_value, can_edit_generated_value, is_owner_generated,
				IFNULL(
					(SELECT MAX(results.score_computed) AS best_score
					FROM results
					WHERE results.item_id = items.id AND results.participant_id = ?), 0) AS best_score,
				child_order,
				EXISTS(SELECT 1 FROM item_dependencies WHERE item_id = items.id AND grant_content_view) AS grants_access_to_items`,
			[]interface{}{participantID},
			`COALESCE(user_strings.language_tag, default_strings.language_tag) AS language_tag,
			 IF(user_strings.language_tag IS NULL, default_strings.title, user_strings.title) AS title,
			 IF(user_strings.language_tag IS NULL, default_strings.subtitle, user_strings.subtitle) AS subtitle`).
			JoinsUserAndDefaultItemStrings(user).
			Scan(&rawData).Error())

	response := srv.childItemsFromRawData(rawData, watchedGroupIDSet, srv.Store.PermissionsGranted())

	render.Respond(rw, httpReq, response)
	return service.NoError
}

func constructItemChildrenQuery(dataStore *database.DataStore, parentItemID, groupID, attemptID int64,
	watchedGroupIDSet bool, watchedGroupID int64, columnList string, columnListValues []interface{},
	externalColumnList string) *database.DB {
	return constructItemListQuery(
		dataStore, groupID, watchedGroupIDSet, watchedGroupID, columnList, columnListValues,
		externalColumnList,
		func(db *database.DB) *database.DB {
			return db.Joins("JOIN items_items ON items_items.parent_item_id = ? AND items_items.child_item_id = items.id", parentItemID)
		},
		func(db *database.DB) *database.DB {
			return db.Joins("JOIN items_items ON items_items.child_item_id = item_id").
				Where("items_items.parent_item_id = ?", parentItemID)
		},
		func(db *database.DB) *database.DB {
			return db.
				Where(
					"WHERE attempts.id IS NULL OR IF(attempts.root_item_id <=> results.item_id, attempts.parent_attempt_id, attempts.id) = ?",
					attemptID)
		})
}

func (srv *Service) childItemsFromRawData(
	rawData []rawListChildItem, watchedGroupIDSet bool, permissionGrantedStore *database.PermissionGrantedStore) []childItem {
	result := make([]childItem, 0, len(rawData))
	var currentChild *childItem
	for index := range rawData {
		if index == 0 || rawData[index].ID != rawData[index-1].ID {
			child := childItem{
				commonItemFields: rawData[index].RawCommonItemFields.asItemCommonFields(permissionGrantedStore),
				BestScore:        rawData[index].BestScore,
				Results:          make([]structures.ItemResult, 0, 1),
				String: listItemString{
					LanguageTag: rawData[index].StringLanguageTag,
					Title:       rawData[index].StringTitle,
				},
				Category:                   rawData[index].Category,
				ScoreWeight:                rawData[index].ScoreWeight,
				ContentViewPropagation:     rawData[index].ContentViewPropagation,
				UpperViewLevelsPropagation: rawData[index].UpperViewLevelsPropagation,
				GrantViewPropagation:       rawData[index].GrantViewPropagation,
				WatchPropagation:           rawData[index].WatchPropagation,
				EditPropagation:            rawData[index].EditPropagation,
				GrantsAccessToItems:        rawData[index].GrantsAccessToItems,
				Order:                      rawData[index].Order,
			}
			if rawData[index].CanViewGeneratedValue >= permissionGrantedStore.ViewIndexByName("content") {
				child.String.listItemStringNotInfo = &listItemStringNotInfo{Subtitle: rawData[index].StringSubtitle}
			}
			child.WatchedGroup = rawData[index].RawWatchedGroupStatFields.asItemWatchedGroupStat(watchedGroupIDSet, srv.Store.PermissionsGranted())
			result = append(result, child)
			currentChild = &result[len(result)-1]
		}

		itemResult := rawData[index].asItemResult()
		if itemResult != nil {
			currentChild.Results = append(currentChild.Results, *itemResult)
		}
	}
	return result
}

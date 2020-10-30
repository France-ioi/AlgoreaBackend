package items

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model prerequisiteItem
type prerequisiteItem struct {
	*commonItemFields

	// item_dependencies.score
	// required: true
	Score int `json:"score"`
	// required: true
	GrantContentView bool `json:"grant_content_view"`

	// required: true
	String listItemString `json:"string"`

	// max among all attempts of the user (or of the team given in `{as_team_id}`)
	// required: true
	BestScore float32 `json:"best_score"`

	WatchedGroup *itemWatchedGroupStat `json:"watched_group,omitempty"`
}

type rawPrerequisiteListItem struct {
	*RawCommonItemFields

	// from items_strings: in the user’s default language or (if not available) default language of the item
	StringLanguageTag string  `sql:"column:language_tag"`
	StringTitle       *string `sql:"column:title"`
	StringSubtitle    *string `sql:"column:subtitle"`

	// max from results of the current participant
	BestScore float32

	// from item_dependencies
	Score            int
	GrantContentView bool

	*RawWatchedGroupStatFields
}

// swagger:operation GET /items/{item_id}/prerequisites items itemPrerequisitesView
// ---
// summary: Get prerequisites for an item
// description: Lists prerequisite items for the specified item
//              and the current user's (or the team's given in `as_team_id`) interactions with them
//              (from tables `items`, `item_dependencies`, `items_string`, `results`, `permissions_generated`).
//              Only items visible to the current user (or to the `{as_team_id}` team) are shown.
//              If `{watched_group_id}` is given, some additional info about the given group's results on the items is shown.
//
//
//              * The current user (or the team given in `as_team_id`) should have at least 'info' permissions on the specified item,
//                otherwise the 'forbidden' response is returned.
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
//         "$ref": "#/definitions/prerequisiteItem"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getItemPrerequisites(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	itemID, err := service.ResolveURLQueryPathInt64Field(httpReq, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)
	participantID := service.ParticipantIDFromContext(httpReq.Context())

	watchedGroupID, watchedGroupIDSet, apiError := srv.resolveWatchedGroupID(httpReq)
	if apiError != service.NoError {
		return apiError
	}

	found, err := srv.Store.Permissions().
		MatchingGroupAncestors(participantID).
		WherePermissionIsAtLeast("view", "info").
		Where("permissions.item_id = ?", itemID).
		HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.InsufficientAccessRightsError
	}

	var rawData []rawPrerequisiteListItem
	service.MustNotBeError(
		constructItemListWithoutResultsQuery(
			srv.Store, participantID, watchedGroupIDSet, watchedGroupID,
			`items.allows_multiple_attempts, items.id, items.type, items.default_language_tag,
				validation_type, display_details_in_parent, duration, entry_participant_type, no_score,
				can_view_generated_value, can_grant_view_generated_value, can_watch_generated_value, can_edit_generated_value, is_owner_generated,
				score, grant_content_view,
				IFNULL(
					(SELECT MAX(results.score_computed) AS best_score
					FROM results
					WHERE results.item_id = items.id AND results.participant_id = ?), 0) AS best_score,
				COALESCE(user_strings.language_tag, default_strings.language_tag) AS language_tag,
				IF(user_strings.language_tag IS NULL, default_strings.title, user_strings.title) AS title,
				IF(user_strings.language_tag IS NULL, default_strings.subtitle, user_strings.subtitle) AS subtitle`,
			[]interface{}{participantID},
			func(db *database.DB) *database.DB {
				return db.Joins("JOIN item_dependencies ON item_dependencies.dependent_item_id = ? AND item_dependencies.item_id = items.id", itemID).
					JoinsUserAndDefaultItemStrings(user)
			},
			func(db *database.DB) *database.DB {
				return db.Joins("JOIN item_dependencies ON item_dependencies.item_id = permissions.item_id").
					Where("item_dependencies.dependent_item_id = ?", itemID)
			}).
			Order("title, subtitle, id").
			Scan(&rawData).Error())

	response := srv.prerequisiteItemsFromRawData(rawData, watchedGroupIDSet, srv.Store.PermissionsGranted())

	render.Respond(rw, httpReq, response)
	return service.NoError
}

func (srv *Service) prerequisiteItemsFromRawData(
	rawData []rawPrerequisiteListItem, watchedGroupIDSet bool, permissionGrantedStore *database.PermissionGrantedStore) []prerequisiteItem {
	result := make([]prerequisiteItem, 0, len(rawData))
	for index := range rawData {
		child := prerequisiteItem{
			commonItemFields: rawData[index].RawCommonItemFields.asItemCommonFields(permissionGrantedStore),
			BestScore:        rawData[index].BestScore,
			String: listItemString{
				LanguageTag: rawData[index].StringLanguageTag,
				Title:       rawData[index].StringTitle,
			},
			Score:            rawData[index].Score,
			GrantContentView: rawData[index].GrantContentView,
		}
		if rawData[index].CanViewGeneratedValue >= permissionGrantedStore.ViewIndexByName("content") {
			child.String.listItemStringNotInfo = &listItemStringNotInfo{Subtitle: rawData[index].StringSubtitle}
		}
		child.WatchedGroup = rawData[index].RawWatchedGroupStatFields.asItemWatchedGroupStat(watchedGroupIDSet, srv.Store.PermissionsGranted())
		result = append(result, child)
	}
	return result
}

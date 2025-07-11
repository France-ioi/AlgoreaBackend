package items

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/structures"
)

// swagger:model parentItem
type parentItem struct {
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

	// max among all attempts of the user (or of the team given in `{as_team_id}`)
	// required: true
	BestScore float32 `json:"best_score"`
	// required:true
	Result *structures.ItemResult `json:"result"`

	WatchedGroup *itemWatchedGroupStat `json:"watched_group,omitempty"`
}

// swagger:operation GET /items/{item_id}/parents items itemParentsView
//
//	---
//	summary: Get item parents
//	description: Lists parents of the specified item
//						 and the current user's (or the team's given in `as_team_id`) interactions with them
//						 (from tables `items`, `items_items`, `items_string`, `results`, `permissions_generated`)
//						 within the context of the given `{attempt_id}`.
//						 Only items visible to the current user (or to the `{as_team_id}` team) are shown.
//						 If `{watched_group_id}` is given, some additional info about the given group's results on the items is shown.
//
//
//						 * The current user (or the team given in `as_team_id`) should have at least 'info' permissions on the specified item,
//							 otherwise the 'forbidden' response is returned.
//
//						 * If `as_team_id` is given, it should be a user's parent team group,
//							 otherwise the "forbidden" error is returned.
//
//						 * If `{watched_group_id}` is given, the user should ba a manager of the group with the 'can_watch_members' permission,
//							 otherwise the "forbidden" error is returned.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: attempt_id
//			description: "`id` of an attempt for the item."
//			in: query
//			type: integer
//			format: int64
//			required: true
//		- name: as_team_id
//			in: query
//			type: integer
//			format: int64
//		- name: watched_group_id
//			in: query
//			type: integer
//			format: int64
//	responses:
//		"200":
//			description: OK. Success response with item parents data
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/parentItem"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getItemParents(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	params, err := srv.resolveGetParentsOrChildrenServiceParams(httpRequest)
	service.MustNotBeError(err)

	store := srv.GetStore(httpRequest)
	found, err := store.Permissions().
		MatchingGroupAncestors(params.participantID).
		WherePermissionIsAtLeast("view", "info").
		Where("permissions.item_id = ?", params.itemID).HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.ErrAPIInsufficientAccessRights
	}

	var rawData []RawListItem
	service.MustNotBeError(
		constructItemParentsQuery(store, params.itemID, params.participantID,
			params.attemptID, params.watchedGroupIDIsSet, params.watchedGroupID).
			JoinsUserAndDefaultItemStrings(params.user).
			Scan(&rawData).Error())

	response := parentItemsFromRawData(rawData, params.watchedGroupIDIsSet, store.PermissionsGranted())

	render.Respond(responseWriter, httpRequest, response)
	return nil
}

type getParentsOrChildrenServiceParams struct {
	itemID              int64
	attemptID           int64
	participantID       int64
	user                *database.User
	watchedGroupID      int64
	watchedGroupIDIsSet bool
}

func (srv *Service) resolveGetParentsOrChildrenServiceParams(httpReq *http.Request) (
	parameters *getParentsOrChildrenServiceParams, err error,
) {
	var params getParentsOrChildrenServiceParams
	params.itemID, err = service.ResolveURLQueryPathInt64Field(httpReq, "item_id")
	if err != nil {
		return nil, service.ErrInvalidRequest(err)
	}

	params.attemptID, err = service.ResolveURLQueryGetInt64Field(httpReq, "attempt_id")
	if err != nil {
		return nil, service.ErrInvalidRequest(err)
	}

	params.user = srv.GetUser(httpReq)
	params.participantID = service.ParticipantIDFromContext(httpReq.Context())

	params.watchedGroupID, params.watchedGroupIDIsSet, err = srv.ResolveWatchedGroupID(httpReq)
	if err != nil {
		return nil, err
	}

	return &params, nil
}

func constructItemParentsQuery(dataStore *database.DataStore, childItemID, groupID, attemptID int64,
	watchedGroupIDIsSet bool, watchedGroupID int64,
) *database.DB {
	return constructItemListQuery(
		dataStore, groupID, "info", watchedGroupIDIsSet, watchedGroupID,
		`items.allows_multiple_attempts, category, items.id, items.type, items.default_language_tag,
			validation_type, display_details_in_parent, duration, entry_participant_type, no_score,
			can_view_generated_value, can_grant_view_generated_value, can_watch_generated_value, can_edit_generated_value, is_owner_generated,
			IFNULL(
				(SELECT MAX(results.score_computed) AS best_score
				FROM results
				WHERE results.item_id = items.id AND results.participant_id = ?), 0) AS best_score,
			child_order`,
		[]interface{}{groupID},
		`COALESCE(user_strings.language_tag, default_strings.language_tag) AS language_tag,
			 IF(user_strings.language_tag IS NULL, default_strings.title, user_strings.title) AS title,
			 IF(user_strings.image_url IS NULL, default_strings.image_url, user_strings.image_url) AS image_url,
			 IF(user_strings.language_tag IS NULL, default_strings.subtitle, user_strings.subtitle) AS subtitle`,
		func(db *database.DB) *database.DB {
			return db.Joins("JOIN items_items ON items_items.child_item_id = ? AND items_items.parent_item_id = items.id", childItemID)
		},
		func(db *database.DB) *database.DB {
			return db.Joins("JOIN items_items ON items_items.parent_item_id = item_id").
				Where("items_items.child_item_id = ?", childItemID)
		},
		func(db *database.DB) *database.DB {
			return db.
				Where(`
					WHERE attempts.id IS NULL OR
						attempts.id = (SELECT IF(root_item_id = ?, parent_attempt_id, id) FROM attempts WHERE id = ? AND participant_id = ?)`,
					childItemID, attemptID, groupID)
		})
}

func parentItemsFromRawData(rawData []RawListItem, watchedGroupIDIsSet bool,
	permissionGrantedStore *database.PermissionGrantedStore,
) []parentItem {
	result := make([]parentItem, 0, len(rawData))
	for index := range rawData {
		item := parentItem{
			commonItemFields: rawData[index].RawCommonItemFields.asItemCommonFields(permissionGrantedStore),
			BestScore:        rawData[index].BestScore,
			Result:           rawData[index].asItemResult(),
			String: listItemString{
				LanguageTag: rawData[index].StringLanguageTag,
				ImageURL:    rawData[index].StringImageURL,
				Title:       rawData[index].StringTitle,
			},
			Category: rawData[index].Category,
			Order:    rawData[index].Order,
		}
		if rawData[index].CanViewGeneratedValue >= permissionGrantedStore.ViewIndexByName("content") {
			item.String.listItemStringNotInfo = &listItemStringNotInfo{Subtitle: rawData[index].StringSubtitle}
		}
		item.WatchedGroup = rawData[index].RawWatchedGroupStatFields.asItemWatchedGroupStat(watchedGroupIDIsSet, permissionGrantedStore)
		result = append(result, item)
	}
	return result
}

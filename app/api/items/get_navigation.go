package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/structures"
)

// swagger:model itemNavigationResponse
type itemNavigationResponse struct {
	*structures.ItemCommonFields

	// required: true
	AttemptID int64 `json:"attempt_id,string"`
	// required: true
	Children []navigationItemChild `json:"children"`
}

type navigationItemChild struct {
	*structures.ItemCommonFields

	// required: true
	RequiresExplicitEntry bool `json:"requires_explicit_entry"`
	// required: true
	// enum: User,Team
	EntryParticipantType string `json:"entry_participant_type"`
	// required: true
	NoScore bool `json:"no_score"`
	// whether the item has children visible to the user and, at the same time,
	// the user has can_view >= 'content' permission on the item
	// required: true
	HasVisibleChildren bool `json:"has_visible_children"`
	// max among all attempts of the user (or of the team given in `{as_team_id}`)
	// required: true
	BestScore float32 `json:"best_score"`
	// required:true
	Results []structures.ItemResult `json:"results"`

	WatchedGroup *itemWatchedGroupStat `json:"watched_group,omitempty"`
}

// only if `{watched_group_id}` is given.
type itemWatchedGroupStat struct {
	// group's view permission on this item
	// required: true
	// enum: none,info,content,content_with_descendants,solution
	CanView string `json:"can_view"`
	// [only if the current user can watch for this item]
	// average of the max scores of every descendant participants of the input group
	// (= 0 if a participant has no result yet on the item)
	AvgScore *float32 `json:"avg_score,omitempty"`
	// [only if the current user can watch for this item]
	// whether all descendant participants have accomplished the item (validated = true)
	AllValidated *bool `json:"all_validated,omitempty"`
}

// swagger:operation GET /items/{item_id}/navigation items itemNavigationView
//
//	---
//	summary: Get navigation data
//	description: >
//
//		Returns data needed to display the navigation menu (for `item_id` and its children)
//		within the context of the given `{attempt_id}`/`{child_attempt_id}` (one of those should be given).
//
//		If the given `item_id` is a Skill, the children returned are only Skills.
//		If it is a Chapter, the children returned are Tasks and Chapters.
//
//		Only items visible to the current user (or to the `{as_team_id}` team) are shown.
//		If `{watched_group_id}` is given, some additional info about the given group's results on the items is shown.
//
//
//		If `{child_attempt_id}` is given, the context-defining attempt id of the input item
//		is either the same `{child_attempt_id}` or the `parent_attempt_id` of the given `{child_attempt_id}`
//		(depending on the `root_item_id` of the `{child_attempt_id}`).
//
//
//		* If the specified `{item_id}` doesn't exist or is not visible to the current user (or to the `{as_team_id}` team),
//			of if there is no started result of the user/`{as_team_id}` for the context attempt id and the item,
//			the 'forbidden' response is returned.
//
//
//		* If `{as_team_id}` is given, it should be a user's parent team group,
//			otherwise the "forbidden" error is returned.
//
//
//		* If `{watched_group_id}` is given, the user should ba a manager of the group with the 'can_watch_members' permission,
//			otherwise the "forbidden" error is returned.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: attempt_id
//			description: "`id` of an attempt for the item. This parameter is incompatible with `{child_attempt_id}`."
//			in: query
//			type: integer
//			format: int64
//		- name: child_attempt_id
//			description: "`id` of an attempt for one of the item's children. This parameter is incompatible with `{attempt_id}`."
//			in: query
//			type: integer
//			format: int64
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
//			description: OK. Navigation data
//			schema:
//				"$ref": "#/definitions/itemNavigationResponse"
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
func (srv *Service) getItemNavigation(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpRequest)
	participantID := service.ParticipantIDFromContext(httpRequest.Context())
	store := srv.GetStore(httpRequest)

	attemptID, err := resolveAttemptIDForNavigationData(store, httpRequest, participantID, itemID)
	service.MustNotBeError(err)

	watchedGroupID, watchedGroupIDIsSet, err := srv.ResolveWatchedGroupID(httpRequest)
	service.MustNotBeError(err)

	rawData := getRawNavigationData(store, itemID, participantID, attemptID, user, watchedGroupID, watchedGroupIDIsSet)

	if len(rawData) == 0 || rawData[0].ID != itemID {
		return service.ErrForbidden(errors.New("insufficient access rights on given item id"))
	}

	response := itemNavigationResponse{
		ItemCommonFields: fillItemCommonFieldsWithDBData(store, &rawData[0]),
		AttemptID:        *rawData[0].AttemptID,
	}
	idMap := map[int64]*rawNavigationItem{}
	for index := range rawData {
		idMap[rawData[index].ID] = &rawData[index]
	}
	fillNavigationWithChildren(store, rawData, watchedGroupIDIsSet, &response.Children)

	render.Respond(responseWriter, httpRequest, response)
	return nil
}

func resolveAttemptIDForNavigationData(store *database.DataStore, httpReq *http.Request, groupID, itemID int64) (int64, error) {
	attemptIDSet := len(httpReq.URL.Query()["attempt_id"]) != 0
	childAttemptIDSet := len(httpReq.URL.Query()["child_attempt_id"]) != 0
	var attemptID, childAttemptID int64
	var err error
	if attemptIDSet {
		if childAttemptIDSet {
			return 0, service.ErrInvalidRequest(errors.New("only one of attempt_id and child_attempt_id can be given"))
		}
		attemptID, err = service.ResolveURLQueryGetInt64Field(httpReq, "attempt_id")
		if err != nil {
			return 0, service.ErrInvalidRequest(err)
		}
	}
	if childAttemptIDSet {
		childAttemptID, err = service.ResolveURLQueryGetInt64Field(httpReq, "child_attempt_id")
		if err != nil {
			return 0, service.ErrInvalidRequest(err)
		}
	}
	if !attemptIDSet && !childAttemptIDSet {
		return 0, service.ErrInvalidRequest(errors.New("one of attempt_id and child_attempt_id should be given"))
	}

	if !attemptIDSet {
		err := store.Table("results AS child_result").
			Where("child_result.participant_id = ? AND child_result.attempt_id = ? AND child_result.started",
				groupID, childAttemptID).
			Joins(`
				JOIN attempts AS child_attempt ON child_attempt.participant_id = child_result.participant_id AND
					child_attempt.id = child_result.attempt_id`).
			Joins(`
				JOIN items_items ON items_items.parent_item_id = ? AND items_items.child_item_id = child_result.item_id`, itemID).
			PluckFirst("IF(child_attempt.root_item_id = child_result.item_id, child_attempt.parent_attempt_id, child_attempt.id)", &attemptID).
			Error()
		if gorm.IsRecordNotFoundError(err) {
			return 0, service.ErrAPIInsufficientAccessRights
		}
		service.MustNotBeError(err)
	}
	return attemptID, nil
}

func fillNavigationWithChildren(
	store *database.DataStore, rawData []rawNavigationItem, watchedGroupIDIsSet bool, target *[]navigationItemChild,
) {
	*target = make([]navigationItemChild, 0, len(rawData)-1)
	var currentChild *navigationItemChild
	if len(rawData) > 0 && rawData[0].CanViewGeneratedValue == store.PermissionsGranted().ViewIndexByName("info") {
		return // Only 'info' access to the parent item
	}
	for index := range rawData {
		if index == 0 {
			continue
		}

		if rawData[index].ID != rawData[index-1].ID {
			child := navigationItemChild{
				ItemCommonFields:      fillItemCommonFieldsWithDBData(store, &rawData[index]),
				RequiresExplicitEntry: rawData[index].RequiresExplicitEntry,
				EntryParticipantType:  rawData[index].EntryParticipantType,
				NoScore:               rawData[index].NoScore,
				HasVisibleChildren:    rawData[index].HasVisibleChildren,
				BestScore:             rawData[index].BestScore,
				Results:               make([]structures.ItemResult, 0, 1),
			}
			if rawData[index].CanViewGeneratedValue < store.PermissionsGranted().ViewIndexByName("content") {
				child.HasVisibleChildren = false
			}
			child.WatchedGroup = rawData[index].asItemWatchedGroupStat(watchedGroupIDIsSet, store.PermissionsGranted())
			*target = append(*target, child)
			currentChild = &(*target)[len(*target)-1]
		}

		result := rawData[index].asItemResult()
		if result != nil {
			currentChild.Results = append(currentChild.Results, *result)
		}
	}
}

func fillItemCommonFieldsWithDBData(store *database.DataStore, rawData *rawNavigationItem) *structures.ItemCommonFields {
	result := &structures.ItemCommonFields{
		ID:          rawData.ID,
		Type:        rawData.Type,
		String:      structures.ItemString{Title: rawData.Title, LanguageTag: rawData.LanguageTag},
		Permissions: *rawData.RawGeneratedPermissionFields.AsItemPermissions(store.PermissionsGranted()),
	}
	return result
}

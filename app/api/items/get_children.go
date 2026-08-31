package items

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/structures"
)

type listItemStringNotInfo struct {
	// Only if `can_view` >= 'content'
	Subtitle *string `json:"subtitle"`
}

type listItemString struct {
	*listItemStringNotInfo

	// required: true
	LanguageTag string `json:"language_tag"`
	// required: true
	Title *string `json:"title"`
	// required: true
	ImageURL *string `json:"image_url"`
}

// childItemDescriptionFields is embedded only when `include_description` is given and `can_view` >= content.
// A nil embed omits `description`; a non-nil embed always emits it (including JSON null when the DB value is NULL).
type childItemDescriptionFields struct {
	Description *string `json:"description"`
}

// only for visible items.
type visibleChildItemString struct {
	*listItemString
	*childItemDescriptionFields
}

type visibleChildItemFields struct {
	String visibleChildItemString `json:"string"`
	// only for visible items
	DefaultLanguageTag string `json:"default_language_tag"`

	// max among all attempts of the user (or of the team given in `{as_team_id}`)
	// (only for visible items)
	BestScore float32 `json:"best_score"`
	// only for visible items
	Results []structures.ItemResult `json:"results"`

	// items

	// only for visible items
	// enum: None,All,AllButOne,Categories,One,Manual
	ValidationType string `json:"validation_type"`
	// only for visible items
	RequiresExplicitEntry bool `json:"requires_explicit_entry"`
	// only for visible items
	AllowsMultipleAttempts bool `json:"allows_multiple_attempts"`
	// only for visible items
	// enum: User,Team
	EntryParticipantType string `json:"entry_participant_type"`
	// only for visible items
	// pattern: ^\d{1,3}:[0-5]?\d:[0-5]?\d$
	// example: 838:59:59
	Duration *string `json:"duration"`
	// only for visible items
	NoScore bool `json:"no_score"`

	// whether solving this item grants access to some items (visible or not)
	// (only for visible items)
	GrantsAccessToItems bool `json:"grants_access_to_items"`

	// JSON object with display/UI settings interpreted by the frontend.
	// Absent for invisible children (like every other field on this struct);
	// when present, it is always an object (possibly `{}`), never `null`.
	// only for visible items
	DisplaySettings database.JSON `json:"display_settings"`
}

// swagger:model childItem
type childItem struct {
	*visibleChildItemFields

	// required: true
	ID int64 `json:"id,string"`
	// required: true
	// enum: Chapter,Task,Skill
	Type string `json:"type"`

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
	// `items_items.request_help_propagation`
	// required: true
	RequestHelpPropagation bool `json:"request_help_propagation"`

	// required: true
	Permissions structures.ItemPermissions `json:"permissions"`

	WatchedGroup *itemWatchedGroupStat `json:"watched_group,omitempty"`

	// The child's own children, same format (without further nesting of `children`).
	// Only on direct children of `{item_id}`, and only if `show_level2_children` is given,
	// the child is not a Task, is visible at content level, and its public `results` has at least
	// one entry with `started_at` not null. A qualifying child with no visible children yields `[]`.
	Children *[]childItem `json:"children,omitempty"`
}

// RawListItem contains raw fields common for itemChildrenView & itemParentsView.
type RawListItem struct {
	*RawCommonItemFields
	*RawItemResultFields
	*RawWatchedGroupStatFields

	// from items_strings: in the user’s default language or (if not available) default language of the item
	StringLanguageTag string  `sql:"column:language_tag"`
	StringTitle       *string `sql:"column:title"`
	StringImageURL    *string `sql:"column:image_url"`
	StringSubtitle    *string `sql:"column:subtitle"`

	// items_items
	Category string
	Order    int32 `sql:"column:child_order"`

	// max from results of the current participant
	BestScore float32
}

type rawListChildItem struct {
	*RawListItem

	// from items_strings (children query only; not mapped onto parents/prerequisites)
	StringDescription *string `sql:"column:description"`

	// items_items
	ScoreWeight                int8
	ContentViewPropagation     string
	UpperViewLevelsPropagation string
	GrantViewPropagation       bool
	WatchPropagation           bool
	EditPropagation            bool
	RequestHelpPropagation     bool
	// Populated only by the level-2 batch scan (selected via additionalColumnList).
	// Left at zero and intentionally unread on level-1 scans — do not wire it up there.
	ParentItemID int64

	// items
	// `items.display_settings` is `NOT NULL` (defaults to `{}`), so we can read it
	// into a non-pointer `database.JSON`; downstream code never needs a nil check.
	DisplaySettings database.JSON

	// item_dependencies
	GrantsAccessToItems bool
}

// swagger:operation GET /items/{item_id}/children items itemChildrenView
//
//	---
//	summary: Get item children
//	description: Lists children of the specified item
//						 and the current user's (or the team's given in `as_team_id`) interactions with them
//						 (from tables `items`, `items_items`, `items_string`, `results`, `permissions_generated`)
//						 within the context of the given `{attempt_id}`.
//						 Only items visible to the current user (or to the `{as_team_id}` team) are shown.
//						 If `{show_invisible_items}` = 1, items invisible to the current user (or to the `{as_team_id}` team) are shown too,
//						 but with a limited set of fields.
//						 If `{watched_group_id}` is given, some additional info about the given group's results on the items is shown.
//						 If `{show_level2_children}` is given (presence-only; any value enables it), each direct child that is
//						 not a Task, is visible at content level, and has at least one started entry in its public `results`
//						 also includes a `children` array with that child's own children (same format, without further
//						 `children` nesting). Tasks and children with empty `results` (e.g. when the attempt’s `root_item_id`
//						 is the child) do not nest.
//						 If `{include_description}` is given (presence-only; any value enables it), each direct child that
//						 exposes `string` and is visible at content level also includes `string.description`
//						 (user language preferred, else the item’s default language; JSON `null` when the resolved
//						 description is NULL). Nested level-2 `children` never include `description`.
//						 The content-level gate is stricter than `itemView`, which returns `description` from `info` view.
//
//
//						 * The current user (or the team given in `as_team_id`) should have at least 'content' permissions on the specified item
//							 and a started result for it, otherwise the 'forbidden' response is returned.
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
//		- name: show_invisible_items
//			in: query
//			description: If 1, show invisible items as well
//			type: integer
//			enum: [0,1]
//			default: 0
//		- name: show_level2_children
//			in: query
//			description: |
//				Presence-only flag. When present (any value, including `0`), each direct child of `{item_id}` that is
//				not a Task, is visible at content level, and has at least one started entry in its public `results`
//				(`results[].started_at` not null) includes a `children` array with that child's own children (nested
//				entries do not themselves include `children`). Tasks and children with empty `results` do not nest.
//			type: string
//			required: false
//		- name: include_description
//			in: query
//			description: |
//				Presence-only flag. When present (any value, including `0`), each direct child of `{item_id}` that
//				exposes `string` and is visible at content level (`can_view` >= content) includes `string.description`
//				from `items_strings` (user language preferred, else the item’s default; JSON `null` when NULL).
//				Nested level-2 children never include `description`. Stricter than `itemView`, which returns
//				`description` already at `info` view.
//			type: string
//			required: false
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
//			description: OK. Success response with item children data
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/childItem"
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
func (srv *Service) getItemChildren(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	params, err := srv.resolveGetParentsOrChildrenServiceParams(httpRequest)
	service.MustNotBeError(err)

	requiredViewPermissionOnItems := "info"
	if len(httpRequest.URL.Query()["show_invisible_items"]) > 0 {
		var showInvisibleItems bool
		showInvisibleItems, err = service.ResolveURLQueryGetBoolField(httpRequest, "show_invisible_items")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}
		if showInvisibleItems {
			requiredViewPermissionOnItems = "none"
		}
	}
	showLevel2Children := len(httpRequest.URL.Query()["show_level2_children"]) > 0
	includeDescription := len(httpRequest.URL.Query()["include_description"]) > 0

	store := srv.GetStore(httpRequest)
	found, err := store.Permissions().
		MatchingGroupAncestors(params.participantID).
		WherePermissionIsAtLeast("view", "content").
		Joins("JOIN results ON results.participant_id = ? AND results.item_id = permissions.item_id", params.participantID).
		Where("permissions.item_id = ?", params.itemID).
		Where("results.attempt_id = ?", params.attemptID).
		Where("results.started").
		HasRows()
	service.MustNotBeError(err)
	if !found {
		return service.ErrAPIInsufficientAccessRights
	}

	var rawData []rawListChildItem
	scanItemChildren(store, []int64{params.itemID}, params, requiredViewPermissionOnItems, "", includeDescription, &rawData)

	response := childItemsFromRawData(rawData, params.watchedGroupIDIsSet, store.PermissionsGranted(), includeDescription)
	if showLevel2Children {
		fillLevel2Children(response, rawData, store, params, requiredViewPermissionOnItems)
	}

	render.Respond(responseWriter, httpRequest, response)
	return nil
}

// itemChildrenColumnList is the list of SQL columns selected for a child item, shared by both
// the level-1 and the level-2 (`show_level2_children`) queries so that they yield the same format.
// The `?` placeholder is the participant id used by the `best_score` sub-query.
const itemChildrenColumnList = `items.allows_multiple_attempts, category, score_weight, content_view_propagation,
				upper_view_levels_propagation, grant_view_propagation, watch_propagation, edit_propagation, request_help_propagation,
				items.id, items.type, items.default_language_tag,
				items.validation_type, items.duration, items.entry_participant_type, items.no_score,
				items.display_settings,
				IFNULL(can_view_generated_value, 1) AS can_view_generated_value,
				IFNULL(can_grant_view_generated_value, 1) AS can_grant_view_generated_value,
				IFNULL(can_watch_generated_value, 1) AS can_watch_generated_value,
				IFNULL(can_edit_generated_value, 1) AS can_edit_generated_value,
				IFNULL(is_owner_generated, 0) is_owner_generated,
				IFNULL(
					(SELECT MAX(results.score_computed) AS best_score
					FROM results
					WHERE results.item_id = items.id AND results.participant_id = ?), 0) AS best_score,
				child_order,
				EXISTS(SELECT 1 FROM item_dependencies WHERE item_id = items.id AND grant_content_view) AS grants_access_to_items`

// itemChildrenExternalColumnList is the list of SQL columns resolved outside of the items sub-query,
// on the strings joined by JoinsUserAndDefaultItemStrings.
const itemChildrenExternalColumnList = `COALESCE(user_strings.language_tag, default_strings.language_tag) AS language_tag,
			 IF(user_strings.language_tag IS NULL, default_strings.title, user_strings.title) AS title,
			 IF(user_strings.image_url IS NULL, default_strings.image_url, user_strings.image_url) AS image_url,
			 IF(user_strings.language_tag IS NULL, default_strings.subtitle, user_strings.subtitle) AS subtitle`

// Appended to itemChildrenExternalColumnList only for the level-1 scan when include_description is set.
const itemChildrenDescriptionColumn = `,
			 IF(user_strings.language_tag IS NULL, default_strings.description, user_strings.description) AS description`

func scanItemChildren(store *database.DataStore, parentItemIDs []int64, params *getParentsOrChildrenServiceParams,
	requiredViewPermissionOnItems, additionalColumnList string, includeDescription bool, dest interface{},
) {
	externalColumnList := itemChildrenExternalColumnList
	if includeDescription {
		externalColumnList += itemChildrenDescriptionColumn
	}
	service.MustNotBeError(
		constructItemChildrenQuery(
			store,
			parentItemIDs,
			params.participantID,
			requiredViewPermissionOnItems,
			params.attemptID,
			params.watchedGroupIDIsSet,
			params.watchedGroupID,
			itemChildrenColumnList+additionalColumnList,
			[]interface{}{params.participantID},
			externalColumnList,
			func(db *database.DB) *database.DB {
				return db.Joins(
					"JOIN items_items ON items_items.parent_item_id IN (?) AND items_items.child_item_id = items.id",
					parentItemIDs)
			},
		).
			JoinsUserAndDefaultItemStrings(params.user).
			Scan(dest).Error())
}

func childHasStartedResult(child *childItem) bool {
	if child.visibleChildItemFields == nil {
		return false
	}
	for index := range child.Results {
		if child.Results[index].StartedAt != nil {
			return true
		}
	}
	return false
}

// level1IDsEligibleForChildren returns IDs of response children that nest under show_level2_children:
// not a Task, content view (from raw rows), and at least one started entry in the public results array.
// Requiring a public started result (not a raw DB started row alone) means attempts whose
// root_item_id is the child itself — which yield empty `results` via HasAttempt — do not nest.
func level1IDsEligibleForChildren(
	response []childItem, rawData []rawListChildItem, contentViewIndex int,
) []int64 {
	hasContentByItemID := make(map[int64]bool, len(response))
	for index := range rawData {
		if rawData[index].CanViewGeneratedValue >= contentViewIndex {
			hasContentByItemID[rawData[index].ID] = true
		}
	}

	startedIDs := make([]int64, 0, len(response))
	for index := range response {
		if response[index].Type == "Task" {
			continue
		}
		// Check public results first so invisible children (nil visible fields) exercise that branch.
		if childHasStartedResult(&response[index]) && hasContentByItemID[response[index].ID] {
			startedIDs = append(startedIDs, response[index].ID)
		}
	}
	return startedIDs
}

func fillLevel2Children(
	response []childItem, rawData []rawListChildItem, store *database.DataStore,
	params *getParentsOrChildrenServiceParams, requiredViewPermissionOnItems string,
) {
	permissionGrantedStore := store.PermissionsGranted()
	startedIDs := level1IDsEligibleForChildren(
		response, rawData, permissionGrantedStore.ViewIndexByName("content"))
	if len(startedIDs) == 0 {
		return
	}

	var rawLevel2 []rawListChildItem
	// Level-2 never selects or maps description (includeDescription=false).
	scanItemChildren(store, startedIDs, params, requiredViewPermissionOnItems, ", items_items.parent_item_id", false, &rawLevel2)

	rawByParent := make(map[int64][]rawListChildItem, len(startedIDs))
	for index := range rawLevel2 {
		parentID := rawLevel2[index].ParentItemID
		rawByParent[parentID] = append(rawByParent[parentID], rawLevel2[index])
	}

	startedIDSet := make(map[int64]struct{}, len(startedIDs))
	for _, id := range startedIDs {
		startedIDSet[id] = struct{}{}
	}
	for index := range response {
		if _, ok := startedIDSet[response[index].ID]; !ok {
			continue
		}
		children := childItemsFromRawData(rawByParent[response[index].ID], params.watchedGroupIDIsSet, permissionGrantedStore, false)
		response[index].Children = &children
	}
}

func constructItemChildrenQuery(
	dataStore *database.DataStore,
	parentItemIDs []int64,
	groupID int64,
	requiredViewPermissionOnItems string,
	attemptID int64,
	watchedGroupIDIsSet bool,
	watchedGroupID int64,
	columnList string,
	columnListValues []interface{},
	externalColumnList string,
	joinItemRelationsToItemsFunc func(*database.DB) *database.DB,
) *database.DB {
	return constructItemListQuery(
		dataStore,
		groupID,
		requiredViewPermissionOnItems,
		watchedGroupIDIsSet,
		watchedGroupID,
		columnList,
		columnListValues,
		externalColumnList,
		joinItemRelationsToItemsFunc,
		func(db *database.DB) *database.DB {
			return db.Joins("JOIN items_items ON items_items.child_item_id = item_id").
				Where("items_items.parent_item_id IN (?)", parentItemIDs)
		},
		func(db *database.DB) *database.DB {
			return db.Where("IF(attempts.root_item_id <=> results.item_id, attempts.parent_attempt_id, attempts.id) = ?", attemptID)
		})
}

func childItemsFromRawData(
	rawData []rawListChildItem, watchedGroupIDIsSet bool, permissionGrantedStore *database.PermissionGrantedStore,
	includeDescription bool,
) []childItem {
	result := make([]childItem, 0, len(rawData))
	var currentChild *childItem
	for index := range rawData {
		if index == 0 || rawData[index].ID != rawData[index-1].ID {
			child := childItem{
				ID:                         rawData[index].ID,
				Order:                      rawData[index].Order,
				Category:                   rawData[index].Category,
				Type:                       rawData[index].Type,
				ScoreWeight:                rawData[index].ScoreWeight,
				ContentViewPropagation:     rawData[index].ContentViewPropagation,
				UpperViewLevelsPropagation: rawData[index].UpperViewLevelsPropagation,
				GrantViewPropagation:       rawData[index].GrantViewPropagation,
				WatchPropagation:           rawData[index].WatchPropagation,
				EditPropagation:            rawData[index].EditPropagation,
				RequestHelpPropagation:     rawData[index].RequestHelpPropagation,
				Permissions:                *rawData[index].AsItemPermissions(permissionGrantedStore),
			}
			if rawData[index].CanViewGeneratedValue >= permissionGrantedStore.ViewIndexByName("info") {
				child.visibleChildItemFields = &visibleChildItemFields{
					String: visibleChildItemString{
						listItemString: &listItemString{
							LanguageTag: rawData[index].StringLanguageTag,
							Title:       rawData[index].StringTitle,
							ImageURL:    rawData[index].StringImageURL,
						},
					},
					DefaultLanguageTag:     rawData[index].DefaultLanguageTag,
					BestScore:              rawData[index].BestScore,
					Results:                make([]structures.ItemResult, 0, 1),
					ValidationType:         rawData[index].ValidationType,
					RequiresExplicitEntry:  rawData[index].RequiresExplicitEntry,
					AllowsMultipleAttempts: rawData[index].AllowsMultipleAttempts,
					EntryParticipantType:   rawData[index].EntryParticipantType,
					Duration:               rawData[index].Duration,
					NoScore:                rawData[index].NoScore,
					GrantsAccessToItems:    rawData[index].GrantsAccessToItems,
					// OrEmpty() defends the documented "never null" contract against a
					// stray DB NULL (the column is NOT NULL, but `JSON.Scan` decodes
					// any NULL it sees into a nil map, which would marshal to `null`).
					DisplaySettings: rawData[index].DisplaySettings.OrEmpty(),
				}
			}
			if rawData[index].CanViewGeneratedValue >= permissionGrantedStore.ViewIndexByName("content") {
				child.String.listItemStringNotInfo = &listItemStringNotInfo{Subtitle: rawData[index].StringSubtitle}
				if includeDescription {
					child.String.childItemDescriptionFields = &childItemDescriptionFields{
						Description: rawData[index].StringDescription,
					}
				}
			}
			child.WatchedGroup = rawData[index].asItemWatchedGroupStat(watchedGroupIDIsSet, permissionGrantedStore)
			result = append(result, child)
			currentChild = &result[len(result)-1]
		}

		itemResult := rawData[index].asItemResult()
		if currentChild.visibleChildItemFields != nil && itemResult != nil {
			currentChild.Results = append(currentChild.Results, *itemResult)
		}
	}
	return result
}

package currentuser

import (
	"net/http"
	"time"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/structures"
)

// rawRootItem represents one row with a root activity/skill returned from the DB
type rawRootItem struct {
	// groups
	GroupID   int64
	GroupName string
	GroupType string

	// items
	ItemID                int64
	ItemType              string
	RequiresExplicitEntry bool
	EntryParticipantType  string
	NoScore               bool

	// title (from items_strings) in the user’s default language or (if not available) default language of the item
	Title       *string
	LanguageTag *string

	// results
	AttemptID        *int64
	ScoreComputed    float32
	Validated        bool
	StartedAt        *database.Time
	LatestActivityAt *database.Time
	EndedAt          *database.Time

	// attempts
	AttemptAllowsSubmissionsUntil database.Time

	// max from results of the current participant
	BestScore float32

	HasVisibleChildren bool

	*database.RawGeneratedPermissionFields
}

type groupInfoForRootItem struct {
	// required: true
	GroupID int64 `json:"group_id,string"`
	// required: true
	Name string `json:"name"`
	// required: true
	// enum: Class,Team,Club,Friends,Other,User,Session,Base,ContestParticipants
	Type string `json:"type"`
}

// swagger:model activitiesViewResponseRow
type activitiesViewResponseRow struct {
	*groupInfoForRootItem

	// required: true
	Activity *rootItem `json:"activity"`
}

type rootItem struct {
	*structures.ItemCommonFields

	// required: true
	RequiresExplicitEntry bool `json:"requires_explicit_entry"`
	// required: true
	// enum: User,Team
	EntryParticipantType string `json:"entry_participant_type"`
	// required: true
	NoScore bool `json:"no_score"`
	// required: true
	HasVisibleChildren bool `json:"has_visible_children"`
	// max among all attempts of the user (or of the team given in `{as_team_id}`)
	// required: true
	BestScore float32 `json:"best_score"`
	// required:true
	Results []structures.ItemResult `json:"results"`
}

// swagger:operation GET /current-user/group-memberships/activities group-memberships activitiesView
// ---
// summary: List root activities
// description:
//   Returns the list of root activities of the groups the current user (or `{as_team_id}`) belongs to.
//
//
//   If `{as_team_id}` is given, it should be a user's parent team group, otherwise the "forbidden" error is returned.
// parameters:
// - name: as_team_id
//   in: query
//   type: integer
// responses:
//   "200":
//     description: OK. Success response with an array of root activities
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/activitiesViewResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getRootActivities(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.getRootItems(w, r, true)
}

func (srv *Service) getRootItems(w http.ResponseWriter, r *http.Request, getActivities bool) service.APIError {
	user := srv.GetUser(r)

	participantID, apiError := service.GetParticipantIDFromRequest(r, user, srv.Store)
	if apiError != service.NoError {
		return apiError
	}

	rawData := srv.getRootItemsFromDB(participantID, user, getActivities)

	var activitiesResult []activitiesViewResponseRow
	var skillsResult []skillsViewResponseRow
	if getActivities {
		activitiesResult = make([]activitiesViewResponseRow, 0, len(rawData))
	} else {
		skillsResult = make([]skillsViewResponseRow, 0, len(rawData))
	}
	var currentItem *rootItem
	for index := range rawData {
		if index == 0 || rawData[index].GroupID != rawData[index-1].GroupID {
			currentItem = srv.generateRootItemInfoFromRawData(&rawData[index])
			if getActivities {
				row := activitiesViewResponseRow{
					groupInfoForRootItem: generateGroupInfoForRootItemFromRawData(rawData, index),
					Activity:             currentItem,
				}
				activitiesResult = append(activitiesResult, row)
			} else {
				row := skillsViewResponseRow{
					groupInfoForRootItem: generateGroupInfoForRootItemFromRawData(rawData, index),
					Skill:                currentItem,
				}
				skillsResult = append(skillsResult, row)
			}
		}

		if rawData[index].AttemptID != nil {
			currentItem.Results = append(currentItem.Results, generateItemResultFromRawData(&rawData[index]))
		}
	}

	if getActivities {
		render.Respond(w, r, activitiesResult)
	} else {
		render.Respond(w, r, skillsResult)
	}
	return service.NoError
}

func generateItemResultFromRawData(rawData *rawRootItem) structures.ItemResult {
	return structures.ItemResult{
		AttemptID:                     *rawData.AttemptID,
		ScoreComputed:                 rawData.ScoreComputed,
		Validated:                     rawData.Validated,
		StartedAt:                     (*time.Time)(rawData.StartedAt),
		LatestActivityAt:              (*time.Time)(rawData.LatestActivityAt),
		EndedAt:                       (*time.Time)(rawData.EndedAt),
		AttemptAllowsSubmissionsUntil: time.Time(rawData.AttemptAllowsSubmissionsUntil),
	}
}

func generateGroupInfoForRootItemFromRawData(rawData []rawRootItem, index int) *groupInfoForRootItem {
	return &groupInfoForRootItem{
		GroupID: rawData[index].GroupID,
		Name:    rawData[index].GroupName,
		Type:    rawData[index].GroupType,
	}
}

func (srv *Service) generateRootItemInfoFromRawData(rawData *rawRootItem) *rootItem {
	return &rootItem{
		ItemCommonFields: &structures.ItemCommonFields{
			ID:          rawData.ItemID,
			Type:        rawData.ItemType,
			String:      structures.ItemString{Title: rawData.Title, LanguageTag: rawData.LanguageTag},
			Permissions: *rawData.RawGeneratedPermissionFields.AsItemPermissions(srv.Store.PermissionsGranted()),
		},
		RequiresExplicitEntry: rawData.RequiresExplicitEntry,
		EntryParticipantType:  rawData.EntryParticipantType,
		NoScore:               rawData.NoScore,
		HasVisibleChildren:    rawData.HasVisibleChildren,
		BestScore:             rawData.BestScore,
		Results:               make([]structures.ItemResult, 0, 1),
	}
}

func (srv *Service) getRootItemsFromDB(participantID int64, user *database.User, selectActivities bool) []rawRootItem {
	hasVisibleChildrenQuery := srv.Store.Permissions().
		MatchingGroupAncestors(participantID).
		WherePermissionIsAtLeast("view", "info").
		Joins("JOIN items_items ON items_items.child_item_id = permissions.item_id").
		Where("items_items.parent_item_id = items.id").
		Select("1").Limit(1).SubQuery()

	itemsWithResultsQuery := srv.Store.ActiveGroupAncestors().
		Where("groups_ancestors_active.child_group_id = ?", participantID).
		Joins("JOIN `groups` ON groups.id = groups_ancestors_active.ancestor_group_id")
	if selectActivities {
		itemsWithResultsQuery = itemsWithResultsQuery.Joins("JOIN items ON items.id = groups.root_activity_id")
	} else {
		itemsWithResultsQuery = itemsWithResultsQuery.Joins("JOIN items ON items.id = groups.root_skill_id")
	}
	itemsWithResultsSubquery := itemsWithResultsQuery.
		JoinsPermissionsForGroupToItemsWherePermissionAtLeast(participantID, "view", "info").
		Joins("LEFT JOIN results ON results.participant_id = ? AND results.item_id = items.id", participantID).
		Joins("LEFT JOIN attempts ON attempts.participant_id = results.participant_id AND attempts.id = results.attempt_id").
		Select(`
			groups.id AS group_id, groups.name AS group_name, groups.type AS group_type, groups.created_at,
			items.id, items.type AS item_type,
			items.requires_explicit_entry, items.entry_participant_type, items.no_score, items.default_language_tag,
			IFNULL(
				(SELECT MAX(results.score_computed) AS best_score
				 FROM results
				 WHERE results.item_id = items.id AND results.participant_id = ?), 0) AS best_score,
			can_grant_view_generated_value, can_watch_generated_value, can_edit_generated_value, is_owner_generated, can_view_generated_value,
			attempts.allows_submissions_until AS attempt_allows_submissions_until,
			IFNULL(?, 0) AS has_visible_children,
			results.attempt_id,
			results.score_computed, results.validated, results.started_at, results.latest_activity_at,
			attempts.ended_at`, participantID, hasVisibleChildrenQuery).
		Group("groups.id, results.participant_id, results.attempt_id").SubQuery()

	query := srv.Store.Raw(`
		SELECT items.*, items.id AS item_id, COALESCE(user_strings.title, default_strings.title) AS title,
			IF(user_strings.title IS NOT NULL, user_strings.language_tag, default_strings.language_tag) AS language_tag
 		FROM ? AS items`, itemsWithResultsSubquery).
		JoinsUserAndDefaultItemStrings(user).
		Order("items.created_at, items.group_id, items.attempt_id")

	var rawData []rawRootItem
	service.MustNotBeError(query.Scan(&rawData).Error())
	return rawData
}

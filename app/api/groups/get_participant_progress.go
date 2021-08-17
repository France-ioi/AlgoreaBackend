package groups

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/structures"
)

// swagger:model groupParticipantProgressResponseRow
type groupParticipantProgressResponseRow struct {
	// required: true
	ItemID int64 `json:"item_id,string"`
	// required: true
	NoScore bool `json:"no_score"`
	// required: true
	// enum: Chapter,Task,Course,Skill
	Type string `json:"type"`

	String structures.ItemString `json:"string"`

	// required: true
	CurrentUserPermissions *structures.ItemPermissions `json:"current_user_permissions"`

	// The best score across all participant's or participant teams' results. If there are no results, the score is 0.
	// required:true
	Score float32 `json:"score"`
	// Whether the participant or one of his teams has the item validated
	// required:true
	Validated bool `json:"validated"`
	// Nullable
	// required:true
	LatestActivityAt *database.Time `json:"latest_activity_at"`
	// Number of hints requested for the result with the best score (if multiple, take the first one, chronologically).
	// If there are no results, the number of hints is 0.
	// required:true
	HintsRequested int32 `json:"hints_requested"`
	// Number of submissions for the result with the best score (if multiple, take the first one, chronologically).
	// If there are no results, the number of submissions is 0.
	// required:true
	Submissions int32 `json:"submissions"`
	// Time spent by the participant (or his teams) (in seconds):
	//
	//   1) if no results yet: 0
	//
	//   2) if one result validated: min(`validated_at`) - min(`started_at`)
	//     (i.e., time between the first time the participant (or one of his teams) started one (any) result
	//      and the time he (or one of his teams) first validated the task)
	//
	//   3) if no results validated: `now` - min(`started_at`)
	// required:true
	TimeSpent int32 `json:"time_spent"`
}

type rawParticipantProgressRow struct {
	// items
	ItemID  int64 `gorm:"column:id"`
	Type    string
	NoScore bool

	*database.RawGeneratedPermissionFields

	// from items_strings: in the user’s default language or (if not available) default language of the item
	StringLanguageTag string  `gorm:"column:language_tag"`
	StringTitle       *string `gorm:"column:title"`

	Score            float32
	Validated        bool
	LatestActivityAt *database.Time
	HintsRequested   int32
	Submissions      int32
	TimeSpent        int32
}

// swagger:operation GET /items/{item_id}/participant-progress groups groupParticipantProgress
// ---
// summary: Get progress of a participant
// description: >
//              Returns the current progress of a participant on a given item.
//
//
//              For all visible children of `{item_id}`,
//              displays the results of the given participant
//              (current user or `as_team_id` (if given) or `watched_group_id` (if given)).
//              Only one of `as_team_id` and `watched_group_id` can be given.
//              The results are sorted by `items_items.child_order`.
//
//
//              If the participant is a user, only the result corresponding to his best score counts
//              (across all his teams and his own results) disregarding whether or not
//              the score was done in a team.
//
//
//              Restrictions:
//
//              * The current user (or the team given in `as_team_id`) should have at least 'content' permissions on `{item_id}`,
//                otherwise the 'forbidden' response is returned.
//
//              * If `as_team_id` is given, it should be a user's parent team group,
//                otherwise the "forbidden" error is returned.
//
//              * If `{watched_group_id}` is given, the user should be a manager of the group with the 'can_watch_members' permission,
//                otherwise the "forbidden" error is returned.
//
//              * If `{watched_group_id}` is given, it should be a user group or a team,
//                otherwise the "forbidden" error is returned.
//
//              * If `{watched_group_id}` is given, the current user should have `can_watch` >= 'result' on the `{item_id}` item,
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
//   format: int64
// - name: watched_group_id
//   in: query
//   type: integer
//   format: int64
// responses:
//   "200":
//     description: OK. Success response with the participant's progress on item's children
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/groupParticipantProgressResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getParticipantProgress(w http.ResponseWriter, r *http.Request) service.APIError {
	itemID, participantID, checkPermissionsForGroupID, participantType, apiError := srv.parseParticipantProgressParameters(r)
	if apiError != service.NoError {
		return apiError
	}

	user := srv.GetUser(r)
	itemIDQuery := srv.Store.ItemItems().
		Select(`
			items.id, items.type, items.no_score, child_order, default_language_tag,
			IFNULL(user_permissions.can_view_generated_value, 1) AS can_view_generated_value,
			IFNULL(user_permissions.can_grant_view_generated_value, 1) AS can_grant_view_generated_value,
			IFNULL(user_permissions.can_watch_generated_value, 1) AS can_watch_generated_value,
			IFNULL(user_permissions.can_edit_generated_value, 1) AS can_edit_generated_value,
			IFNULL(user_permissions.is_owner_generated, 0) AS is_owner_generated`).
		Joins("JOIN items ON items.id = items_items.child_item_id").
		Where("parent_item_id = ?", itemID).
		JoinsPermissionsForGroupToItemsWherePermissionAtLeast(checkPermissionsForGroupID, "view", "info").
		Joins(
			"LEFT JOIN LATERAL ? AS user_permissions ON user_permissions.item_id = items.id",
			srv.Store.Permissions().AggregatedPermissionsForItems(user.GroupID).
				Where("permissions.item_id = items.id").SubQuery())

	var fieldVariables []interface{}
	var participantProgressQuery *database.DB
	fields := `
		items.id, items.type, items.no_score, items.default_language_tag,
		can_view_generated_value, can_grant_view_generated_value, can_watch_generated_value,
		can_edit_generated_value, is_owner_generated`
	if participantType == groupTypeUser {
		participantProgressQuery = srv.Store.Raw("WITH visible_items AS ? ?",
			itemIDQuery.SubQuery(),
			// nolint:gosec
			joinUserProgressResults(
				srv.Store.Raw(`
					SELECT STRAIGHT_JOIN`+fields+", "+userProgressFields+`
					FROM visible_items AS items`, fieldVariables...), participantID).
				Group("items.id").
				Order("MAX(items.child_order)").
				QueryExpr())
	} else {
		participantProgressQuery = srv.Store.Raw("WITH visible_items AS ? ?",
			itemIDQuery.SubQuery(),
			srv.Store.Table("visible_items AS items").
				Select(
					fields+", "+`
					IFNULL(result_with_best_score.score_computed, 0) AS score,
					IFNULL(result_with_best_score.validated, 0) AS validated,
					(SELECT MAX(latest_activity_at) FROM results WHERE participant_id = ? AND item_id = items.id) AS latest_activity_at,
					IFNULL(result_with_best_score.hints_cached, 0) AS hints_requested,
					IFNULL(result_with_best_score.submissions, 0) AS submissions,
					IF(result_with_best_score.participant_id IS NULL,
						0,
						(
							SELECT GREATEST(IF(result_with_best_score.validated,
								TIMESTAMPDIFF(SECOND, MIN(started_at), MIN(validated_at)),
								TIMESTAMPDIFF(SECOND, MIN(started_at), NOW())
							), 0)
							FROM results
							WHERE participant_id = ? AND item_id = items.id
						)
					) AS time_spent`, participantID, participantID).
				Joins(`
					LEFT JOIN LATERAL (
						SELECT score_computed, validated, hints_cached, submissions, participant_id
						FROM results
						WHERE participant_id = ? AND item_id = items.id
						ORDER BY participant_id, item_id, score_computed DESC, score_obtained_at
						LIMIT 1
					) AS result_with_best_score ON 1`, participantID).QueryExpr()).
			Order("items.child_order")
	}

	var result []rawParticipantProgressRow
	service.MustNotBeError(srv.Store.Raw(`
		SELECT items.*,
			COALESCE(user_strings.language_tag, default_strings.language_tag) AS language_tag,
			IF(user_strings.language_tag IS NULL, default_strings.title, user_strings.title) AS title
		FROM (?) AS items`, participantProgressQuery.SubQuery()).
		JoinsUserAndDefaultItemStrings(user).
		Scan(&result).Error())

	converted := make([]groupParticipantProgressResponseRow, 0, len(result))
	for i := range result {
		converted = append(converted, groupParticipantProgressResponseRow{
			ItemID:  result[i].ItemID,
			NoScore: result[i].NoScore,
			Type:    result[i].Type,
			String: structures.ItemString{
				Title:       result[i].StringTitle,
				LanguageTag: result[i].StringLanguageTag,
			},
			CurrentUserPermissions: result[i].AsItemPermissions(srv.Store.PermissionsGranted()),
			Score:                  result[i].Score,
			Validated:              result[i].Validated,
			LatestActivityAt:       result[i].LatestActivityAt,
			HintsRequested:         result[i].HintsRequested,
			Submissions:            result[i].Submissions,
			TimeSpent:              result[i].TimeSpent,
		})
	}
	render.Respond(w, r, converted)
	return service.NoError
}

func (srv *Service) parseParticipantProgressParameters(r *http.Request) (
	itemID, participantID, checkPermissionsForGroupID int64, participantType string, apiError service.APIError) {
	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return 0, 0, 0, "", service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)
	participantType = groupTypeUser
	participantID = service.ParticipantIDFromContext(r.Context())
	if participantID != user.GroupID {
		participantType = groupTypeTeam
	}
	checkPermissionsForGroupID = participantID

	watchedGroupID, watchedGroupIDSet, apiError := srv.ResolveWatchedGroupID(r)
	if apiError != service.NoError {
		return 0, 0, 0, "", apiError
	}

	if watchedGroupIDSet {
		if len(r.URL.Query()["as_team_id"]) != 0 {
			return 0, 0, 0, "", service.ErrInvalidRequest(errors.New("only one of as_team_id and watched_group_id can be given"))
		}

		participantID = watchedGroupID
		var found bool
		found, err = srv.Store.Permissions().MatchingUserAncestors(user).
			WherePermissionIsAtLeast("watch", "result").
			Where("item_id = ?", itemID).HasRows()
		service.MustNotBeError(err)
		if !found {
			return 0, 0, 0, "", service.InsufficientAccessRightsError
		}

		service.MustNotBeError(srv.Store.Groups().ByID(watchedGroupID).PluckFirst("type", &participantType).Error())
		if participantType != groupTypeUser && participantType != groupTypeTeam {
			return 0, 0, 0, "", service.ErrForbidden(errors.New("watched group should be a user or a team"))
		}
	}

	found, err := srv.Store.Permissions().MatchingGroupAncestors(checkPermissionsForGroupID).
		WherePermissionIsAtLeast("view", "content").
		Where("item_id = ?", itemID).HasRows()
	service.MustNotBeError(err)
	if !found {
		return 0, 0, 0, "", service.InsufficientAccessRightsError
	}

	return itemID, participantID, checkPermissionsForGroupID, participantType, service.NoError
}

package items

import (
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

// swagger:operation GET /attempts/{attempt_id}/task-token items itemTaskTokenGet
// ---
// summary: Get a task token
// description: >
//   Get a task token with the refreshed active attempt.
//
//
//   * If there is no row for the current user and the given item in `users_items`, the service creates one.
//
//   * If the active attempt (`active_attempt_id`) is not set in the `users_items` for the item and the user,
//   the service makes `attempt_id` active.
//
//   * Then `started_at` (if it is NULL) and `latest_activity_at` of `groups_attempts` & `user_items` are set to the current time.
//
//   * Finally, the service returns a task token with fresh data for the active attempt for the given item.
//
//
//   Restrictions:
//
//     * the `groups_attempts.group_id` should have at least 'content' access to the item,
//     * the item should be either 'Task' or 'Course',
//     * if `groups_attempts.group_id` != current user's `group_id`, it should be a team with `groups.team_item_id`
//       pointing to one of ancestors of `groups_attempts.item_id` or the `groups_attempts.item_id` itself,
//       and the current user should be a member of this team,
//
//   otherwise the 'forbidden' error is returned.
// parameters:
// - name: attempt_id
//   in: path
//   type: integer
//   required: true
// responses:
//   "200":
//     description: "OK. Success response with the fresh task token"
//     schema:
//       type: object
//       required: [success, message, data]
//       properties:
//         success:
//           description: "true"
//           type: boolean
//           enum: [true]
//         message:
//           description: updated
//           type: string
//           enum: [updated]
//         data:
//           type: object
//           required: [task_token]
//           properties:
//             task_token:
//               type: string
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getTaskToken(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error

	attemptID, err := service.ResolveURLQueryPathInt64Field(r, "attempt_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(r)

	var itemInfo struct {
		HasAttempts       bool
		AccessSolutions   bool
		HintsAllowed      bool
		TextID            *string
		URL               string
		SupportedLangProg *string
		ActiveAttemptID   *int64
	}

	var groupsAttemptInfo struct {
		ID               int64
		HintsRequested   *string
		HintsCachedCount int32 `gorm:"column:hints_cached"`
		GroupID          int64
		ItemID           int64
	}
	apiError := service.NoError
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		// load the attempt data
		err = store.GroupAttempts().ByID(attemptID).WithWriteLock().
			Select("id, hints_requested, hints_cached, group_id, item_id").Take(&groupsAttemptInfo).Error()

		if gorm.IsRecordNotFoundError(err) {
			apiError = service.InsufficientAccessRightsError
			return err // rollback
		}
		service.MustNotBeError(err)

		// if the attempt doesn't belong to the user, it should belong to the user's team related to the item
		if groupsAttemptInfo.GroupID != user.GroupID {
			var found bool
			found, err = store.Groups().TeamGroupForItemAndUser(groupsAttemptInfo.ItemID, user).
				Where("groups.id = ?", groupsAttemptInfo.GroupID).HasRows()
			service.MustNotBeError(err)
			if !found {
				apiError = service.InsufficientAccessRightsError
				return err // rollback
			}
		}

		// the attempt's group should have can_view >= 'content' permission on the item
		err = store.Items().ByID(groupsAttemptInfo.ItemID).
			WhereGroupHasViewPermissionOnItems(groupsAttemptInfo.GroupID, "content").
			Where("items.type IN('Task','Course')").
			Joins("LEFT JOIN users_items ON users_items.item_id = items.id AND users_items.user_id = ?", user.GroupID).
			Select(`
					can_view_generated_value = ? AS access_solutions,
					has_attempts, hints_allowed, text_id, url, supported_lang_prog,
					users_items.active_attempt_id`,
				store.PermissionsGranted().ViewIndexByName("solution")).
			Take(&itemInfo).Error()
		if gorm.IsRecordNotFoundError(err) {
			apiError = service.InsufficientAccessRightsError
			return apiError.Error // rollback
		}
		service.MustNotBeError(err)

		// update groups_attempts
		service.MustNotBeError(store.GroupAttempts().ByID(groupsAttemptInfo.ID).UpdateColumn(map[string]interface{}{
			"started_at":         gorm.Expr("IFNULL(started_at, ?)", database.Now()),
			"latest_activity_at": database.Now(),
		}).Error())

		// update users_items.active_attempt_id
		if itemInfo.ActiveAttemptID == nil {
			service.MustNotBeError(store.UserItems().SetActiveAttempt(user.GroupID, groupsAttemptInfo.ItemID, groupsAttemptInfo.ID))
		}

		return nil
	})
	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	taskToken := token.Task{
		AccessSolutions:    &itemInfo.AccessSolutions,
		SubmissionPossible: ptrBool(true),
		HintsAllowed:       &itemInfo.HintsAllowed,
		HintsRequested:     groupsAttemptInfo.HintsRequested,
		HintsGivenCount:    ptrString(strconv.Itoa(int(groupsAttemptInfo.HintsCachedCount))),
		IsAdmin:            ptrBool(false),
		ReadAnswers:        ptrBool(true),
		UserID:             strconv.FormatInt(user.GroupID, 10),
		LocalItemID:        strconv.FormatInt(groupsAttemptInfo.ItemID, 10),
		ItemID:             itemInfo.TextID,
		AttemptID:          strconv.FormatInt(groupsAttemptInfo.ID, 10),
		ItemURL:            itemInfo.URL,
		SupportedLangProg:  itemInfo.SupportedLangProg,
		RandomSeed:         strconv.FormatInt(groupsAttemptInfo.ID, 10),
		PlatformName:       srv.TokenConfig.PlatformName,
	}
	signedTaskToken, err := taskToken.Sign(srv.TokenConfig.PrivateKey)
	service.MustNotBeError(err)

	render.Respond(w, r, map[string]interface{}{
		"task_token": signedTaskToken,
	})
	return service.NoError
}

func ptrString(s string) *string { return &s }
func ptrBool(b bool) *bool       { return &b }

package items

import (
	"fmt"
	"hash/crc64"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/token"
)

// swagger:operation POST /items/{item_id}/participant/{participant_id}/attempts/{attempt_id}/generate-task-token items itemObsTaskTokenGen
//
//	---
//	summary: Generate a read-only task token for an observed user's attempt
//	description: >
//		Generate a read-only task token for another user's started attempt on a task,
//		so a task can read that user's editor history.
//
//		The token's `idAttempt` is `{participant_id}/{attempt_id}` and `idUser` / `sLogin` are the observed user.
//		`bSubmissionPossible` and `bHintsAllowed` are false.
//		`bAccessSolutions` is true when the caller has `can_view` >= 'solution' on the item,
//		or has a validated result on the item (any attempt).
//
//		This service does not update `results.latest_activity_at` and does not schedule result propagation.
//		Loading an attempt in order to read editor history is not that participant's activity.
//		The signed-in user's own token
//		(`POST /items/{item_id}/attempts/{attempt_id}/generate-task-token`) is unchanged and still refreshes activity.
//		`allows_submissions_until` is not checked: history stays readable after submissions are closed.
//		`as_team_id` is not accepted. A team participant is refused.
//
//		Restrictions (otherwise the 'forbidden' error is returned):
//
//			* the item is a task,
//			* `{participant_id}` is a user with a started result on the item for `{attempt_id}`,
//			* the caller has `can_watch` >= 'answer' on the item,
//			* the caller may watch that user (`can_watch_members` on an ancestor group of the user).
//
//		That is the same rule as generating a task token from an answer whose participant is watched.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: participant_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: attempt_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//	responses:
//		"200":
//			description: "OK. Success response with the read-only task token"
//			schema:
//				type: object
//				required: [success, message, data]
//				properties:
//					success:
//						description: "true"
//						type: boolean
//						enum: [true]
//					message:
//						description: created
//						type: string
//						enum: [created]
//					data:
//						type: object
//						required: [task_token]
//						properties:
//							task_token:
//								type: string
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
func (srv *Service) generateObservedUserTaskToken(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	participantID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "participant_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}
	attemptID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "attempt_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpRequest)
	info, err := loadObservedUserTaskTokenInfo(srv.GetStore(httpRequest), user, itemID, participantID, attemptID)
	if gorm.IsRecordNotFoundError(err) {
		return service.ErrAPIInsufficientAccessRights
	}
	service.MustNotBeError(err)

	fullAttemptID := fmt.Sprintf("%d/%d", participantID, attemptID)
	randomSeed := crc64.Checksum([]byte(fullAttemptID), crc64.MakeTable(crc64.ECMA))

	taskToken := &token.Token[payloads.TaskToken]{Payload: payloads.TaskToken{
		AccessSolutions:    &info.AccessSolutions,
		SubmissionPossible: new(false),
		HintsAllowed:       new(false),
		HintsRequested:     info.HintsRequested,
		HintsGivenCount:    new(strconv.Itoa(int(info.HintsCachedCount))),
		IsAdmin:            new(false),
		ReadAnswers:        new(true),
		UserID:             strconv.FormatInt(participantID, 10),
		LocalItemID:        strconv.FormatInt(itemID, 10),
		ItemID:             info.TextID,
		AttemptID:          fullAttemptID,
		ItemURL:            info.URL,
		SupportedLangProg:  info.SupportedLangProg,
		RandomSeed:         strconv.FormatUint(randomSeed, 10),
		PlatformName:       srv.TokenConfig.PlatformName,
		Login:              &info.Login,
	}}
	signedTaskToken, err := taskToken.Sign(srv.TokenConfig.PrivateKey)
	service.MustNotBeError(err)

	render.Respond(responseWriter, httpRequest, service.CreationSuccess(map[string]any{
		"task_token": signedTaskToken,
	}))
	return nil
}

type observedUserTaskTokenInfo struct {
	AccessSolutions   bool
	TextID            *string
	URL               string
	SupportedLangProg *string
	Login             string
	HintsRequested    *string
	HintsCachedCount  int32 `gorm:"column:hints_cached"`
}

// loadObservedUserTaskTokenInfo returns the task and result fields for a read-only token.
// It returns a not-found error when the caller may not watch this user's answers on this attempt.
func loadObservedUserTaskTokenInfo(
	store *database.DataStore, user *database.User, itemID, participantID, attemptID int64,
) (*observedUserTaskTokenInfo, error) {
	// Same observer rule as POST /answers/{answer_id}/generate-task-token for a watched participant.
	canWatchAnswer := store.Permissions().MatchingUserAncestors(user).
		WherePermissionIsAtLeast("watch", "answer").
		Where("permissions.item_id = results.item_id").
		Select("1").Limit(1)
	canWatchParticipant := store.ActiveGroupAncestors().ManagedByUser(user).
		Where("groups_ancestors_active.child_group_id = results.participant_id").
		Where("can_watch_members").
		Select("1").Limit(1)
	callerCanViewSolution := store.Permissions().MatchingUserAncestors(user).
		WherePermissionIsAtLeast("view", "solution").
		Where("permissions.item_id = results.item_id").
		Select("1").Limit(1)
	// Bound to itemID so the inner `results` is not the outer attempt row.
	callerHasValidated := store.Results().
		Where("results.participant_id = ?", user.GroupID).
		Where("results.item_id = ?", itemID).
		Where("results.validated").
		Select("1").Limit(1)

	var info observedUserTaskTokenInfo
	err := store.Results().
		Joins("JOIN items ON items.id = results.item_id").
		Joins("JOIN users ON users.group_id = results.participant_id").
		Where("results.participant_id = ?", participantID).
		Where("results.attempt_id = ?", attemptID).
		Where("results.item_id = ?", itemID).
		Where("results.started").
		Where("items.type = 'Task'").
		Where("EXISTS (?)", canWatchAnswer.SubQuery()).
		Where("EXISTS (?)", canWatchParticipant.SubQuery()).
		Select(`
			(EXISTS (?) OR EXISTS (?)) AS access_solutions,
			items.text_id AS text_id,
			items.url AS url,
			items.supported_lang_prog AS supported_lang_prog,
			users.login AS login,
			results.hints_requested AS hints_requested,
			results.hints_cached AS hints_cached`,
			callerCanViewSolution.SubQuery(), callerHasValidated.SubQuery()).
		Take(&info).Error()
	if err != nil {
		return nil, err
	}
	return &info, nil
}

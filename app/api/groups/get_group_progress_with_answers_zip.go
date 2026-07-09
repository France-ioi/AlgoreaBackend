package groups

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

const (
	maxUsersInProgressZIP = 100
	maxItemsInProgressZIP = 100
)

type progressZIPSubtreeItem struct {
	ItemID  int64
	DirPath string
}

type progressZIPUser struct {
	GroupID        int64
	Login          string
	SanitizedLogin string
}

type progressZIPSubmission struct {
	ParticipantID int64
	ItemID        int64
	Number        int
	AttemptID     int64
	Answer        string
	Score         *float32
	AnswerID      int64
}

type progressZIPUserItemData struct {
	HintsRequested   int32          `json:"hints_requested"`
	LatestActivityAt *database.Time `json:"latest_activity_at"`
	Score            float32        `json:"score"`
	Submissions      int32          `json:"submissions"`
	TimeSpent        int32          `json:"time_spent"`
	Validated        bool           `json:"validated"`
}

type progressZIPItemChild struct {
	ChildItemID int64
	ChildOrder  int32
}

// swagger:operation GET /groups/{group_id}/group-progress-with-answers-zip groups groupGroupProgressWithAnswersZIP
//
//	---
//	summary: Get group progress with answers as a ZIP file
//	description: >
//		Returns the current progress of users on a subset of items with their submission answers.
//
//		Content of the ZIP file:
//
//		* `group_progress.csv` at the root, with the same content as `groupGroupProgressCSV`
//
//		* directories reflecting the visible items subtree, with directory names in the format
//		  `{child_order}-{title}-{id}/` where `title` is sanitized and root parent items use `child_order` 0.
//		  Descendants are nested at arbitrary depth under their parent directory.
//
//		* each item directory contains a `submissions/` subdirectory with one directory per user login.
//		  Each user directory contains:
//
//		  - `data.json` with `hints_requested`, `latest_activity_at`, `score`, `submissions`, `time_spent`, `validated`
//
//		  - for each Submission answer of the user on the item, a file named
//		    `{number}-{attempt_id}-{score}-{answers.id}.txt` where `number` is the 0-based order by `created_at`
//		    and the file content is the `answer` field
//
//		Restrictions:
//
//		* The current user should be a manager of the group (or of one of its ancestors)
//		  with `can_watch_members` set to true,
//
//		* The current user should have `can_watch` >= 'answer' on each of `{parent_item_ids}` items,
//
//		* The export is limited to 100 users and 100 items in the visible descendant subtree.
//		  If either limit is exceeded, a distinct 400 error is returned.
//
//		Otherwise the 'forbidden' error is returned.
//	parameters:
//		- name: group_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: parent_item_ids
//			in: query
//			type: array
//			required: true
//			items:
//				type: integer
//				format: int64
//	responses:
//		"200":
//			description: OK. Success response with users progress and answers on items
//			content:
//				application/zip:
//					schema:
//						type: string
//						format: binary
//		"400":
//			description: >
//				Invalid request. This includes exceeding export limits:
//				"The number of items exceeds the limit (100)" or
//				"The number of users exceeds the limit (100)".
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getGroupProgressWithAnswersZIP(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	groupID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if !user.CanWatchGroupMembers(store, groupID) {
		return service.ErrAPIInsufficientAccessRights
	}

	itemParentIDs, err := resolveAndCheckParentIDs(store, httpRequest, user, "answer")
	service.MustNotBeError(err)

	var subtreeItems []progressZIPSubtreeItem
	var zipUsers []progressZIPUser
	if len(itemParentIDs) > 0 {
		var itemCount int
		subtreeItems, itemCount = buildProgressZIPVisibleSubtree(store, user, itemParentIDs)
		if itemCount > maxItemsInProgressZIP {
			return service.ErrInvalidRequest(errors.New("The number of items exceeds the limit (100)")) //nolint:staticcheck // API-specified message
		}

		zipUsers = getProgressZIPUsers(store, groupID)
		if len(zipUsers) > maxUsersInProgressZIP {
			return service.ErrInvalidRequest(errors.New("The number of users exceeds the limit (100)")) //nolint:staticcheck // API-specified message
		}
	}

	responseWriter.Header().Set("Content-Type", "application/zip")
	itemParentIDsString := make([]string, len(itemParentIDs))
	for i, id := range itemParentIDs {
		itemParentIDsString[i] = strconv.FormatInt(id, 10)
	}
	responseWriter.Header().Set(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=groups_progress_with_answers_for_group-%d-and_child_items_of-%s.zip",
			groupID, strings.Join(itemParentIDsString, "-")),
	)

	// Buffer the ZIP in memory so a mid-generation failure (raised as a panic via
	// service.MustNotBeError and recovered by AppHandler) yields a clean error response
	// instead of a 200 with a truncated archive.
	zipBuffer := &bytes.Buffer{}
	zipWriter := zip.NewWriter(zipBuffer)
	writeProgressZIPArchive(zipWriter, store, user, itemParentIDs, groupID, subtreeItems, zipUsers)
	service.MustNotBeError(zipWriter.Close())

	_, err = io.Copy(responseWriter, zipBuffer)
	service.MustNotBeError(err)
	return nil
}

func writeProgressZIPArchive(
	zipWriter *zip.Writer, store *database.DataStore, user *database.User,
	itemParentIDs []int64, groupID int64, subtreeItems []progressZIPSubtreeItem, zipUsers []progressZIPUser,
) {
	csvBuffer := &bytes.Buffer{}
	csvWriter := csv.NewWriter(csvBuffer)
	csvWriter.Comma = ';'
	if len(itemParentIDs) == 0 {
		// No error check needed: WriteString on a bytes.Buffer never fails.
		_, _ = csvBuffer.WriteString("Group name\n")
	} else {
		writeGroupProgressCSV(csvWriter, store, user, itemParentIDs, groupID)
		csvWriter.Flush()
	}

	// The ZIP is written to an in-memory buffer, so these writes cannot fail in practice;
	// MustNotBeError turns any unexpected failure into a recovered 500 rather than a partial 200.
	groupProgressFile, err := zipWriter.Create("group_progress.csv")
	service.MustNotBeError(err)
	_, err = io.Copy(groupProgressFile, csvBuffer)
	service.MustNotBeError(err)

	if len(itemParentIDs) == 0 {
		return
	}

	writeProgressZIPContent(zipWriter, store, subtreeItems, zipUsers)
}

func writeProgressZIPContent(
	zipWriter *zip.Writer, store *database.DataStore,
	subtreeItems []progressZIPSubtreeItem, zipUsers []progressZIPUser,
) {
	if len(zipUsers) == 0 {
		return
	}

	subtreeItemIDs := make([]int64, len(subtreeItems))
	for itemIndex := range subtreeItems {
		subtreeItemIDs[itemIndex] = subtreeItems[itemIndex].ItemID
	}

	userProgressByGroupAndItem := getProgressZIPUserItemData(store, zipUsers, subtreeItemIDs)
	submissionsByParticipantAndItem := getProgressZIPSubmissions(store, zipUsers, subtreeItemIDs)

	for _, subtreeItem := range subtreeItems {
		for _, zipUser := range zipUsers {
			writeProgressZIPUserItemFiles(
				zipWriter, subtreeItem, zipUser, userProgressByGroupAndItem, submissionsByParticipantAndItem,
			)
		}
	}
}

func writeProgressZIPUserItemFiles(
	zipWriter *zip.Writer, subtreeItem progressZIPSubtreeItem, zipUser progressZIPUser,
	userProgressByGroupAndItem map[int64]map[int64]progressZIPUserItemData,
	submissionsByParticipantAndItem map[int64]map[int64][]progressZIPSubmission,
) {
	// The ZIP is written to an in-memory buffer, so these Create/Write/Marshal calls cannot
	// fail in practice; MustNotBeError guards against unexpected failures without dead branches.
	dataPath := subtreeItem.DirPath + "submissions/" + zipUser.SanitizedLogin + "/data.json"
	dataFile, err := zipWriter.Create(dataPath)
	service.MustNotBeError(err)

	progressData := userProgressByGroupAndItem[zipUser.GroupID][subtreeItem.ItemID]
	encodedData, err := json.Marshal(progressData)
	service.MustNotBeError(err)
	_, err = dataFile.Write(encodedData)
	service.MustNotBeError(err)

	for _, submission := range submissionsByParticipantAndItem[zipUser.GroupID][subtreeItem.ItemID] {
		scoreString := ""
		if submission.Score != nil {
			scoreString = fmt.Sprintf("%v", *submission.Score)
		}
		filePath := fmt.Sprintf("%ssubmissions/%s/%d-%d-%s-%d.txt",
			subtreeItem.DirPath, zipUser.SanitizedLogin,
			submission.Number, submission.AttemptID, scoreString, submission.AnswerID)
		answerFile, err := zipWriter.Create(filePath)
		service.MustNotBeError(err)
		_, err = answerFile.Write([]byte(submission.Answer))
		service.MustNotBeError(err)
	}
}

func buildProgressZIPVisibleSubtree(
	store *database.DataStore, user *database.User, itemParentIDs []int64,
) (subtreeItems []progressZIPSubtreeItem, itemCount int) {
	permissionsSubQuery := store.Permissions().MatchingUserAncestors(user).
		Select("item_id").
		WherePermissionIsAtLeast("view", "info").SubQuery()

	var childrenByParent map[int64][]progressZIPItemChild
	childrenByParent, itemCount = loadProgressZIPChildrenByParent(store, itemParentIDs, permissionsSubQuery)
	if itemCount > maxItemsInProgressZIP {
		return nil, itemCount
	}

	itemTitles := getProgressZIPItemTitles(store, user, itemParentIDs, childrenByParent)

	visited := make(map[int64]bool)
	subtreeItems = make([]progressZIPSubtreeItem, 0, itemCount)
	for _, parentItemID := range itemParentIDs {
		appendProgressZIPSubtreeItem(
			parentItemID, 0, "", itemTitles, childrenByParent, visited, &subtreeItems,
		)
	}

	return subtreeItems, itemCount
}

func loadProgressZIPChildrenByParent(
	store *database.DataStore, itemParentIDs []int64, permissionsSubQuery interface{},
) (childrenByParent map[int64][]progressZIPItemChild, itemCount int) {
	childrenByParent = make(map[int64][]progressZIPItemChild)
	parentFrontier := append([]int64(nil), itemParentIDs...)
	knownParents := make(map[int64]bool, len(itemParentIDs))
	for _, parentID := range itemParentIDs {
		knownParents[parentID] = true
	}

	for len(parentFrontier) > 0 {
		var childRelations []struct {
			ParentItemID int64
			ChildItemID  int64
			ChildOrder   int32
		}
		service.MustNotBeError(store.ItemItems().
			Select("items_items.parent_item_id, items_items.child_item_id, items_items.child_order").
			Where("items_items.parent_item_id IN (?)", parentFrontier).
			Joins("JOIN ? AS permissions ON permissions.item_id = items_items.child_item_id", permissionsSubQuery).
			Order("items_items.parent_item_id, items_items.child_order, items_items.child_item_id").
			Scan(&childRelations).Error())

		nextFrontier := make([]int64, 0, len(childRelations))
		for i := range childRelations {
			parentID := childRelations[i].ParentItemID
			childID := childRelations[i].ChildItemID
			childrenByParent[parentID] = append(childrenByParent[parentID], progressZIPItemChild{
				ChildItemID: childID,
				ChildOrder:  childRelations[i].ChildOrder,
			})
			if !knownParents[childID] {
				knownParents[childID] = true
				itemCount = len(knownParents)
				if itemCount > maxItemsInProgressZIP {
					return childrenByParent, itemCount
				}
				nextFrontier = append(nextFrontier, childID)
			}
		}
		parentFrontier = nextFrontier
	}

	return childrenByParent, len(knownParents)
}

func getProgressZIPItemTitles(
	store *database.DataStore, user *database.User, itemParentIDs []int64, childrenByParent map[int64][]progressZIPItemChild,
) map[int64]string {
	itemIDSet := make(map[int64]bool, len(itemParentIDs))
	for _, parentItemID := range itemParentIDs {
		itemIDSet[parentItemID] = true
	}
	for parentID := range childrenByParent {
		for _, child := range childrenByParent[parentID] {
			itemIDSet[child.ChildItemID] = true
		}
	}

	itemIDs := make([]int64, 0, len(itemIDSet))
	for itemID := range itemIDSet {
		itemIDs = append(itemIDs, itemID)
	}

	var items []struct {
		ID    int64
		Title string
	}
	service.MustNotBeError(store.Items().
		JoinsUserAndDefaultItemStrings(user).
		Where("items.id IN (?)", itemIDs).
		Select("items.id, COALESCE(user_strings.title, default_strings.title) AS title").
		Scan(&items).Error())

	itemTitles := make(map[int64]string, len(items))
	for i := range items {
		itemTitles[items[i].ID] = items[i].Title
	}
	return itemTitles
}

func appendProgressZIPSubtreeItem(
	itemID int64, childOrder int32, parentPath string, itemTitles map[int64]string,
	childrenByParent map[int64][]progressZIPItemChild, visited map[int64]bool, subtreeItems *[]progressZIPSubtreeItem,
) {
	if visited[itemID] {
		return
	}
	visited[itemID] = true

	dirPath := parentPath + fmt.Sprintf("%d-%s-%d/", childOrder, sanitizeZIPPathSegment(itemTitles[itemID], "untitled"), itemID)
	*subtreeItems = append(*subtreeItems, progressZIPSubtreeItem{
		ItemID:  itemID,
		DirPath: dirPath,
	})

	for _, child := range childrenByParent[itemID] {
		appendProgressZIPSubtreeItem(
			child.ChildItemID, child.ChildOrder, dirPath, itemTitles, childrenByParent, visited, subtreeItems,
		)
	}
}

func sanitizeZIPPathSegment(segment, emptyFallback string) string {
	segment = strings.ReplaceAll(segment, "/", "-")
	segment = strings.ReplaceAll(segment, "\\", "-")
	for strings.Contains(segment, "..") {
		segment = strings.ReplaceAll(segment, "..", "-")
	}
	segment = strings.Trim(segment, ". ")
	if segment == "" {
		return emptyFallback
	}
	return segment
}

func getProgressZIPUsers(store *database.DataStore, groupID int64) []progressZIPUser {
	var users []progressZIPUser
	service.MustNotBeError(store.Groups().
		Select("groups.id AS group_id, users.login").
		Joins("JOIN groups_groups_active ON groups_groups_active.child_group_id = groups.id").
		Joins(`
			JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups_groups_active.parent_group_id AND
				groups_ancestors_active.ancestor_group_id = ?`, groupID).
		Joins("JOIN users ON users.group_id = groups.id").
		Where("groups.type = 'User'").
		Group("groups.id").
		Order("groups.name, groups.id").
		Scan(&users).Error())
	for i := range users {
		users[i].SanitizedLogin = sanitizeZIPPathSegment(users[i].Login, "user")
	}
	return users
}

func getProgressZIPUserItemData(
	store *database.DataStore, zipUsers []progressZIPUser, subtreeItemIDs []int64,
) map[int64]map[int64]progressZIPUserItemData {
	defaultData := progressZIPUserItemData{
		HintsRequested: 0,
		Score:          0,
		Submissions:    0,
		TimeSpent:      0,
		Validated:      false,
	}

	result := make(map[int64]map[int64]progressZIPUserItemData, len(zipUsers))
	for _, zipUser := range zipUsers {
		result[zipUser.GroupID] = make(map[int64]progressZIPUserItemData, len(subtreeItemIDs))
		for _, itemID := range subtreeItemIDs {
			result[zipUser.GroupID][itemID] = defaultData
		}
	}

	userIDs := make([]string, len(zipUsers))
	for i := range zipUsers {
		userIDs[i] = strconv.FormatInt(zipUsers[i].GroupID, 10)
	}
	itemIDs := make([]string, len(subtreeItemIDs))
	for i, itemID := range subtreeItemIDs {
		itemIDs[i] = strconv.FormatInt(itemID, 10)
	}
	userIDsList := strings.Join(userIDs, ", ")
	itemsSubQuery := gorm.Expr(`JSON_TABLE('[` + strings.Join(itemIDs, ", ") + `]', "$[*]" COLUMNS(id BIGINT PATH "$"))`)

	var progressRows []struct {
		GroupID          int64
		ItemID           int64
		Score            float32
		Validated        bool
		LatestActivityAt *database.Time
		HintsRequested   int32
		Submissions      int32
		TimeSpent        int32
	}
	service.MustNotBeError(
		joinUserOnlyProgressResults(
			store.Raw(`
				SELECT STRAIGHT_JOIN
					items.id AS item_id,
					users.group_id AS group_id, `+userProgressFields+`
				FROM JSON_TABLE('[`+userIDsList+`]', "$[*]" COLUMNS(group_id BIGINT PATH "$")) AS users`).
				Joins("JOIN ? AS items", itemsSubQuery),
			gorm.Expr("users.group_id"),
		).
			Group("users.group_id, items.id").
			Scan(&progressRows).Error())

	for rowIndex := range progressRows {
		result[progressRows[rowIndex].GroupID][progressRows[rowIndex].ItemID] = progressZIPUserItemData{
			HintsRequested:   progressRows[rowIndex].HintsRequested,
			LatestActivityAt: progressRows[rowIndex].LatestActivityAt,
			Score:            progressRows[rowIndex].Score,
			Submissions:      progressRows[rowIndex].Submissions,
			TimeSpent:        progressRows[rowIndex].TimeSpent,
			Validated:        progressRows[rowIndex].Validated,
		}
	}

	return result
}

func joinUserOnlyProgressResults(db *database.DB, userID interface{}) *database.DB {
	// Deliberately omits team-result merging from joinUserProgressResults: this export
	// scopes participants to user self-groups only (no teams), per product decision.
	return db.
		Joins(`
			LEFT JOIN LATERAL (
				SELECT participant_id, attempt_id, score_computed, score_obtained_at
				FROM results AS result_with_best_score_for_user
				WHERE participant_id = ? AND item_id = items.id
				ORDER BY participant_id, item_id, score_computed DESC, score_obtained_at
				LIMIT 1
			) AS result_with_best_score_for_user ON 1`, userID).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT participant_id, attempt_id, latest_activity_at
				FROM results AS last_result_of_user USE INDEX (participant_id_item_id_latest_activity_at_desc)
				WHERE participant_id = ? AND item_id = items.id AND latest_activity_at IS NOT NULL
				ORDER BY participant_id, item_id, latest_activity_at DESC LIMIT 1
			) AS last_result_of_user ON 1`, userID).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT last_result_of_user.latest_activity_at AS latest_activity_at
			) AS last_result ON 1`).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT participant_id, attempt_id, started_at
				FROM results AS first_result_of_user USE INDEX (participant_id_item_id_started_started_at)
				WHERE participant_id = ? AND item_id = items.id AND started = 1
				ORDER BY participant_id, item_id, started, started_at LIMIT 1
			) AS first_result_of_user ON 1`, userID).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT first_result_of_user.started_at AS started_at
			) AS first_result ON 1`).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT participant_id, attempt_id, validated_at
				FROM results AS first_validated_result_of_user USE INDEX (participant_id_item_id_validated_validated_at)
				WHERE participant_id = ? AND item_id = items.id AND validated = 1
				ORDER BY participant_id, item_id, validated, validated_at LIMIT 1
			) AS first_validated_result_of_user ON 1`, userID).
		Joins(`
			LEFT JOIN LATERAL (
				SELECT first_validated_result_of_user.validated_at AS validated_at
			) AS first_validated_result ON 1`).
		Joins(`
			LEFT JOIN results AS result_with_best_score
			ON result_with_best_score.participant_id = result_with_best_score_for_user.participant_id
			AND result_with_best_score.attempt_id = result_with_best_score_for_user.attempt_id
			AND result_with_best_score.item_id = items.id`)
}

func getProgressZIPSubmissions(
	store *database.DataStore, zipUsers []progressZIPUser, subtreeItemIDs []int64,
) map[int64]map[int64][]progressZIPSubmission {
	result := make(map[int64]map[int64][]progressZIPSubmission, len(zipUsers))
	participantIDs := make([]int64, len(zipUsers))
	for i := range zipUsers {
		participantIDs[i] = zipUsers[i].GroupID
		result[zipUsers[i].GroupID] = make(map[int64][]progressZIPSubmission, len(subtreeItemIDs))
	}

	var answers []struct {
		ParticipantID int64
		ItemID        int64
		ID            int64
		AttemptID     int64
		Answer        string
		Score         *float32
	}
	service.MustNotBeError(store.Answers().
		WithGradings().
		Where("answers.type = 'Submission'").
		Where("answers.participant_id IN (?)", participantIDs).
		Where("answers.item_id IN (?)", subtreeItemIDs).
		Order("answers.participant_id, answers.item_id, answers.created_at").
		Select("answers.participant_id, answers.item_id, answers.id, answers.attempt_id, answers.answer, gradings.score").
		Scan(&answers).Error())

	answerNumbers := make(map[int64]map[int64]int, len(zipUsers))
	for answerIndex := range answers {
		participantID := answers[answerIndex].ParticipantID
		itemID := answers[answerIndex].ItemID
		if answerNumbers[participantID] == nil {
			answerNumbers[participantID] = make(map[int64]int)
		}
		submissionNumber := answerNumbers[participantID][itemID]
		answerNumbers[participantID][itemID]++

		result[participantID][itemID] = append(result[participantID][itemID], progressZIPSubmission{
			ParticipantID: participantID,
			ItemID:        itemID,
			Number:        submissionNumber,
			AttemptID:     answers[answerIndex].AttemptID,
			Answer:        answers[answerIndex].Answer,
			Score:         answers[answerIndex].Score,
			AnswerID:      answers[answerIndex].ID,
		})
	}

	return result
}

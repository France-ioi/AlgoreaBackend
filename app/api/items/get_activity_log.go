package items

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model itemActivityLogResponseRow
type itemActivityLogResponseRow struct {
	// required: true
	At *database.Time `json:"at"`
	// required: true
	// enum: result_started,submission,result_validated
	ActivityType string `json:"activity_type"`
	// required: true
	AttemptID int64 `json:"attempt_id,string"`
	// `answers.id`
	AnswerID *int64 `json:"answer_id,string,omitempty"`
	// use this as `{from.asnwer_id}` for pagination
	// required: true
	FromAnswerID int64    `json:"from_answer_id,string"`
	Score        *float32 `json:"score,omitempty"`
	// required: true
	Participant struct {
		// required: true
		ID int64 `json:"id,string"`
		// required: true
		Name string `json:"name"`
		// required: true
		// enum: Team,User
		Type string `json:"type"`
	} `json:"participant" gorm:"embedded;embedded_prefix:participant__"`
	User *struct {
		// required: true
		ID *int64 `json:"id,string"`
		// required: true
		Login string `json:"login"`
		// Nullable
		// required: true
		FirstName *string `json:"first_name"`
		// Nullable
		// required: true
		LastName *string `json:"last_name"`
	} `json:"user,omitempty" gorm:"embedded;embedded_prefix:user__"`
	// required: true
	Item struct {
		// required: true
		ID int64 `json:"id,string"`
		// required: true
		// enum: Chapter,Task,Course
		Type string `json:"type"`
		// required: true
		String struct {
			// Nullable
			// required: true
			Title *string `json:"title"`
		} `json:"string" gorm:"embedded;embedded_prefix:string__"`
	} `json:"item" gorm:"embedded;embedded_prefix:item__"`
}

// swagger:operation GET /items/{item_id}/log items itemActivityLog
// ---
// summary: Activity log on an item
// description: >
//   Returns rows from `answers` having `type` = "Submission" and started/validated `results`
//   with additional info on users and items for the participant or the `{watched_group_id}` group
//   (only one of `{as_team_id}` and `{watched_group_id}` can be given).
//
//
//   If possible, items titles are shown in the authenticated user's default language.
//   Otherwise, the item's default language is used.
//
//
//   If `{watched_group_id}` is given, all rows of the result are related to descendant groups of `{watched_group_id}`
//   and items that are descendants of `{item_id}` (+ `{item_id}` itself) and visible to the current user
//   (at least 'info' access with `can_watch` >= 'result').
//
//
//   If `{watched_group_id}` is not given, all rows of the result are related to the participant group (the current user or `{as_team_id}`)
//   and items that are descendants of `{item_id}` (+ `{item_id}` itself) and visible to the current user (at least 'info' access).
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   required: true
// - name: as_team_id
//   in: query
//   type: integer
// - name: watched_group_id
//   description: The current user should be a manager of the watched group with `can_watch_members` = true,
//                otherwise the 'forbidden' error is returned
//   in: query
//   type: integer
// - name: from.at
//   description: Start the page from the row next to the row with `at` = `{from.at}`
//                (all other `from.*` parameters are required when `from.at` is present)
//   in: query
//   type: string
// - name: from.item_id
//   description: Start the page from the row next to the row with `item_id`=`{from.item_id}`
//                (all other `from.*` parameters are required when `from.item_id` is present)
//   in: query
//   type: integer
// - name: from.participant_id
//   description: Start the page from the row next to the row with `participant_id`=`{from.participant_id}`
//                (all other `from.*` parameters are required when `from.participant_id` is present)
//   in: query
//   type: integer
// - name: from.attempt_id
//   description: Start the page from the row next to the row with `attempt_id`=`{from.attempt_id}`
//                (all other `from.*` parameters are required when `from.attempt_id` is present)
//   in: query
//   type: integer
// - name: from.answer_id
//   description: Start the page from the row next to the row with `from_answer_id`=`{from.answer_id}`
//                (all other `from.*` parameters are required when `from.answer_id` is present)
//   in: query
//   type: integer
// - name: from.activity_type
//   description: Start the page from the row next to the row with `activity_type`=`{from.activity_type}`
//                (all other `from.*` parameters are required when `from.activity_type` is present)
//   in: query
//   type: string
//   enum: [result_started,submission,result_validated]
// - name: limit
//   description: Display the first N rows
//   in: query
//   type: integer
//   maximum: 1000
//   default: 500
// responses:
//   "200":
//     description: OK. The array of users answers
//     schema:
//       type: array
//       items:
//         "$ref": "#/definitions/itemActivityLogResponseRow"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getActivityLog(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	// check and patch from.activity_type to make it integer
	urlParams := r.URL.Query()
	if len(urlParams["from.activity_type"]) > 0 {
		stringValue := r.URL.Query().Get("from.activity_type")
		var intValue int
		var ok bool
		if intValue, ok = map[string]int{"result_started": 1, "submission": 2, "result_validated": 3}[stringValue]; !ok {
			return service.ErrInvalidRequest(
				errors.New("wrong value for from.activity_type (should be one of (result_started, submission, result_validated))"))
		}
		urlParams["from.activity_type"] = []string{strconv.Itoa(intValue)}
		r.URL.RawQuery = urlParams.Encode()
	}

	query, apiError := srv.constructActivityLogQuery(r, itemID, user)
	if apiError != service.NoError {
		return apiError
	}

	var result []itemActivityLogResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	fromAnswerID := int64(-1)
	if len(urlParams["from.answer_id"]) > 0 {
		// the error checking has been already done in constructActivityLogQuery()
		fromAnswerID, _ = service.ResolveURLQueryGetInt64Field(r, "from.answer_id")
	}
	for index := range result {
		if *result[index].AnswerID != -1 {
			result[index].FromAnswerID = *result[index].AnswerID
		} else {
			result[index].FromAnswerID = fromAnswerID
			result[index].AnswerID = nil
		}
		fromAnswerID = result[index].FromAnswerID
		if result[index].User.ID == nil {
			result[index].User = nil
		}
	}

	render.Respond(w, r, result)
	return service.NoError
}

func (srv *Service) constructActivityLogQuery(r *http.Request, itemID int64, user *database.User) (*database.DB, service.APIError) {
	participantID := service.ParticipantIDFromContext(r.Context())
	watchedGroupID, watchedGroupIDSet, apiError := srv.resolveWatchedGroupID(r)
	if apiError != service.NoError {
		return nil, apiError
	}
	participantsQuery := srv.Store.Raw("SELECT ? AS id", participantID)
	if watchedGroupIDSet {
		if len(r.URL.Query()["as_team_id"]) != 0 {
			return nil, service.ErrInvalidRequest(errors.New("only one of as_team_id and watched_group_id can be given"))
		}
		participantsQuery = srv.Store.ActiveGroupAncestors().Where("ancestor_group_id = ?", watchedGroupID).
			Select("child_group_id AS id")
	}

	itemDescendants := srv.Store.ItemAncestors().DescendantsOf(itemID).Select("child_item_id")
	visibleItemDescendants := srv.Store.Permissions().MatchingGroupAncestors(user.GroupID).
		Select("item_id AS id").
		Where("item_id = ? OR item_id IN ?", itemID, itemDescendants.SubQuery()).
		Group("item_id").
		HavingMaxPermissionAtLeast("view", "info")
	if watchedGroupIDSet {
		visibleItemDescendants = visibleItemDescendants.HavingMaxPermissionAtLeast("watch", "result")
	}

	// the number of started results is much smaller than the number of answers, so we start from results
	answersQuery := srv.Store.Answers().
		Select(`
			STRAIGHT_JOIN /* tell the optimizer we don't want to convert IN(...) into JOIN */
			'submission' AS activity_type,
			answers.created_at AS at,
			answers.id AS answer_id,
			answers.attempt_id, answers.participant_id,
			answers.item_id, author_id AS user_id`).
		Where("answers.type = 'Submission'").
		Where("answers.created_at <= NOW()").
		Where("answers.participant_id <= (SELECT MAX(id) FROM participants)").
		Where("answers.participant_id >= (SELECT MIN(id) FROM participants)").
		Where("answers.participant_id IN (SELECT id FROM participants)").
		Where("answers.item_id <= (SELECT MAX(id) FROM items_to_show)").
		Where("answers.item_id >= (SELECT MIN(id) FROM items_to_show)").
		Where("answers.item_id IN (SELECT id FROM items_to_show)")

	answersQuery = service.NewQueryLimiter().Apply(r, answersQuery)
	answersQuery, apiError = service.ApplySortingAndPaging(r, answersQuery,
		map[string]*service.FieldSortingParams{
			"at":             {ColumnName: "answers.created_at", FieldType: "time"},
			"item_id":        {ColumnName: "answers.item_id", FieldType: "int64"},
			"participant_id": {ColumnName: "answers.participant_id", FieldType: "int64"},
			"attempt_id":     {ColumnName: "answers.attempt_id", FieldType: "int64"},
			"answer_id":      {ColumnName: "answers.id", FieldType: "int64"},
			"activity_type":  {FieldType: "int64", Ignore: true},
		},
		"-at,item_id,participant_id,-attempt_id,-activity_type,answer_id", []string{"answer_id"}, true)
	if apiError != service.NoError {
		return nil, apiError
	}

	answersQuery = srv.Store.Raw("SELECT limited_answers.*, gradings.score FROM ? AS limited_answers", answersQuery.SubQuery()).
		Joins("LEFT JOIN gradings ON gradings.answer_id = limited_answers.answer_id")

	startedResultsQuery := srv.Store.Table("results AS started_results").
		Select(`
			STRAIGHT_JOIN /* tell the optimizer we don't want to convert IN(...) into JOIN */
			'result_started' AS activity_type,
			started_at AS at,
			-1 AS answer_id,
			attempt_id, participant_id, item_id, participant_id AS user_id,
			NULL AS score`).
		Where("item_id IN (SELECT id FROM items_to_show)").
		Where("started_at <= NOW()").
		Where("participant_id IN (SELECT id FROM participants)")

	startedResultsQuery = service.NewQueryLimiter().Apply(r, startedResultsQuery)
	startedResultsQuery, _ = service.ApplySortingAndPaging(r, startedResultsQuery,
		map[string]*service.FieldSortingParams{
			"at":             {ColumnName: "started_at", FieldType: "time"},
			"item_id":        {ColumnName: "item_id", FieldType: "int64"},
			"participant_id": {ColumnName: "participant_id", FieldType: "int64"},
			"attempt_id":     {ColumnName: "attempt_id", FieldType: "int64"},
			"activity_type":  {ColumnName: "1", FieldType: "int64"},
			"answer_id":      {FieldType: "int64", Ignore: true},
		},
		"-at,item_id,participant_id,-attempt_id,-activity_type,answer_id", []string{"participant_id", "attempt_id", "item_id"}, true)

	validatedResultsQuery := srv.Store.Results().
		Select(`
			STRAIGHT_JOIN /* tell the optimizer we don't want to convert IN(...) into JOIN */
			'result_validated' AS activity_type,
			results.validated_at AS at,
			-1 AS answer_id,
			results.attempt_id, results.participant_id,
			item_id, participant_id AS user_id,
			NULL AS score`).
		Where("results.item_id IN (SELECT id FROM items_to_show)").
		Where("results.validated_at <= NOW()").
		Where("results.participant_id IN (SELECT id FROM participants)")

	validatedResultsQuery = service.NewQueryLimiter().Apply(r, validatedResultsQuery)
	validatedResultsQuery, _ = service.ApplySortingAndPaging(r, validatedResultsQuery,
		map[string]*service.FieldSortingParams{
			"at":             {ColumnName: "validated_at", FieldType: "time"},
			"item_id":        {ColumnName: "item_id", FieldType: "int64"},
			"participant_id": {ColumnName: "participant_id", FieldType: "int64"},
			"attempt_id":     {ColumnName: "attempt_id", FieldType: "int64"},
			"activity_type":  {ColumnName: "3", FieldType: "int64"},
			"answer_id":      {FieldType: "int64", Ignore: true},
		},
		"-at,item_id,participant_id,-attempt_id,-activity_type,answer_id", []string{"participant_id", "attempt_id", "item_id"}, true)

	// There is a bug in Gorm. They assume that queries constructed with Raw() already contain WHERE.
	// It is easier to add a workaround here than to patch Gorm because many programs depend on this behavior.
	unionQueryString := "SELECT * FROM (? UNION ALL ? UNION ALL ?) AS un"
	if len(r.URL.Query()["from.answer_id"]) > 0 {
		unionQueryString += " WHERE "
	}

	unionQuery := srv.Store.Raw(unionQueryString,
		answersQuery.SubQuery(), startedResultsQuery.SubQuery(), validatedResultsQuery.SubQuery())
	unionQuery = service.NewQueryLimiter().Apply(r, unionQuery)
	unionQuery, _ = service.ApplySortingAndPaging(r, unionQuery,
		map[string]*service.FieldSortingParams{
			"at":             {ColumnName: "at", FieldType: "time"},
			"participant_id": {ColumnName: "participant_id", FieldType: "int64"},
			"attempt_id":     {ColumnName: "attempt_id", FieldType: "int64"},
			"item_id":        {ColumnName: "item_id", FieldType: "int64"},
			"activity_type": {
				ColumnName: "CASE activity_type WHEN 'result_started' THEN 1 WHEN 'submission' THEN 2 WHEN 'result_validated' THEN 3 END",
				FieldType:  "int64",
			},
			"answer_id": {ColumnName: "answer_id", FieldType: "int64"},
		},
		"-at,item_id,participant_id,-attempt_id,-activity_type,answer_id",
		[]string{"participant_id", "attempt_id", "item_id", "activity_type", "answer_id"}, true)

	query := srv.Store.Raw(`
		WITH items_to_show AS ?, participants AS ?
		SELECT STRAIGHT_JOIN activity_type, at, answer_id, attempt_id, participant_id, score,
			items.id AS item__id, items.type AS item__type,
			groups.id AS participant__id,
			groups.name AS participant__name,
			groups.type AS participant__type,
			users.login AS user__login,
			users.group_id AS user__id,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.first_name, NULL) AS user__first_name,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.last_name, NULL) AS user__last_name,
			IF(user_strings.language_tag IS NULL, default_strings.title, user_strings.title) AS item__string__title
		FROM ? AS activities`, visibleItemDescendants.SubQuery(), participantsQuery.SubQuery(), user.GroupID, user.GroupID,
		unionQuery.SubQuery()).
		Joins("JOIN items ON items.id = item_id").
		Joins("JOIN `groups` ON groups.id = participant_id").
		Joins("LEFT JOIN users ON users.group_id = user_id").
		WithPersonalInfoViewApprovals(user).
		JoinsUserAndDefaultItemStrings(user)
	return query, service.NoError
}

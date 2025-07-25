package items

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/app/structures"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

const itemActivityLogStraightJoinBoundary = 10000

// swagger:model itemActivityLogResponseRow
type itemActivityLogResponseRow struct {
	// required: true
	At *database.Time `json:"at"`
	// required: true
	// enum: result_started,submission,result_validated,saved_answer,current_answer
	ActivityType string `json:"activity_type"`
	// required: true
	AttemptID int64 `json:"attempt_id,string"`
	// `answers.id`
	AnswerID *int64 `json:"answer_id,string,omitempty"`
	// use this as `{from.answer_id}` for pagination
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

		*structures.UserPersonalInfo
		ShowPersonalInfo bool `json:"-"`
	} `gorm:"embedded;embedded_prefix:user__" json:"user,omitempty"`
	// required: true
	Item struct {
		// required: true
		ID int64 `json:"id,string"`
		// required: true
		// enum: Chapter,Task,Skill
		Type string `json:"type"`
		// required: true
		String struct {
			// required: true
			Title *string `json:"title"`
		} `json:"string" gorm:"embedded;embedded_prefix:string__"`
	} `json:"item" gorm:"embedded;embedded_prefix:item__"`
	// only when `{watched_group_id}` is given
	CanWatchAnswer *bool `json:"can_watch_answer,omitempty"`
}

// swagger:operation GET /items/{ancestor_item_id}/log items itemActivityLogForItem
//
//	---
//	summary: Activity log on an item
//	description: >
//		Returns rows from `answers` and started/validated `results`
//		with additional info on users and items for the participant or the `{watched_group_id}` group
//		(only one of `{as_team_id}` and `{watched_group_id}` can be given).
//
//
//		If possible, items titles are shown in the authenticated user's default language.
//		Otherwise, the item's default language is used.
//
//
//		`first_name` and `last_name` of users are only visible to the users themselves and
//		to managers of those users' groups to which they provided view access to personal data.
//
//
//		If `{watched_group_id}` is given, all rows of the result are related to descendant groups of `{watched_group_id}`
//		and items that are descendants of `{ancestor_item_id}` (+ `{ancestor_item_id}` itself) and visible to the current user
//		(at least 'info' access with `can_watch` >= 'result').
//
//
//		If `{watched_group_id}` is not given, all rows of the result are related to the participant group (the current user or `{as_team_id}`)
//		and items that are descendants of `{ancestor_item_id}` (+ `{ancestor_item_id}` itself) and
//		visible to the current user (at least 'info' access).
//	parameters:
//		- name: ancestor_item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: as_team_id
//			in: query
//			type: integer
//			format: int64
//		- name: watched_group_id
//			description: The current user should be a manager of the watched group with `can_watch_members` = true,
//							 otherwise the 'forbidden' error is returned
//			in: query
//			type: integer
//			format: int64
//		- name: from.item_id
//			description: Start the page from the row next to the row with `item_id`=`{from.item_id}`
//							 (all other `{from.*}` parameters are required when `{from.item_id}` is present)
//			in: query
//			type: integer
//			format: int64
//		- name: from.participant_id
//			description: Start the page from the row next to the row with `participant_id`=`{from.participant_id}`
//							 (all other `{from.*}` parameters are required when `{from.participant_id}` is present)
//			in: query
//			type: integer
//			format: int64
//		- name: from.attempt_id
//			description: Start the page from the row next to the row with `attempt_id`=`{from.attempt_id}`
//							 (all other `{from.*}` parameters are required when `{from.attempt_id}` is present)
//			in: query
//			type: integer
//			format: int64
//		- name: from.answer_id
//			description: Start the page from the row next to the row with `from_answer_id`=`{from.answer_id}`
//							 (all other `{from.*}` parameters are required when `{from.answer_id}` is present)
//			in: query
//			type: integer
//			format: int64
//		- name: from.activity_type
//			description: Start the page from the row next to the row with `activity_type`=`{from.activity_type}`
//							 (all other `{from.*}` parameters are required when `{from.activity_type}` is present)
//			in: query
//			type: string
//			enum: [result_started,submission,result_validated,saved_answer,current_answer]
//		- name: limit
//			description: Display the first N rows
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//			description: OK. The array of users answers
//			schema:
//				type: array
//				items:
//			 		"$ref": "#/definitions/itemActivityLogResponseRow"
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
func (srv *Service) getActivityLogForItem(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	itemID, err := service.ResolveURLQueryPathInt64Field(httpRequest, "ancestor_item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	return srv.getActivityLog(responseWriter, httpRequest, &itemID)
}

// swagger:operation GET /items/log items itemActivityLogForAllItems
//
//	---
//	summary: Activity log for all visible items
//	description: >
//		Returns rows from `answers` and started/validated `results`
//		with additional info on users and items for the participant or the `{watched_group_id}` group
//		(only one of `{as_team_id}` and `{watched_group_id}` can be given).
//
//
//		If possible, items titles are shown in the authenticated user's default language.
//		Otherwise, the item's default language is used.
//
//
//		`first_name` and `last_name` of users are only visible to the users themselves and
//		to managers of those users' groups to which they provided view access to personal data.
//
//
//		If `{watched_group_id}` is given, all rows of the result are related to descendant groups of `{watched_group_id}`
//		and items that are visible to the current user (at least 'info' access with `can_watch` >= 'result').
//
//
//		If `{watched_group_id}` is not given, all rows of the result are related to the participant group (the current user or `{as_team_id}`)
//		and items that are visible to the current user (at least 'info' access).
//	parameters:
//		- name: as_team_id
//			in: query
//			type: integer
//		- name: watched_group_id
//			description: The current user should be a manager of the watched group with `can_watch_members` = true,
//							 otherwise the 'forbidden' error is returned
//			in: query
//			type: integer
//		- name: from.item_id
//			description: Start the page from the row next to the row with `item_id`=`{from.item_id}`
//							 (all other `{from.*}` parameters are required when `{from.item_id}` is present)
//			in: query
//			type: integer
//		- name: from.participant_id
//			description: Start the page from the row next to the row with `participant_id`=`{from.participant_id}`
//							 (all other `{from.*}` parameters are required when `{from.participant_id}` is present)
//			in: query
//			type: integer
//		- name: from.attempt_id
//			description: Start the page from the row next to the row with `attempt_id`=`{from.attempt_id}`
//							 (all other `{from.*}` parameters are required when `{from.attempt_id}` is present)
//			in: query
//			type: integer
//		- name: from.answer_id
//			description: Start the page from the row next to the row with `from_answer_id`=`{from.answer_id}`
//							 (all other `{from.*}` parameters are required when `{from.answer_id}` is present)
//			in: query
//			type: integer
//		- name: from.activity_type
//			description: Start the page from the row next to the row with `activity_type`=`{from.activity_type}`
//							 (all other `{from.*}` parameters are required when `{from.activity_type}` is present)
//			in: query
//			type: string
//			enum: [result_started,submission,result_validated,saved_answer,current_answer]
//		- name: limit
//			description: Display the first N rows
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//			description: OK. The array of users answers
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/itemActivityLogResponseRow"
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
func (srv *Service) getActivityLogForAllItems(w http.ResponseWriter, r *http.Request) error {
	return srv.getActivityLog(w, r, nil)
}

func (srv *Service) getActivityLog(responseWriter http.ResponseWriter, httpRequest *http.Request, itemID *int64) error {
	user := srv.GetUser(httpRequest)

	const (
		resultStarted  = 1
		submission     = 2
		resultValidate = 3
		savedAnswer    = 4
		currentAnswer  = 5
	)

	// check and patch from.activity_type to make it integer
	urlParams := httpRequest.URL.Query()
	if len(urlParams["from.activity_type"]) > 0 {
		stringValue := httpRequest.URL.Query().Get("from.activity_type")
		var intValue int
		var fromActivityTypeIsCorrect bool
		if intValue, fromActivityTypeIsCorrect = map[string]int{
			"result_started":   resultStarted,
			"submission":       submission,
			"result_validated": resultValidate,
			"saved_answer":     savedAnswer,
			"current_answer":   currentAnswer,
		}[stringValue]; !fromActivityTypeIsCorrect {
			return service.ErrInvalidRequest(
				errors.New(
					"wrong value for from.activity_type (should be one of (result_started, submission, result_validated, saved_answer, current_answer))"))
		}
		urlParams["from.activity_type"] = []string{strconv.Itoa(intValue)}
		httpRequest.URL.RawQuery = urlParams.Encode()
	}

	fromValues, err := service.ParsePagingParameters(
		httpRequest, service.SortingAndPagingTieBreakers{
			"activity_type":  service.FieldTypeInt64,
			"participant_id": service.FieldTypeInt64,
			"attempt_id":     service.FieldTypeInt64,
			"item_id":        service.FieldTypeInt64,
			"answer_id":      service.FieldTypeInt64,
		})
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	query, err := srv.constructActivityLogQuery(srv.GetStore(httpRequest), httpRequest, itemID, user, fromValues)
	service.MustNotBeError(err)

	var result []itemActivityLogResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	fromAnswerID := int64(-1)
	if len(urlParams["from.answer_id"]) > 0 {
		// the error checking has been already done in constructActivityLogQuery()
		fromAnswerID, _ = service.ResolveURLQueryGetInt64Field(httpRequest, "from.answer_id")
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
		} else if !result[index].User.ShowPersonalInfo {
			result[index].User.UserPersonalInfo = nil
		}
	}

	render.Respond(responseWriter, httpRequest, result)
	return nil
}

func (srv *Service) constructActivityLogQuery(store *database.DataStore, httpRequest *http.Request, itemID *int64,
	user *database.User, fromValues map[string]interface{},
) (*database.DB, error) {
	participantID := service.ParticipantIDFromContext(httpRequest.Context())
	watchedGroupID, watchedGroupIDIsSet, err := srv.ResolveWatchedGroupID(httpRequest)
	if err != nil {
		return nil, err
	}
	participantsQuery := store.Raw("SELECT ? AS id", participantID)

	visibleItemDescendants := store.Permissions().MatchingGroupAncestors(participantID).
		Select("item_id AS id").
		Group("item_id").
		HavingMaxPermissionAtLeast("view", "info")

	if watchedGroupIDIsSet {
		if len(httpRequest.URL.Query()["as_team_id"]) != 0 {
			return nil, service.ErrInvalidRequest(errors.New("only one of as_team_id and watched_group_id can be given"))
		}
		participantsQuery = store.ActiveGroupAncestors().Where("ancestor_group_id = ?", watchedGroupID).
			Select("child_group_id AS id")
		visibleItemDescendants = visibleItemDescendants.HavingMaxPermissionAtLeast("watch", "result")
	}

	if itemID != nil {
		itemDescendants := store.ItemAncestors().DescendantsOf(*itemID).Select("child_item_id")
		visibleItemDescendants = visibleItemDescendants.
			Where("item_id = ? OR item_id IN ?", *itemID, itemDescendants.SubQuery())
	}

	participantsQuerySubQuery := participantsQuery.SubQuery()
	visibleItemDescendantsSubQuery := visibleItemDescendants.SubQuery()

	var cnt struct {
		Cnt int
	}
	service.MustNotBeError(store.Raw(`
		WITH items_to_show AS ?, participants AS ?
		SELECT COUNT(*) AS cnt FROM answers
		WHERE (answers.item_id IN (SELECT id FROM items_to_show)) AND
			(answers.participant_id IN (SELECT id FROM participants))`,
		visibleItemDescendantsSubQuery, participantsQuerySubQuery).Scan(&cnt).Error())

	const answersQueryDefaultSelect = `
			CASE answers.type
				WHEN 'Submission' THEN 2
				WHEN 'Saved' THEN 4
				WHEN 'Current' THEN 5
			END AS activity_type_int,
			answers.type + 0 AS type,
			answers.created_at AS at,
			answers.id AS answer_id,
			answers.attempt_id, answers.participant_id,
			answers.item_id,
			author_id AS user_id`
	answersQuerySelect := answersQueryDefaultSelect

	answersQuery := store.Answers().DB

	if cnt.Cnt > itemActivityLogStraightJoinBoundary ||
		httpRequest.Context().Value(service.APIServiceContextVariableName("forceStraightJoinInItemActivityLog")) == "force" {
		// it will be faster to go through all the answers table with limit in this case because sorting is too expensive
		answersQuerySelect = "STRAIGHT_JOIN /* tell the optimizer we don't want to convert IN(...) into JOIN */\n" + answersQueryDefaultSelect
		// also, we need to FORCE INDEX to do the sorted index scan
		answersQuery = store.Table("answers FORCE INDEX (created_at_d_item_id_participant_id_attempt_id_d_type_d_id_autho)").
			Where("answers.created_at <= NOW()").
			Where("answers.participant_id <= (SELECT MAX(id) FROM participants)").
			Where("answers.participant_id >= (SELECT MIN(id) FROM participants)").
			Where("answers.item_id <= (SELECT MAX(id) FROM items_to_show)").
			Where("answers.item_id >= (SELECT MIN(id) FROM items_to_show)")
	}
	answersQuery = applyMandatoryAnswersConditions(answersQuery.Select(answersQuerySelect))

	startedResultsQuery := store.Table("results AS started_results").
		Select(`
			STRAIGHT_JOIN /* tell the optimizer we don't want to convert IN(...) into JOIN */
			1 AS activity_type_int,
			65535 AS type, /* results don't have a type */
			started_at AS at,
			-1 AS answer_id,
			started_results.attempt_id, started_results.participant_id, started_results.item_id, started_results.participant_id AS user_id,
			NULL AS score`).
		Where("started_results.item_id <= (SELECT MAX(id) FROM items_to_show)").
		Where("started_results.item_id >= (SELECT MIN(id) FROM items_to_show)").
		Where("started_results.item_id IN (SELECT id FROM items_to_show)").
		Where("started_results.started_at <= NOW()").
		Where("started_results.participant_id <= (SELECT MAX(id) FROM participants)").
		Where("started_results.participant_id >= (SELECT MIN(id) FROM participants)").
		Where("started_results.participant_id IN (SELECT id FROM participants)")

	validatedResultsQuery := store.Table("results AS validated_results").
		Select(`
			STRAIGHT_JOIN /* tell the optimizer we don't want to convert IN(...) into JOIN */
			3 AS activity_type_int,
			65535 AS type, /* results don't have a type */
			validated_results.validated_at AS at,
			-1 AS answer_id,
			validated_results.attempt_id, validated_results.participant_id,
			validated_results.item_id, validated_results.participant_id AS user_id,
			NULL AS score`).
		Where("validated_results.item_id IN (SELECT id FROM items_to_show)").
		Where("validated_results.validated_at <= NOW()").
		Where("validated_results.participant_id IN (SELECT id FROM participants)")

	startFromRowSubQuery, startFromRowCTESubQuery := srv.generateSubQueriesForPagination(
		store, httpRequest.URL.Query().Get("from.activity_type"), startedResultsQuery, validatedResultsQuery,
		applyMandatoryAnswersConditions(store.Answers().Select(answersQueryDefaultSelect)),
		fromValues)

	answersQuery = service.NewQueryLimiter().Apply(httpRequest, answersQuery)
	// we have already checked for possible errors in constructActivityLogQuery()
	answersQuery, _ = service.ApplySortingAndPaging(
		nil, answersQuery,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"at":             {ColumnName: "answers.created_at"},
				"item_id":        {ColumnName: "answers.item_id"},
				"participant_id": {ColumnName: "answers.participant_id"},
				"attempt_id":     {ColumnName: "answers.attempt_id"},
				"type":           {ColumnName: `answers.type`},
				"answer_id":      {ColumnName: "answers.id"},
			},
			DefaultRules:         "-at,item_id,participant_id,-attempt_id,-type,answer_id",
			IgnoreSortParameter:  true,
			StartFromRowSubQuery: startFromRowSubQuery,
		})

	answersQuery = store.Raw("SELECT limited_answers.*, gradings.score FROM ? AS limited_answers", answersQuery.SubQuery()).
		Joins("LEFT JOIN gradings ON gradings.answer_id = limited_answers.answer_id")

	startedResultsQuery = service.NewQueryLimiter().Apply(httpRequest, startedResultsQuery)
	// we have already checked for possible errors in constructActivityLogQuery()
	startedResultsQuery, _ = service.ApplySortingAndPaging(
		nil, startedResultsQuery,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"at":             {ColumnName: "started_results.started_at"},
				"item_id":        {ColumnName: "started_results.item_id"},
				"participant_id": {ColumnName: "started_results.participant_id"},
				"attempt_id":     {ColumnName: "started_results.attempt_id"},
			},
			DefaultRules:         "-at,item_id,participant_id,-attempt_id",
			IgnoreSortParameter:  true,
			StartFromRowSubQuery: startFromRowSubQuery,
		})

	validatedResultsQuery = service.NewQueryLimiter().Apply(httpRequest, validatedResultsQuery)
	// we have already checked for possible errors in constructActivityLogQuery()
	validatedResultsQuery, _ = service.ApplySortingAndPaging(
		nil, validatedResultsQuery,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"at":             {ColumnName: "validated_results.validated_at"},
				"item_id":        {ColumnName: "validated_results.item_id"},
				"participant_id": {ColumnName: "validated_results.participant_id"},
				"attempt_id":     {ColumnName: "validated_results.attempt_id"},
			},
			DefaultRules:         "-at,item_id,participant_id,-attempt_id",
			IgnoreSortParameter:  true,
			StartFromRowSubQuery: startFromRowSubQuery,
		})

	unionCTEQuery := store.Raw("SELECT * FROM (? UNION ALL ? UNION ALL ?) AS un",
		answersQuery.SubQuery(), startedResultsQuery.SubQuery(), validatedResultsQuery.SubQuery())
	unionQuery := store.Table("un")
	unionQuery = service.NewQueryLimiter().Apply(httpRequest, unionQuery)
	unionQuery, _ = service.ApplySortingAndPaging(
		nil, unionQuery,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"at":                {ColumnName: "un.at"},
				"participant_id":    {ColumnName: "un.participant_id"},
				"attempt_id":        {ColumnName: "un.attempt_id"},
				"item_id":           {ColumnName: "un.item_id"},
				"activity_type_int": {ColumnName: "un.activity_type_int"},
				"answer_id":         {ColumnName: "un.answer_id"},
			},
			DefaultRules:         "-at,item_id,participant_id,-attempt_id,-activity_type_int,answer_id",
			IgnoreSortParameter:  true,
			StartFromRowSubQuery: startFromRowSubQuery,
		})

	query := store.Raw(`
		WITH items_to_show AS ?, participants AS ?, start_from_row AS ?, un AS ?
		SELECT STRAIGHT_JOIN
			CASE activity_type_int
				WHEN 1 THEN 'result_started'
				WHEN 2 THEN 'submission'
				WHEN 3 THEN 'result_validated'
				WHEN 4 THEN 'saved_answer'
				WHEN 5 THEN 'current_answer'
			END AS activity_type,
			at, answer_id, attempt_id, participant_id, score,
			items.id AS item__id, items.type AS item__type,`+
		golang.LazyIf(watchedGroupIDIsSet,
			func() string {
				return fmt.Sprintf(`
			permissions.can_watch_generated_value >= %d AS can_watch_answer,`,
					store.PermissionsGranted().WatchIndexByName("answer"))
			})+`
			groups.id AS participant__id,
			groups.name AS participant__name,
			groups.type AS participant__type,
			users.login AS user__login,
			users.group_id AS user__id,
			users.group_id = ? OR personal_info_view_approvals.approved AS user__show_personal_info,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.first_name, NULL) AS user__first_name,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.last_name, NULL) AS user__last_name,
			IF(user_strings.language_tag IS NULL, default_strings.title, user_strings.title) AS item__string__title
		FROM ? AS activities`, visibleItemDescendantsSubQuery, participantsQuerySubQuery,
		startFromRowCTESubQuery, unionCTEQuery.SubQuery(),
		user.GroupID, user.GroupID, user.GroupID,
		unionQuery.SubQuery()).
		Joins("JOIN items ON items.id = item_id").
		Joins("JOIN `groups` ON groups.id = participant_id").
		Joins("LEFT JOIN users ON users.group_id = user_id").
		WithPersonalInfoViewApprovals(user).
		JoinsUserAndDefaultItemStrings(user)

	if watchedGroupIDIsSet {
		query = query.JoinsPermissionsForGroupToItems(participantID)
	}

	return query, nil
}

func applyMandatoryAnswersConditions(answersQuery *database.DB) *database.DB {
	return answersQuery.
		Where("answers.participant_id IN (SELECT id FROM participants)").
		Where("answers.item_id IN (SELECT id FROM items_to_show)")
}

func (srv *Service) generateSubQueriesForPagination(
	store *database.DataStore, activityTypeIndex string, startedResultsQuery, validatedResultsQuery,
	answersQuery *database.DB, fromValues map[string]interface{}) (
	startFromRowSubQuery, startFromRowCTESubQuery interface{},
) {
	startFromRowSubQuery = store.Table("start_from_row").SubQuery()
	var startFromRowQuery *database.DB
	switch activityTypeIndex {
	case "1": // result_started
		startFromRowQuery = startedResultsQuery.
			Where("started_results.participant_id = ?", fromValues["participant_id"]).
			Where("started_results.attempt_id = ?", fromValues["attempt_id"]).
			Where("started_results.item_id = ?", fromValues["item_id"])
	case "2", "4", "5": // submission/saved_answer/current_answer
		startFromRowQuery = answersQuery.Where("answers.id = ?", fromValues["answer_id"])
	case "3": // result_validated
		startFromRowQuery = validatedResultsQuery.
			Where("validated_results.participant_id = ?", fromValues["participant_id"]).
			Where("validated_results.attempt_id = ?", fromValues["attempt_id"]).
			Where("validated_results.item_id = ?", fromValues["item_id"])
	default:
		startFromRowQuery = store.Raw("SELECT 1")
		startFromRowSubQuery = service.FromFirstRow
	}
	return startFromRowSubQuery, startFromRowQuery.Limit(1).SubQuery()
}

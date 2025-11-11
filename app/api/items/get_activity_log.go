package items

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
		*structures.UserPersonalInfo

		// required: true
		ID *int64 `json:"id,string"`
		// required: true
		Login string `json:"login"`

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
	// only when `{watched_group_id}` is given or when getting activity log for a thread
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
//		If possible, item titles are shown in the authenticated user's default language.
//		Otherwise, the item's default language is used.
//
//
//		`first_name` and `last_name` (from `profile`) of users are only visible to the users themselves and
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

	return srv.getActivityLogForParticipantOrWatchedGroup(responseWriter, httpRequest, &itemID)
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
//		`first_name` and `last_name` (from `profile`) of users are only visible to the users themselves and
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
	return srv.getActivityLogForParticipantOrWatchedGroup(w, r, nil)
}

func (srv *Service) getActivityLogForParticipantOrWatchedGroup(
	responseWriter http.ResponseWriter, httpRequest *http.Request, ancestorItemID *int64,
) error {
	participantID := service.ParticipantIDFromContext(httpRequest.Context())
	watchedGroupID, watchedGroupIDIsSet, err := srv.ResolveWatchedGroupID(httpRequest)
	if err != nil {
		return err
	}

	store := srv.GetStore(httpRequest)

	participantsQuery := store.Raw("SELECT ? AS id", participantID)
	visibleItemDescendantsQuery := store.Permissions().MatchingGroupAncestors(participantID).
		Select("item_id AS id").
		Group("item_id").
		HavingMaxPermissionAtLeast("view", "info")

	if watchedGroupIDIsSet {
		if len(httpRequest.URL.Query()["as_team_id"]) != 0 {
			return service.ErrInvalidRequest(errors.New("only one of as_team_id and watched_group_id can be given"))
		}
		participantsQuery = store.ActiveGroupAncestors().Where("ancestor_group_id = ?", watchedGroupID).
			Select("child_group_id AS id")
		visibleItemDescendantsQuery = visibleItemDescendantsQuery.HavingMaxPermissionAtLeast("watch", "result")
	}

	if ancestorItemID != nil {
		itemDescendants := store.ItemAncestors().DescendantsOf(*ancestorItemID).Select("child_item_id")
		visibleItemDescendantsQuery = visibleItemDescendantsQuery.
			Where("item_id = ? OR item_id IN ?", *ancestorItemID, itemDescendants.SubQuery())
	}

	return srv.getActivityLog(
		responseWriter, httpRequest,
		[]string{"activity_type", "participant_id", "attempt_id", "item_id", "answer_id"},
		func(query *database.DB) *database.DB {
			return query.
				With("items_to_show", visibleItemDescendantsQuery).
				With("participants", participantsQuery)
		},
		func(answersQuery *database.DB) *database.DB {
			return answersQuery.
				Where("answers.participant_id IN (SELECT id FROM participants)").
				Where("answers.item_id IN (SELECT id FROM items_to_show)")
		},
		func(answersQuery *database.DB) *database.DB {
			return answersQuery.
				Where("answers.created_at <= NOW()").
				Where("answers.participant_id <= (SELECT MAX(id) FROM participants)").
				Where("answers.participant_id >= (SELECT MIN(id) FROM participants)").
				Where("answers.item_id <= (SELECT MAX(id) FROM items_to_show)").
				Where("answers.item_id >= (SELECT MIN(id) FROM items_to_show)")
		},
		func(startedResultsQuery *database.DB) *database.DB {
			return startedResultsQuery.
				Where("started_results.item_id <= (SELECT MAX(id) FROM items_to_show)").
				Where("started_results.item_id >= (SELECT MIN(id) FROM items_to_show)").
				Where("started_results.item_id IN (SELECT id FROM items_to_show)").
				Where("started_results.started_at <= NOW()").
				Where("started_results.participant_id <= (SELECT MAX(id) FROM participants)").
				Where("started_results.participant_id >= (SELECT MIN(id) FROM participants)").
				Where("started_results.participant_id IN (SELECT id FROM participants)")
		},
		func(validatedResultsQuery *database.DB) *database.DB {
			return validatedResultsQuery.
				Where("validated_results.item_id IN (SELECT id FROM items_to_show)").
				Where("validated_results.validated_at <= NOW()").
				Where("validated_results.participant_id IN (SELECT id FROM participants)")
		},
		func(unQuery *database.DB) *database.DB {
			return unQuery.JoinsPermissionsForGroupToItems(participantID)
		},
		golang.LazyIfElse(watchedGroupIDIsSet,
			func() string {
				return fmt.Sprintf(`permissions.can_watch_generated_value >= %d AS can_watch_answer`,
					store.PermissionsGranted().WatchIndexByName("answer"))
			},
			func() string { return "NULL AS can_watch_answer" }),
		"-at,item_id,participant_id,-attempt_id",
		true)
}

func (srv *Service) getActivityLog(responseWriter http.ResponseWriter, httpRequest *http.Request,
	pagingColumns []string,
	addWithTablesFunc,
	answersQueryMandatoryConditionsFunc, answersQueryStraightJoinConditionsFunc,
	startedResultsQueryConditionsFunc, validatedResultsQueryConditionsFunc,
	unCanWatchAnswersConditionsFunc func(*database.DB) *database.DB,
	canWatchAnswerColumnString, resultsSortRules string,
	doStraightJoinAndForceIndexInAnswersQueryWhenNeeded bool,
) error {
	user := srv.GetUser(httpRequest)

	const (
		resultStarted   = 1
		submission      = 2
		resultValidated = 3
		savedAnswer     = 4
		currentAnswer   = 5
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
			"result_validated": resultValidated,
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

	fromValues, err := constructFromValuesForActivityLog(pagingColumns, httpRequest)
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	query := constructActivityLogQuery(
		srv.GetStore(httpRequest), httpRequest, user, fromValues,
		addWithTablesFunc,
		answersQueryMandatoryConditionsFunc, answersQueryStraightJoinConditionsFunc,
		startedResultsQueryConditionsFunc, validatedResultsQueryConditionsFunc,
		unCanWatchAnswersConditionsFunc, canWatchAnswerColumnString,
		resultsSortRules,
		doStraightJoinAndForceIndexInAnswersQueryWhenNeeded)

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

func constructFromValuesForActivityLog(pagingColumns []string, httpRequest *http.Request) (map[string]interface{}, error) {
	pagingParameters := make(service.SortingAndPagingTieBreakers, len(pagingColumns))
	for _, column := range pagingColumns {
		pagingParameters[column] = service.FieldTypeInt64
	}
	fromValues, err := service.ParsePagingParameters(httpRequest, pagingParameters)
	if err != nil {
		return nil, err
	}
	return fromValues, nil
}

func constructActivityLogQuery(store *database.DataStore, httpRequest *http.Request,
	user *database.User, fromValues map[string]interface{},
	addWithTablesFunc,
	answersQueryMandatoryConditionsFunc, answersQueryStraightJoinConditionsFunc,
	startedResultsQueryConditionsFunc, validatedResultsQueryConditionsFunc,
	unCanWatchAnswersConditionsFunc func(*database.DB) *database.DB,
	canWatchAnswerColumnString, resultsSortRules string,
	doStraightJoinAndForceIndexInAnswersQueryWhenNeeded bool,
) *database.DB {
	const answersQueryDefaultSelect = `
			answers.activity_type_int,
			answers.type + 0 AS type,
			answers.created_at AS at,
			answers.id AS answer_id,
			answers.attempt_id, answers.participant_id,
			answers.item_id,
			author_id AS user_id`
	answersQuerySelect := answersQueryDefaultSelect

	answersQuery := store.Answers().DB

	if doStraightJoinAndForceIndexInAnswersQueryWhenNeeded {
		var cnt struct {
			Cnt int
		}
		service.MustNotBeError(
			addWithTablesFunc(answersQueryMandatoryConditionsFunc(store.Answers().Select("count(*) AS cnt"))).
				Scan(&cnt).Error())

		if cnt.Cnt > itemActivityLogStraightJoinBoundary ||
			httpRequest.Context().Value(service.APIServiceContextVariableName("forceStraightJoinInItemActivityLog")) == "force" {
			// it will be faster to go through all the answers table with limit in this case because sorting is too expensive
			answersQuerySelect = "STRAIGHT_JOIN /* tell the optimizer we don't want to convert IN(...) into JOIN */\n" + answersQueryDefaultSelect
			// also, we need to FORCE INDEX to do the sorted index scan
			answersQuery = store.Table("answers FORCE INDEX (created_at_d_item_id_participant_id_attempt_id_d_atype_d_id_a_t)")
			answersQuery = answersQueryStraightJoinConditionsFunc(answersQuery)
		}
	}

	answersQuery = answersQueryMandatoryConditionsFunc(answersQuery.Select(answersQuerySelect))

	startedResultsQuery := startedResultsQueryConditionsFunc(store.Table("results AS started_results").
		Select(`
			STRAIGHT_JOIN /* tell the optimizer we don't want to convert IN(...) into JOIN */
			1 AS activity_type_int,
			65535 AS type, /* results don't have a type */
			started_at AS at,
			-1 AS answer_id,
			started_results.attempt_id, started_results.participant_id, started_results.item_id, started_results.participant_id AS user_id,
			NULL AS score`))

	validatedResultsQuery := validatedResultsQueryConditionsFunc(store.Table("results AS validated_results").
		Select(`
			STRAIGHT_JOIN /* tell the optimizer we don't want to convert IN(...) into JOIN */
			3 AS activity_type_int,
			65535 AS type, /* results don't have a type */
			validated_results.validated_at AS at,
			-1 AS answer_id,
			validated_results.attempt_id, validated_results.participant_id,
			validated_results.item_id, validated_results.participant_id AS user_id,
			NULL AS score`))

	startFromRowQuery, startFromRowCTEQuery := generateQueriesForActivityLogPagination(
		store, httpRequest.URL.Query().Get("from.activity_type"), startedResultsQuery, validatedResultsQuery,
		answersQueryMandatoryConditionsFunc(store.Answers().Select(answersQueryDefaultSelect)),
		fromValues)

	answersQuery = service.NewQueryLimiter().Apply(httpRequest, answersQuery)

	resultsSortRules += ",-activity_type_int"
	answersSortRules := resultsSortRules + ",answer_id"
	answersSortFields := constructSortingAndPagingFieldsForActivityLog("answers", answersSortRules)
	answersSortFields["answer_id"] = &service.FieldSortingParams{ColumnName: "answers.id"}  // not answers.answer_id
	answersSortFields["at"] = &service.FieldSortingParams{ColumnName: "answers.created_at"} // not answers.at

	// we have already checked for possible errors in constructActivityLogQuery()
	answersQuery, _ = service.ApplySortingAndPaging(
		nil, answersQuery,
		&service.SortingAndPagingParameters{
			Fields:              answersSortFields,
			DefaultRules:        answersSortRules,
			IgnoreSortParameter: true,
			StartFromRowQuery:   startFromRowQuery,
		})

	answersQuery = store.Raw("SELECT limited_answers.*, gradings.score FROM ? AS limited_answers", answersQuery.SubQuery()).
		Joins("LEFT JOIN gradings ON gradings.answer_id = limited_answers.answer_id")

	startedResultsQuery = service.NewQueryLimiter().Apply(httpRequest, startedResultsQuery)
	startedResultsSortFields := constructSortingAndPagingFieldsForActivityLog("started_results", resultsSortRules)
	startedResultsSortFields["at"] = &service.FieldSortingParams{ColumnName: "started_results.started_at"} // not started_results.at
	startedResultsSortFields["activity_type_int"] = &service.FieldSortingParams{ColumnName: "1"}

	// we have already checked for possible errors in constructActivityLogQuery()
	startedResultsQuery, _ = service.ApplySortingAndPaging(
		nil, startedResultsQuery,
		&service.SortingAndPagingParameters{
			Fields:              startedResultsSortFields,
			DefaultRules:        resultsSortRules,
			IgnoreSortParameter: true,
			StartFromRowQuery:   startFromRowQuery,
		})

	validatedResultsQuery = service.NewQueryLimiter().Apply(httpRequest, validatedResultsQuery)
	validatedResultsSortFields := constructSortingAndPagingFieldsForActivityLog("validated_results", resultsSortRules)
	validatedResultsSortFields["at"] = &service.FieldSortingParams{ColumnName: "validated_results.validated_at"} // not validated_results.at
	validatedResultsSortFields["activity_type_int"] = &service.FieldSortingParams{ColumnName: "3"}

	// we have already checked for possible errors in constructActivityLogQuery()
	validatedResultsQuery, _ = service.ApplySortingAndPaging(
		nil, validatedResultsQuery,
		&service.SortingAndPagingParameters{
			Fields:              validatedResultsSortFields,
			DefaultRules:        resultsSortRules,
			IgnoreSortParameter: true,
			StartFromRowQuery:   startFromRowQuery,
		})

	//nolint:unqueryvet // we select all columns from subqueries having explicitly listed columns
	unionCTEQuery := store.Raw("SELECT * FROM (? UNION ALL ? UNION ALL ?) AS un",
		answersQuery.SubQuery(), startedResultsQuery.SubQuery(), validatedResultsQuery.SubQuery())
	unionQuery := store.Table("un")
	unionQuery = service.NewQueryLimiter().Apply(httpRequest, unionQuery)
	unionSortRules := answersSortRules
	unionQuery, _ = service.ApplySortingAndPaging(
		nil, unionQuery,
		&service.SortingAndPagingParameters{
			Fields:              constructSortingAndPagingFieldsForActivityLog("un", unionSortRules),
			DefaultRules:        unionSortRules,
			IgnoreSortParameter: true,
			StartFromRowQuery:   startFromRowQuery,
		})

	query := addWithTablesFunc(store.Raw(`
		SELECT STRAIGHT_JOIN
			CASE activity_type_int
				WHEN 1 THEN 'result_started'
				WHEN 2 THEN 'submission'
				WHEN 3 THEN 'result_validated'
				WHEN 4 THEN 'saved_answer'
				WHEN 5 THEN 'current_answer'
			END AS activity_type,
			at, answer_id, attempt_id, participant_id, score,
			items.id AS item__id, items.type AS item__type,
			`+canWatchAnswerColumnString+`,
			groups.id AS participant__id,
			groups.name AS participant__name,
			groups.type AS participant__type,
			users.login AS user__login,
			users.group_id AS user__id,
			users.group_id = ? OR personal_info_view_approvals.approved AS user__show_personal_info,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.profile->>'$.first_name', NULL) AS user__first_name,
			IF(users.group_id = ? OR personal_info_view_approvals.approved, users.profile->>'$.last_name', NULL) AS user__last_name,
			IF(user_strings.language_tag IS NULL, default_strings.title, user_strings.title) AS item__string__title
		FROM ? AS activities`,
		user.GroupID, user.GroupID, user.GroupID,
		unionQuery.SubQuery()).
		Joins("JOIN items ON items.id = item_id").
		Joins("JOIN `groups` ON groups.id = participant_id").
		Joins("LEFT JOIN users ON users.group_id = user_id").
		WithPersonalInfoViewApprovals(user).
		JoinsUserAndDefaultItemStrings(user)).
		With("start_from_row", startFromRowCTEQuery).
		With("un", unionCTEQuery)

	query = unCanWatchAnswersConditionsFunc(query)

	query, _ = service.ApplySortingAndPaging(
		nil, query,
		&service.SortingAndPagingParameters{
			Fields:              constructSortingAndPagingFieldsForActivityLog("activities", unionSortRules),
			DefaultRules:        unionSortRules,
			IgnoreSortParameter: true,
			StartFromRowQuery:   service.FromFirstRow(),
		})

	return query
}

func constructSortingAndPagingFieldsForActivityLog(tableName, rules string) service.SortingAndPagingFields {
	columns := strings.Split(rules, ",")
	result := make(service.SortingAndPagingFields, len(columns))
	for _, column := range columns {
		if column[0] == '-' {
			column = column[1:]
		}
		result[column] = &service.FieldSortingParams{ColumnName: tableName + "." + column}
	}
	return result
}

func generateQueriesForActivityLogPagination(
	store *database.DataStore, activityTypeIndex string, startedResultsQuery, validatedResultsQuery,
	answersQuery *database.DB, fromValues map[string]interface{}) (
	startFromRowSubQuery, startFromRowCTESubQuery *database.DB,
) {
	startFromRowSubQuery = store.Table("start_from_row")
	var startFromRowQuery *database.DB
	switch activityTypeIndex {
	case "1": // result_started
		startFromRowQuery = startedResultsQuery.
			Where("started_results.attempt_id = ?", fromValues["attempt_id"])
		if _, ok := fromValues["participant_id"]; ok {
			startFromRowQuery = startFromRowQuery.
				Where("started_results.participant_id = ?", fromValues["participant_id"])
		}
		if _, ok := fromValues["item_id"]; ok {
			startFromRowQuery = startFromRowQuery.
				Where("started_results.item_id = ?", fromValues["item_id"])
		}
	case "2", "4", "5": // submission/saved_answer/current_answer
		startFromRowQuery = answersQuery.Where("answers.id = ?", fromValues["answer_id"])
	case "3": // result_validated
		startFromRowQuery = validatedResultsQuery.
			Where("validated_results.attempt_id = ?", fromValues["attempt_id"])
		if _, ok := fromValues["participant_id"]; ok {
			startFromRowQuery = startFromRowQuery.
				Where("validated_results.participant_id = ?", fromValues["participant_id"])
		}
		if _, ok := fromValues["participant_id"]; ok {
			startFromRowQuery = startFromRowQuery.
				Where("validated_results.item_id = ?", fromValues["item_id"])
		}
	default:
		startFromRowQuery = store.Raw("SELECT 1")
		startFromRowSubQuery = service.FromFirstRow()
	}
	return startFromRowSubQuery, startFromRowQuery.Limit(1)
}

package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type itemStringCommon struct {
	// required: true
	LanguageID int64 `json:"language_id,string"`
	// Nullable
	// required: true
	Title *string `json:"title"`
	// Nullable
	// required: true
	ImageURL *string `json:"image_url"`
}

type itemStringNotGrayed struct {
	// Nullable; only if not grayed
	Subtitle *string `json:"subtitle"`
	// Nullable; only if not grayed
	Description *string `json:"description"`
}

type itemStringRootNodeWithSolutionAccess struct {
	// Nullable; only if the user has access to solutions
	EduComment *string `json:"edu_comment"`
}

// Item-related strings (from `items_strings`) in the user's default language (preferred) or the item's language
type itemString struct {
	*itemStringCommon
	*itemStringNotGrayed
}

// Item-related strings (from `items_strings`) in the user's default language (preferred) or the item's language
type itemStringRoot struct {
	*itemStringCommon
	*itemStringNotGrayed
	*itemStringRootNodeWithSolutionAccess
}

// from `groups_attempts`
type itemUserActiveAttempt struct {
	// Nullable; only if not grayed
	AttemptID int64 `json:"attempt_id,string"`
	// only if not grayed
	Score float32 `json:"score"`
	// only if not grayed
	SubmissionsAttempts int32 `json:"submissions_attempts"`
	// only if not grayed
	Validated bool `json:"validated"`
	// only if not grayed
	Finished bool `json:"finished"`
	// only if not grayed
	KeyObtained bool `json:"key_obtained"`
	// only if not grayed
	HintsCached int32 `json:"hints_cached"`
	// Nullable; only if not grayed
	// example: 2019-09-11T07:30:56Z
	StartedAt *database.Time `json:"started_at,string"`
	// only if not grayed
	// example: 2019-09-11T07:30:56Z
	// type: string
	ValidatedAt *database.Time `json:"validated_at,string"`
	// Nullable; only if not grayed
	// example: 2019-09-11T07:30:56Z
	FinishedAt *database.Time `json:"finished_at,string"`
}

type itemCommonFields struct {
	// items

	// required: true
	ID int64 `json:"id,string"`
	// required: true
	// enum: Root,Category,Chapter,Task,Course
	Type string `json:"type"`
	// required: true
	DisplayDetailsInParent bool `json:"display_details_in_parent"`
	// required: true
	// enum: None,All,AllButOne,Categories,One,Manual
	ValidationType string `json:"validation_type"`
	// whether `items.unlocked_item_ids` is empty
	// required: true
	HasUnlockedItems bool `json:"has_unlocked_items"`
	// required: true
	ScoreMinUnlock int32 `json:"score_min_unlock"`
	// required: true
	// enum: All,Half,One,None
	ContestEnteringCondition string `json:"contest_entering_condition"`
	// required: true
	TeamsEditable bool `json:"teams_editable"`
	// required: true
	ContestMaxTeamSize int32 `json:"contest_max_team_size"`
	// required: true
	HasAttempts bool `json:"has_attempts"`
	// pattern: ^\d{1,3}:[0-5]?\d:[0-5]?\d$
	// example: 838:59:59
	// Nullable
	// required: true
	Duration *string `json:"duration"`
	// required: true
	NoScore bool `json:"no_score"`
	// Nullable
	// required: true
	GroupCodeEnter *bool `json:"group_code_enter"`
}

type itemRootNodeNotChapterFields struct {
	// Nullable; only if not a chapter
	URL *string `json:"url"`
	// only if not a chapter
	UsesAPI bool `json:"uses_api"`
	// only if not a chapter
	HintsAllowed bool `json:"hints_allowed"`
}

type itemChildNode struct {
	*itemCommonFields

	// required: true
	String itemString `json:"string"`

	// items_items (child nodes only)

	// `items_items.order`
	// required: true
	Order int32 `json:"order"`
	// `items_items.category`
	// required: true
	// enum: Undefined,Discovery,Application,Validation,Challenge
	Category string `json:"category"`
	// The rule used to propagate can_view='content' between the parent item and this child
	// enum: none,as_info,as_content
	// required: true
	ContentViewPropagation string `json:"content_view_propagation"`

	// Nullable
	// required: true
	UserActiveAttempt *itemUserActiveAttempt `json:"user_active_attempt"`
}

// swagger:model itemResponse
type itemResponse struct {
	*itemCommonFields

	// root node only

	// required: true
	TitleBarVisible bool `json:"title_bar_visible"`
	// required: true
	ReadOnly bool `json:"read_only"`
	// required: true
	// enum: forceYes,,forceNo,default
	FullScreen string `json:"full_screen"`
	// required: true
	ShowSource bool `json:"show_source"`
	// Nullable
	// required: true
	ValidationMin *int32 `json:"validation_min"`
	// required: true
	ShowUserInfos bool `json:"show_user_infos"`
	// required: true
	// enum: Running,Analysis,Closed
	ContestPhase string `json:"contest_phase"`

	// Nullable
	// required: true
	UserActiveAttempt *itemUserActiveAttempt `json:"user_active_attempt"`

	// required: true
	String itemStringRoot `json:"string"`

	*itemRootNodeNotChapterFields

	// required: true
	Children []itemChildNode `json:"children"`
}

// swagger:operation GET /items/{item_id} items itemView
// ---
// summary: Get an item
// description: Returns data related to the specified item, its children,
//              and the current user's interactions with them
//              (from tables `items`, `items_items`, `items_string`, and `groups_attempts` for the active attempt).
//
//
//              * If the specified item is not visible by the current user, the 'not found' response is returned.
//
//              * If the current user has only grayed access on the specified item, the 'forbidden' error is returned.
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   format: int64
//   required: true
// responses:
//   "200":
//     description: OK. Success response with item data
//     schema:
//       "$ref": "#/definitions/itemResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "404":
//     "$ref": "#/responses/notFoundResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getItem(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	req := &GetItemRequest{}
	if err := req.Bind(httpReq); err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)
	rawData := getRawItemData(srv.Store.Items(), req.ID, user)

	if len(rawData) == 0 || rawData[0].ID != req.ID {
		return service.ErrNotFound(errors.New("insufficient access rights on the given item id"))
	}

	if rawData[0].CanViewGeneratedValue == srv.Store.PermissionsGranted().ViewIndexByName("info") {
		return service.ErrForbidden(errors.New("the item is grayed"))
	}

	permissionGrantedStore := srv.Store.PermissionsGranted()
	response := constructItemResponseFromDBData(&rawData[0], permissionGrantedStore)

	setItemResponseRootNodeFields(response, &rawData, permissionGrantedStore)
	srv.fillItemResponseWithChildren(response, &rawData, permissionGrantedStore)

	render.Respond(rw, httpReq, response)
	return service.NoError
}

// rawItem represents one row of the getItem service data returned from the DB
type rawItem struct {
	// items
	ID                       int64
	Type                     string
	DisplayDetailsInParent   bool
	ValidationType           string
	HasUnlockedItems         bool // whether items.unlocked_item_ids is empty
	ScoreMinUnlock           int32
	ContestEnteringCondition string
	TeamsEditable            bool
	ContestMaxTeamSize       int32
	HasAttempts              bool
	Duration                 *string
	NoScore                  bool
	GroupCodeEnter           *bool

	// root node only
	TitleBarVisible bool
	ReadOnly        bool
	FullScreen      string
	ShowSource      bool
	ValidationMin   *int32
	ShowUserInfos   bool
	ContestPhase    string
	URL             *string // only if not a chapter
	UsesAPI         bool    // only if not a chapter
	HintsAllowed    bool    // only if not a chapter

	// from items_strings: in the user’s default language or (if not available) default language of the item
	StringLanguageID  int64   `sql:"column:language_id"`
	StringTitle       *string `sql:"column:title"`
	StringImageURL    *string `sql:"column:image_url"`
	StringSubtitle    *string `sql:"column:subtitle"`
	StringDescription *string `sql:"column:description"`
	StringEduComment  *string `sql:"column:edu_comment"`

	// from groups_attempts for the active attempt of the current user
	UserActiveAttemptID     *int64         `sql:"column:attempt_id"`
	UserScore               float32        `sql:"column:score"`
	UserSubmissionsAttempts int32          `sql:"column:submissions_attempts"`
	UserValidated           bool           `sql:"column:validated"`
	UserFinished            bool           `sql:"column:finished"`
	UserKeyObtained         bool           `sql:"column:key_obtained"`
	UserHintsCached         int32          `sql:"column:hints_cached"`
	UserStartedAt           *database.Time `sql:"column:started_at"`
	UserValidatedAt         *database.Time `sql:"column:validated_at"`
	UserFinishedAt          *database.Time `sql:"column:finished_at"`

	// items_items
	Order                  int32 `sql:"column:child_order"`
	Category               string
	ContentViewPropagation string

	CanViewGeneratedValue int
}

// getRawItemData reads data needed by the getItem service from the DB and returns an array of rawItem's
func getRawItemData(s *database.ItemStore, rootID int64, user *database.User) []rawItem {
	var result []rawItem

	accessRights := s.AccessRights(user)
	service.MustNotBeError(accessRights.Error())

	commonColumns := `items.id AS id,
		items.type,
		items.display_details_in_parent,
		items.validation_type,
		items.unlocked_item_ids,
		items.score_min_unlock,
		items.contest_entering_condition,
		items.teams_editable,
		items.contest_max_team_size,
		items.has_attempts,
		items.duration,
		items.no_score,
		items.default_language_id,
		items.group_code_enter, `

	rootItemQuery := s.ByID(rootID).Select(
		commonColumns + `items.title_bar_visible,
		items.read_only,
		items.full_screen,
		items.show_source,
		items.validation_min,
		items.show_user_infos,
		items.contest_phase,
		items.url,
		IF(items.type <> 'Chapter', items.uses_api, NULL) AS uses_api,
		IF(items.type <> 'Chapter', items.hints_allowed, NULL) AS hints_allowed,
		NULL AS child_order, NULL AS category, NULL AS content_view_propagation`)

	childrenQuery := s.Select(
		commonColumns+`NULL AS title_bar_visible,
		NULL AS read_only,
		NULL AS full_screen,
		NULL AS show_source,
		NULL AS validation_min,
		NULL AS show_user_infos,
		NULL AS contest_phase,
		NULL AS url,
		NULL AS uses_api,
		NULL AS hints_allowed,
		child_order, category, content_view_propagation`).
		Joins("JOIN items_items ON items.id=child_item_id AND parent_item_id=?", rootID)

	unionQuery := rootItemQuery.UnionAll(childrenQuery.QueryExpr())
	// This query can be simplified if we add a column for relation degrees into `items_ancestors`
	query := s.Raw(`
    SELECT
		  items.id,
      items.type,
		  items.display_details_in_parent,
      items.validation_type,`+
		// unlocked_item_ids is a comma-separated list of item IDs which will be unlocked if this item is validated
		// Here we consider both NULL and an empty string as FALSE
		` COALESCE(items.unlocked_item_ids, '')<>'' as has_unlocked_items,
			items.score_min_unlock,
			items.contest_entering_condition,
			items.teams_editable,
			items.contest_max_team_size,
			items.has_attempts,
			items.duration,
			items.no_score,
			items.group_code_enter,

			COALESCE(user_strings.language_id, default_strings.language_id) AS language_id,
			IF(user_strings.language_id IS NULL, default_strings.title, user_strings.title) AS title,
			IF(user_strings.language_id IS NULL, default_strings.image_url, user_strings.image_url) AS image_url,
			IF(user_strings.language_id IS NULL, default_strings.subtitle, user_strings.subtitle) AS subtitle,
			IF(user_strings.language_id IS NULL, default_strings.description, user_strings.description) AS description,
			IF(user_strings.language_id IS NULL, default_strings.edu_comment, user_strings.edu_comment) AS edu_comment,

			groups_attempts.id AS attempt_id,
			groups_attempts.score AS score,
			groups_attempts.submissions_attempts AS submissions_attempts,
			groups_attempts.validated AS validated,
			groups_attempts.finished AS finished,
			groups_attempts.key_obtained AS key_obtained,
			groups_attempts.hints_cached AS hints_cached,
			groups_attempts.started_at AS started_at,
			groups_attempts.validated_at AS validated_at,
			groups_attempts.finished_at AS finished_at,

			items.child_order AS child_order,
			items.category AS category,
			items.content_view_propagation, `+
		// inputItem only
		` items.title_bar_visible,
			items.read_only,
			items.full_screen,
			items.show_source,
			items.validation_min,
			items.show_user_infos,
			items.contest_phase,
			items.url,
			items.uses_api,
			items.hints_allowed,
			access_rights.can_view_generated_value
    FROM ? items `, unionQuery.SubQuery()).
		JoinsUserAndDefaultItemStrings(user).
		Joins("LEFT JOIN users_items ON users_items.item_id=items.id AND users_items.user_id=?", user.GroupID).
		Joins("LEFT JOIN groups_attempts ON groups_attempts.id=users_items.active_attempt_id").
		Joins("JOIN ? access_rights on access_rights.item_id=items.id AND can_view_generated_value > ?",
			accessRights.SubQuery(), s.PermissionsGranted().ViewIndexByName("none")).
		Order("child_order")

	service.MustNotBeError(query.Scan(&result).Error())
	return result
}

func setItemResponseRootNodeFields(response *itemResponse, rawData *[]rawItem, permissionGrantedStore *database.PermissionGrantedStore) {
	if (*rawData)[0].CanViewGeneratedValue == permissionGrantedStore.ViewIndexByName("solution") {
		response.String.itemStringRootNodeWithSolutionAccess = &itemStringRootNodeWithSolutionAccess{
			EduComment: (*rawData)[0].StringEduComment,
		}
	}
	if (*rawData)[0].Type != "Chapter" {
		response.itemRootNodeNotChapterFields = &itemRootNodeNotChapterFields{
			URL:          (*rawData)[0].URL,
			UsesAPI:      (*rawData)[0].UsesAPI,
			HintsAllowed: (*rawData)[0].HintsAllowed,
		}
	}
	response.TitleBarVisible = (*rawData)[0].TitleBarVisible
	response.ReadOnly = (*rawData)[0].ReadOnly
	response.FullScreen = (*rawData)[0].FullScreen
	response.ShowSource = (*rawData)[0].ShowSource
	response.ValidationMin = (*rawData)[0].ValidationMin
	response.ShowUserInfos = (*rawData)[0].ShowUserInfos
	response.ContestPhase = (*rawData)[0].ContestPhase
}

func constructItemResponseFromDBData(rawData *rawItem, permissionGrantedStore *database.PermissionGrantedStore) *itemResponse {
	result := &itemResponse{
		itemCommonFields: fillItemCommonFieldsWithDBData(rawData),
		String: itemStringRoot{
			itemStringCommon: constructItemStringCommon(rawData),
		},
	}
	result.String.itemStringNotGrayed = constructStringNotGrayed(rawData, permissionGrantedStore)
	result.UserActiveAttempt = constructUserActiveAttempt(rawData, permissionGrantedStore)
	return result
}

func constructItemStringCommon(rawData *rawItem) *itemStringCommon {
	return &itemStringCommon{
		LanguageID: rawData.StringLanguageID,
		Title:      rawData.StringTitle,
		ImageURL:   rawData.StringImageURL,
	}
}

func constructStringNotGrayed(rawData *rawItem, permissionGrantedStore *database.PermissionGrantedStore) *itemStringNotGrayed {
	if rawData.CanViewGeneratedValue == permissionGrantedStore.ViewIndexByName("info") {
		return nil
	}
	return &itemStringNotGrayed{
		Subtitle:    rawData.StringSubtitle,
		Description: rawData.StringDescription,
	}
}

func constructUserActiveAttempt(rawData *rawItem, permissionGrantedStore *database.PermissionGrantedStore) *itemUserActiveAttempt {
	if rawData.CanViewGeneratedValue == permissionGrantedStore.ViewIndexByName("info") || rawData.UserActiveAttemptID == nil {
		return nil
	}
	return &itemUserActiveAttempt{
		AttemptID:           *rawData.UserActiveAttemptID,
		Score:               rawData.UserScore,
		SubmissionsAttempts: rawData.UserSubmissionsAttempts,
		Validated:           rawData.UserValidated,
		Finished:            rawData.UserFinished,
		KeyObtained:         rawData.UserKeyObtained,
		HintsCached:         rawData.UserHintsCached,
		StartedAt:           rawData.UserStartedAt,
		ValidatedAt:         rawData.UserValidatedAt,
		FinishedAt:          rawData.UserFinishedAt,
	}
}

func fillItemCommonFieldsWithDBData(rawData *rawItem) *itemCommonFields {
	result := &itemCommonFields{
		ID:                       rawData.ID,
		Type:                     rawData.Type,
		DisplayDetailsInParent:   rawData.DisplayDetailsInParent,
		ValidationType:           rawData.ValidationType,
		HasUnlockedItems:         rawData.HasUnlockedItems,
		ScoreMinUnlock:           rawData.ScoreMinUnlock,
		ContestEnteringCondition: rawData.ContestEnteringCondition,
		TeamsEditable:            rawData.TeamsEditable,
		ContestMaxTeamSize:       rawData.ContestMaxTeamSize,
		HasAttempts:              rawData.HasAttempts,
		Duration:                 rawData.Duration,
		NoScore:                  rawData.NoScore,
		GroupCodeEnter:           rawData.GroupCodeEnter,
	}
	return result
}

func (srv *Service) fillItemResponseWithChildren(response *itemResponse, rawData *[]rawItem,
	permissionGrantedStore *database.PermissionGrantedStore) {
	response.Children = make([]itemChildNode, 0, len(*rawData))
	for index := range *rawData {
		if index == 0 {
			continue
		}

		child := &itemChildNode{itemCommonFields: fillItemCommonFieldsWithDBData(&(*rawData)[index])}
		child.String.itemStringCommon = constructItemStringCommon(&(*rawData)[index])
		child.String.itemStringNotGrayed = constructStringNotGrayed(&(*rawData)[index], permissionGrantedStore)
		child.UserActiveAttempt = constructUserActiveAttempt(&(*rawData)[index], permissionGrantedStore)
		child.Order = (*rawData)[index].Order
		child.Category = (*rawData)[index].Category
		child.ContentViewPropagation = (*rawData)[index].ContentViewPropagation
		response.Children = append(response.Children, *child)
	}
}

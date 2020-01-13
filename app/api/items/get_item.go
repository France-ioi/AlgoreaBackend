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

type itemStringNotInfo struct {
	// Nullable; only if `can_view` >= 'content'
	Subtitle *string `json:"subtitle"`
	// Nullable; only if `can_view` >= 'content'
	Description *string `json:"description"`
}

type itemStringRootNodeWithSolutionAccess struct {
	// Nullable; only if the user has access to solutions
	EduComment *string `json:"edu_comment"`
}

// Item-related strings (from `items_strings`) in the user's default language (preferred) or the item's language
type itemString struct {
	*itemStringCommon
	*itemStringNotInfo
}

// Item-related strings (from `items_strings`) in the user's default language (preferred) or the item's language
type itemStringRoot struct {
	*itemStringCommon
	*itemStringNotInfo
	*itemStringRootNodeWithSolutionAccess
}

// from `groups_attempts`
type itemUserActiveAttempt struct {
	// Nullable; only if `can_view` >= 'content'
	AttemptID int64 `json:"attempt_id,string"`
	// only if `can_view` >= 'content'
	ScoreComputed float32 `json:"score_computed"`
	// only if `can_view` >= 'content'
	Submissions int32 `json:"submissions"`
	// only if `can_view` >= 'content'
	Validated bool `json:"validated"`
	// only if `can_view` >= 'content'
	Finished bool `json:"finished"`
	// only if `can_view` >= 'content'
	HintsCached int32 `json:"hints_cached"`
	// Nullable; only if `can_view` >= 'content'
	// example: 2019-09-11T07:30:56Z
	StartedAt *database.Time `json:"started_at,string"`
	// only if `can_view` >= 'content'
	// example: 2019-09-11T07:30:56Z
	// type: string
	ValidatedAt *database.Time `json:"validated_at,string"`
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
	ShowUserInfos bool `json:"show_user_infos"`

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
//              * If the current user has only 'info' access on the specified item, the 'forbidden' error is returned.
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
		return service.ErrForbidden(errors.New("only 'info' access to the item"))
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
	ShowUserInfos   bool
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
	UserActiveAttemptID *int64         `sql:"column:attempt_id"`
	UserScoreComputed   float32        `sql:"column:score_computed"`
	UserSubmissions     int32          `sql:"column:submissions"`
	UserValidated       bool           `sql:"column:validated"`
	UserFinished        bool           `sql:"column:finished"`
	UserHintsCached     int32          `sql:"column:hints_cached"`
	UserStartedAt       *database.Time `sql:"column:started_at"`
	UserValidatedAt     *database.Time `sql:"column:validated_at"`

	// items_items
	Order                  int32 `sql:"column:child_order"`
	Category               string
	ContentViewPropagation string

	CanViewGeneratedValue int
}

// getRawItemData reads data needed by the getItem service from the DB and returns an array of rawItem's
func getRawItemData(s *database.ItemStore, rootID int64, user *database.User) []rawItem {
	var result []rawItem

	accessRights := s.Permissions().WithViewPermissionForUser(user, "info")

	commonColumns := `items.id AS id,
		items.type,
		items.display_details_in_parent,
		items.validation_type,
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
		items.show_user_infos,
		items.url,
		IF(items.type <> 'Chapter', items.uses_api, NULL) AS uses_api,
		IF(items.type <> 'Chapter', items.hints_allowed, NULL) AS hints_allowed,
		NULL AS child_order, NULL AS category, NULL AS content_view_propagation`)

	childrenQuery := s.Select(
		commonColumns+`NULL AS title_bar_visible,
		NULL AS read_only,
		NULL AS full_screen,
		NULL AS show_user_infos,
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
			items.validation_type,
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
			groups_attempts.score_computed AS score_computed,
			groups_attempts.submissions AS submissions,
			groups_attempts.validated AS validated,
			groups_attempts.finished AS finished,
			groups_attempts.hints_cached AS hints_cached,
			groups_attempts.started_at AS started_at,
			groups_attempts.validated_at AS validated_at,

			items.child_order AS child_order,
			items.category AS category,
			items.content_view_propagation, `+
		// inputItem only
		`	items.title_bar_visible,
			items.read_only,
			items.full_screen,
			items.show_user_infos,
			items.url,
			items.uses_api,
			items.hints_allowed,
			access_rights.can_view_generated_value
		FROM ? items `, unionQuery.SubQuery()).
		JoinsUserAndDefaultItemStrings(user).
		Joins("LEFT JOIN users_items ON users_items.item_id=items.id AND users_items.user_id=?", user.GroupID).
		Joins("LEFT JOIN groups_attempts ON groups_attempts.id=users_items.active_attempt_id").
		Joins("JOIN ? access_rights on access_rights.item_id=items.id", accessRights.SubQuery()).
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
	response.ShowUserInfos = (*rawData)[0].ShowUserInfos
}

func constructItemResponseFromDBData(rawData *rawItem, permissionGrantedStore *database.PermissionGrantedStore) *itemResponse {
	result := &itemResponse{
		itemCommonFields: fillItemCommonFieldsWithDBData(rawData),
		String: itemStringRoot{
			itemStringCommon: constructItemStringCommon(rawData),
		},
	}
	result.String.itemStringNotInfo = constructStringNotInfo(rawData, permissionGrantedStore)
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

func constructStringNotInfo(rawData *rawItem, permissionGrantedStore *database.PermissionGrantedStore) *itemStringNotInfo {
	if rawData.CanViewGeneratedValue == permissionGrantedStore.ViewIndexByName("info") {
		return nil
	}
	return &itemStringNotInfo{
		Subtitle:    rawData.StringSubtitle,
		Description: rawData.StringDescription,
	}
}

func constructUserActiveAttempt(rawData *rawItem, permissionGrantedStore *database.PermissionGrantedStore) *itemUserActiveAttempt {
	if rawData.CanViewGeneratedValue == permissionGrantedStore.ViewIndexByName("info") || rawData.UserActiveAttemptID == nil {
		return nil
	}
	return &itemUserActiveAttempt{
		AttemptID:     *rawData.UserActiveAttemptID,
		ScoreComputed: rawData.UserScoreComputed,
		Submissions:   rawData.UserSubmissions,
		Validated:     rawData.UserValidated,
		Finished:      rawData.UserFinished,
		HintsCached:   rawData.UserHintsCached,
		StartedAt:     rawData.UserStartedAt,
		ValidatedAt:   rawData.UserValidatedAt,
	}
}

func fillItemCommonFieldsWithDBData(rawData *rawItem) *itemCommonFields {
	result := &itemCommonFields{
		ID:                       rawData.ID,
		Type:                     rawData.Type,
		DisplayDetailsInParent:   rawData.DisplayDetailsInParent,
		ValidationType:           rawData.ValidationType,
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
		child.String.itemStringNotInfo = constructStringNotInfo(&(*rawData)[index], permissionGrantedStore)
		child.UserActiveAttempt = constructUserActiveAttempt(&(*rawData)[index], permissionGrantedStore)
		child.Order = (*rawData)[index].Order
		child.Category = (*rawData)[index].Category
		child.ContentViewPropagation = (*rawData)[index].ContentViewPropagation
		response.Children = append(response.Children, *child)
	}
}

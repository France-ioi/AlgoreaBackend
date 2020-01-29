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
	LanguageTag string `json:"language_tag"`
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

type itemCommonFields struct {
	// items

	// required: true
	ID int64 `json:"id,string"`
	// required: true
	// enum: Chapter,Task,Course
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
	AllowsMultipleAttempts bool `json:"allows_multiple_attempts"`
	// pattern: ^\d{1,3}:[0-5]?\d:[0-5]?\d$
	// example: 838:59:59
	// Nullable
	// required: true
	Duration *string `json:"duration"`
	// required: true
	NoScore bool `json:"no_score"`
	// required: true
	DefaultLanguageTag string `json:"default_language_tag"`
	// Nullable
	// required: true
	GroupCodeEnter *bool `json:"group_code_enter"`
	// Whether the current user (or the `as_team_id` team) made at least one attempt to solve the item
	// required: true
	HasAttempts bool `json:"has_attempts"`
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
//              and the current user's (or the team's given in `as_team_id`) interactions with them
//              (from tables `items`, `items_items`, `items_string`, `attempts`).
//
//
//              * If the specified item is not visible by the current user (or the team given in `as_team_id`),
//                the 'not found' response is returned.
//
//              * If the current user (or the team given in `as_team_id`) has only 'info' access on the specified item,
//                the 'forbidden' error is returned.
//
//              * If `as_team_id` is given, it should be a user's parent team group,
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
	itemID, err := service.ResolveURLQueryPathInt64Field(httpReq, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)
	groupID := user.GroupID
	if len(httpReq.URL.Query()["as_team_id"]) != 0 {
		groupID, err = service.ResolveURLQueryGetInt64Field(httpReq, "as_team_id")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}

		var found bool
		found, err = srv.Store.Groups().ByID(groupID).Where("type = 'Team'").
			Joins("JOIN groups_groups_active ON groups_groups_active.parent_group_id = groups.id").
			Where("groups_groups_active.child_group_id = ?", user.GroupID).HasRows()
		service.MustNotBeError(err)
		if !found {
			return service.ErrForbidden(errors.New("can't use given as_team_id as a user's team"))
		}
	}

	rawData := getRawItemData(srv.Store.Items(), itemID, groupID, user)

	if len(rawData) == 0 || rawData[0].ID != itemID {
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
	AllowsMultipleAttempts   bool
	Duration                 *string
	NoScore                  bool
	DefaultLanguageTag       string
	GroupCodeEnter           *bool
	HasAttempts              bool

	// root node only
	TitleBarVisible bool
	ReadOnly        bool
	FullScreen      string
	ShowUserInfos   bool
	URL             *string // only if not a chapter
	UsesAPI         bool    // only if not a chapter
	HintsAllowed    bool    // only if not a chapter

	// from items_strings: in the user’s default language or (if not available) default language of the item
	StringLanguageTag string  `sql:"column:language_tag"`
	StringTitle       *string `sql:"column:title"`
	StringImageURL    *string `sql:"column:image_url"`
	StringSubtitle    *string `sql:"column:subtitle"`
	StringDescription *string `sql:"column:description"`
	StringEduComment  *string `sql:"column:edu_comment"`

	// items_items
	Order                  int32 `sql:"column:child_order"`
	Category               string
	ContentViewPropagation string

	CanViewGeneratedValue int
}

// getRawItemData reads data needed by the getItem service from the DB and returns an array of rawItem's
func getRawItemData(s *database.ItemStore, rootID, groupID int64, user *database.User) []rawItem {
	var result []rawItem

	accessRights := s.Permissions().WithViewPermissionForGroup(groupID, "info")

	commonColumns := `items.id AS id,
		items.type,
		items.display_details_in_parent,
		items.validation_type,
		items.contest_entering_condition,
		items.teams_editable,
		items.contest_max_team_size,
		items.allows_multiple_attempts,
		items.duration,
		items.no_score,
		items.default_language_tag,
		items.group_code_enter,
		EXISTS(SELECT 1 FROM attempts WHERE group_id = ? AND item_id = items.id AND started_at IS NOT NULL) AS has_attempts, `

	rootItemQuery := s.ByID(rootID).Select(
		commonColumns+`items.title_bar_visible,
		items.read_only,
		items.full_screen,
		items.show_user_infos,
		items.url,
		IF(items.type <> 'Chapter', items.uses_api, NULL) AS uses_api,
		IF(items.type <> 'Chapter', items.hints_allowed, NULL) AS hints_allowed,
		NULL AS child_order, NULL AS category, NULL AS content_view_propagation`, groupID)

	childrenQuery := s.Select(
		commonColumns+`NULL AS title_bar_visible,
		NULL AS read_only,
		NULL AS full_screen,
		NULL AS show_user_infos,
		NULL AS url,
		NULL AS uses_api,
		NULL AS hints_allowed,
		child_order, category, content_view_propagation`, groupID).
		Joins("JOIN items_items ON items.id=child_item_id AND parent_item_id=?", rootID)

	unionQuery := rootItemQuery.UnionAll(childrenQuery.QueryExpr())
	// This query can be simplified if we add a column for relation degrees into `items_ancestors`
	query := s.Raw(`
		SELECT
			`+commonColumns+`

			COALESCE(user_strings.language_tag, default_strings.language_tag) AS language_tag,
			IF(user_strings.language_tag IS NULL, default_strings.title, user_strings.title) AS title,
			IF(user_strings.language_tag IS NULL, default_strings.image_url, user_strings.image_url) AS image_url,
			IF(user_strings.language_tag IS NULL, default_strings.subtitle, user_strings.subtitle) AS subtitle,
			IF(user_strings.language_tag IS NULL, default_strings.description, user_strings.description) AS description,
			IF(user_strings.language_tag IS NULL, default_strings.edu_comment, user_strings.edu_comment) AS edu_comment,

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
		FROM ? items `, groupID, unionQuery.SubQuery()).
		JoinsUserAndDefaultItemStrings(user).
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
	return result
}

func constructItemStringCommon(rawData *rawItem) *itemStringCommon {
	return &itemStringCommon{
		LanguageTag: rawData.StringLanguageTag,
		Title:       rawData.StringTitle,
		ImageURL:    rawData.StringImageURL,
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

func fillItemCommonFieldsWithDBData(rawData *rawItem) *itemCommonFields {
	result := &itemCommonFields{
		ID:                       rawData.ID,
		Type:                     rawData.Type,
		DisplayDetailsInParent:   rawData.DisplayDetailsInParent,
		ValidationType:           rawData.ValidationType,
		ContestEnteringCondition: rawData.ContestEnteringCondition,
		TeamsEditable:            rawData.TeamsEditable,
		ContestMaxTeamSize:       rawData.ContestMaxTeamSize,
		AllowsMultipleAttempts:   rawData.AllowsMultipleAttempts,
		Duration:                 rawData.Duration,
		NoScore:                  rawData.NoScore,
		DefaultLanguageTag:       rawData.DefaultLanguageTag,
		GroupCodeEnter:           rawData.GroupCodeEnter,
		HasAttempts:              rawData.HasAttempts,
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
		child.Order = (*rawData)[index].Order
		child.Category = (*rawData)[index].Category
		child.ContentViewPropagation = (*rawData)[index].ContentViewPropagation
		response.Children = append(response.Children, *child)
	}
}

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
	// required: true
	Title string `json:"title"`
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

type itemUserNotGrayed struct {
	// Nullable; only if not grayed
	ActiveAttemptID *int64 `json:"active_attempt_id,string"`
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
	// Nullable; only if not grayed; iso8601
	// example: 2019-09-11T07:30:56Z
	StartDate *string `json:"start_date"`
	// only if not grayed; iso8601
	// example: 2019-09-11T07:30:56Z
	ValidationDate *string `json:"validation_date"`
	// Nullable; only if not grayed; iso8601
	// example: 2019-09-11T07:30:56Z
	FinishDate *string `json:"finish_date"`
	// Nullable; only if not grayed; iso8601
	// example: 2019-09-11T07:30:56Z
	ContestStartDate *string `json:"contest_start_date"`
}

type itemUserRootNodeNotChapter struct {
	// Nullable; only if not a chapter
	State *string `json:"state"`
	// Nullable; only if not a chapter
	Answer *string `json:"answer"`
}

// from `users_items`
type itemUser struct {
	*itemUserNotGrayed
}

// from `users_items`
type itemUserRoot struct {
	*itemUserNotGrayed
	*itemUserRootNodeNotChapter
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
	// whether `items.idItemUnlocked` is empty
	// required: true
	HasUnlockedItems bool `json:"has_unlocked_items"`
	// required: true
	ScoreMinUnlock int32 `json:"score_min_unlock"`
	// Nullable
	// required: true
	// enum: All,Half,One,None
	TeamMode *string `json:"team_mode"`
	// required: true
	TeamsEditable bool `json:"teams_editable"`
	// required: true
	TeamMaxMembers int32 `json:"team_max_members"`
	// required: true
	HasAttempts bool `json:"has_attempts"`
	// Nullable; iso8601
	// required: true
	// example: 2019-09-11T07:30:56Z
	AccessOpenDate *string `json:"access_open_date"`
	// pattern: ^\d{1,3}:[0-5]?\d:[0-5]?\d$
	// example: 838:59:59
	// Nullable
	// required: true
	Duration *string `json:"duration"`
	// Nullable; iso8601
	// required: true
	// example: 2019-09-11T07:30:56Z
	EndContestDate *string `json:"end_contest_date"`
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

	// `items_items.iOrder`
	// required: true
	Order int32 `json:"order"`
	// `items_items.sCategory`
	// required: true
	Category string `json:"category"`
	// `items_items.bAlwaysVisible`
	// enum: Undefined,Discovery,Application,Validation,Challenge
	// required: true
	AlwaysVisible bool `json:"always_visible"`
	// `items_items.bAccessRestricted`
	// required: true
	AccessRestricted bool `json:"access_restricted"`

	// from `users_items`
	// required: true
	User itemUser `json:"user"`
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

	// required: true
	User itemUserRoot `json:"user"`

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
//              (from tables `items`, `items_items`, `items_string`, and `users_items`).
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

	if !rawData[0].FullAccess && !rawData[0].PartialAccess {
		return service.ErrForbidden(errors.New("the item is grayed"))
	}

	response := constructItemResponseFromDBData(&rawData[0])

	setItemResponseRootNodeFields(response, &rawData)
	srv.fillItemResponseWithChildren(response, &rawData)

	render.Respond(rw, httpReq, response)
	return service.NoError
}

// rawItem represents one row of the getItem service data returned from the DB
type rawItem struct {
	// items
	ID                     int64   `sql:"column:ID"`
	Type                   string  `sql:"column:sType"`
	DisplayDetailsInParent bool    `sql:"column:bDisplayDetailsInParent"`
	ValidationType         string  `sql:"column:sValidationType"`
	HasUnlockedItems       bool    `sql:"column:hasUnlockedItems"` // whether items.idItemUnlocked is empty
	ScoreMinUnlock         int32   `sql:"column:iScoreMinUnlock"`
	TeamMode               *string `sql:"column:sTeamMode"`
	TeamsEditable          bool    `sql:"column:bTeamsEditable"`
	TeamMaxMembers         int32   `sql:"column:iTeamMaxMembers"`
	HasAttempts            bool    `sql:"column:bHasAttempts"`
	AccessOpenDate         *string `sql:"column:sAccessOpenDate"` // iso8601 str
	Duration               *string `sql:"column:sDuration"`
	EndContestDate         *string `sql:"column:sEndContestDate"` // iso8601 str
	NoScore                bool    `sql:"column:bNoScore"`
	GroupCodeEnter         *bool   `sql:"column:groupCodeEnter"`

	// root node only
	TitleBarVisible bool    `sql:"column:bTitleBarVisible"`
	ReadOnly        bool    `sql:"column:bReadOnly"`
	FullScreen      string  `sql:"column:sFullScreen"`
	ShowSource      bool    `sql:"column:bShowSource"`
	ValidationMin   *int32  `sql:"column:iValidationMin"`
	ShowUserInfos   bool    `sql:"column:bShowUserInfos"`
	ContestPhase    string  `sql:"column:sContestPhase"`
	URL             *string `sql:"column:sUrl"`          // only if not a chapter
	UsesAPI         bool    `sql:"column:bUsesAPI"`      // only if not a chapter
	HintsAllowed    bool    `sql:"column:bHintsAllowed"` // only if not a chapter

	// from items_strings: in the user’s default language or (if not available) default language of the item
	StringLanguageID  int64   `sql:"column:idLanguage"`
	StringTitle       string  `sql:"column:sTitle"`
	StringImageURL    *string `sql:"column:sImageUrl"`
	StringSubtitle    *string `sql:"column:sSubtitle"`
	StringDescription *string `sql:"column:sDescription"`
	StringEduComment  *string `sql:"column:sEduComment"`

	// from users_items for current user
	UserActiveAttemptID     *int64  `sql:"column:idAttemptActive"`
	UserScore               float32 `sql:"column:iScore"`
	UserSubmissionsAttempts int32   `sql:"column:nbSubmissionsAttempts"`
	UserValidated           bool    `sql:"column:bValidated"`
	UserFinished            bool    `sql:"column:bFinished"`
	UserKeyObtained         bool    `sql:"column:bKeyObtained"`
	UserHintsCached         int32   `sql:"column:nbHintsCached"`
	UserStartDate           *string `sql:"column:sStartDate"`        // iso8601 str
	UserValidationDate      *string `sql:"column:sValidationDate"`   // iso8601 str
	UserFinishDate          *string `sql:"column:sFinishDate"`       // iso8601 str
	UserContestStartDate    *string `sql:"column:sContestStartDate"` // iso8601 str
	UserState               *string `sql:"column:sState"`            // only if not a chapter
	UserAnswer              *string `sql:"column:sAnswer"`           // only if not a chapter

	// items_items
	Order            int32  `sql:"column:iChildOrder"`
	Category         string `sql:"column:sCategory"`
	AlwaysVisible    bool   `sql:"column:bAlwaysVisible"`
	AccessRestricted bool   `sql:"column:bAccessRestricted"`

	*database.ItemAccessDetails
}

// getRawItemData reads data needed by the getItem service from the DB and returns an array of rawItem's
func getRawItemData(s *database.ItemStore, rootID int64, user *database.User) []rawItem {
	var result []rawItem

	accessRights := s.AccessRights(user)
	service.MustNotBeError(accessRights.Error())

	commonColumns := `items.ID AS ID,
		items.sType,
		items.bDisplayDetailsInParent,
		items.sValidationType,
		items.idItemUnlocked,
		items.iScoreMinUnlock,
		items.sTeamMode,
		items.bTeamsEditable,
		items.iTeamMaxMembers,
		items.bHasAttempts,
		items.sAccessOpenDate,
		items.sDuration,
		items.sEndContestDate,
		items.bNoScore,
		items.idDefaultLanguage,
		items.groupCodeEnter, `

	rootItemQuery := s.ByID(rootID).Select(
		commonColumns + `items.bTitleBarVisible,
		items.bReadOnly,
		items.sFullScreen,
		items.bShowSource,
		items.iValidationMin,
		items.bShowUserInfos,
		items.sContestPhase,
		items.sUrl,
		IF(items.sType <> 'Chapter', items.bUsesAPI, NULL) AS bUsesAPI,
		IF(items.sType <> 'Chapter', items.bHintsAllowed, NULL) AS bHintsAllowed,
		NULL AS iChildOrder, NULL AS sCategory, NULL AS bAlwaysVisible, NULL AS bAccessRestricted`)

	childrenQuery := s.Select(
		commonColumns+`NULL AS bTitleBarVisible,
		NULL AS bReadOnly,
		NULL AS sFullScreen,
		NULL AS bShowSource,
		NULL AS iValidationMin,
		NULL AS bShowUserInfos,
		NULL AS sContestPhase,
		NULL AS sUrl,
		NULL AS bUsesAPI,
		NULL AS bHintsAllowed,
		iChildOrder, sCategory, bAlwaysVisible, bAccessRestricted`).
		Joins("JOIN items_items ON items.ID=idItemChild AND idItemParent=?", rootID)

	unionQuery := rootItemQuery.UnionAll(childrenQuery.QueryExpr())
	// This query can be simplified if we add a column for relation degrees into `items_ancestors`
	query := s.Raw(`
    SELECT
		  items.ID,
      items.sType,
		  items.bDisplayDetailsInParent,
      items.sValidationType,`+
		// idItemUnlocked is a comma-separated list of item IDs which will be unlocked if this item is validated
		// Here we consider both NULL and an empty string as FALSE
		` COALESCE(items.idItemUnlocked, '')<>'' as hasUnlockedItems,
			items.iScoreMinUnlock,
			items.sTeamMode,
			items.bTeamsEditable,
			items.iTeamMaxMembers,
			items.bHasAttempts,
			items.sAccessOpenDate,
			items.sDuration,
			items.sEndContestDate,
			items.bNoScore,
			items.groupCodeEnter,

			COALESCE(user_strings.idLanguage, default_strings.idLanguage) AS idLanguage,
			IF(user_strings.idLanguage IS NULL, default_strings.sTitle, user_strings.sTitle) AS sTitle,
			IF(user_strings.idLanguage IS NULL, default_strings.sImageUrl, user_strings.sImageUrl) AS sImageUrl,
			IF(user_strings.idLanguage IS NULL, default_strings.sSubtitle, user_strings.sSubtitle) AS sSubtitle,
			IF(user_strings.idLanguage IS NULL, default_strings.sDescription, user_strings.sDescription) AS sDescription,
			IF(user_strings.idLanguage IS NULL, default_strings.sEduComment, user_strings.sEduComment) AS sEduComment,

			users_items.idAttemptActive AS idAttemptActive,
			users_items.iScore AS iScore,
			users_items.nbSubmissionsAttempts AS nbSubmissionsAttempts,
			users_items.bValidated AS bValidated,
			users_items.bFinished AS bFinished,
			users_items.bKeyObtained AS bKeyObtained,
			users_items.nbHintsCached AS nbHintsCached,
			users_items.sStartDate AS sStartDate,
			users_items.sValidationDate AS sValidationDate,
			users_items.sFinishDate AS sFinishDate,
			users_items.sContestStartDate AS sContestStartDate,
			IF(items.sType <> 'Chapter', users_items.sState, NULL) as sState,
			users_items.sAnswer,

			items.iChildOrder AS iChildOrder,
			items.sCategory AS sCategory,
			items.bAlwaysVisible,
			items.bAccessRestricted, `+
		// inputItem only
		` items.bTitleBarVisible,
			items.bReadOnly,
			items.sFullScreen,
			items.bShowSource,
			items.iValidationMin,
			items.bShowUserInfos,
			items.sContestPhase,
			items.sUrl,
			items.bUsesAPI,
			items.bHintsAllowed,
			accessRights.fullAccess, accessRights.partialAccess, accessRights.grayedAccess, accessRights.accessSolutions
    FROM ? items `, unionQuery.SubQuery()).
		JoinsUserAndDefaultItemStrings(user).
		Joins("LEFT JOIN users_items ON users_items.idItem=items.ID AND users_items.idUser=?", user.ID).
		Joins("JOIN ? accessRights on accessRights.idItem=items.ID AND (fullAccess>0 OR partialAccess>0 OR grayedAccess>0)",
			accessRights.SubQuery()).
		Order("iChildOrder")

	service.MustNotBeError(query.Scan(&result).Error())
	return result
}

func setItemResponseRootNodeFields(response *itemResponse, rawData *[]rawItem) {
	if (*rawData)[0].AccessSolutions {
		response.String.itemStringRootNodeWithSolutionAccess = &itemStringRootNodeWithSolutionAccess{
			EduComment: (*rawData)[0].StringEduComment,
		}
	}
	if (*rawData)[0].Type != "Chapter" {
		response.User.itemUserRootNodeNotChapter = &itemUserRootNodeNotChapter{
			State:  (*rawData)[0].UserState,
			Answer: (*rawData)[0].UserAnswer,
		}
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

func constructItemResponseFromDBData(rawData *rawItem) *itemResponse {
	result := &itemResponse{
		itemCommonFields: fillItemCommonFieldsWithDBData(rawData),
		String: itemStringRoot{
			itemStringCommon: constructItemStringCommon(rawData),
		},
	}
	result.String.itemStringNotGrayed = constructStringNotGrayed(rawData)
	result.User.itemUserNotGrayed = constructUserNotGrayed(rawData)
	return result
}

func constructItemStringCommon(rawData *rawItem) *itemStringCommon {
	return &itemStringCommon{
		LanguageID: rawData.StringLanguageID,
		Title:      rawData.StringTitle,
		ImageURL:   rawData.StringImageURL,
	}
}

func constructStringNotGrayed(rawData *rawItem) *itemStringNotGrayed {
	if !rawData.FullAccess && !rawData.PartialAccess {
		return nil
	}
	return &itemStringNotGrayed{
		Subtitle:    rawData.StringSubtitle,
		Description: rawData.StringDescription,
	}
}

func constructUserNotGrayed(rawData *rawItem) *itemUserNotGrayed {
	if !rawData.FullAccess && !rawData.PartialAccess {
		return nil
	}
	return &itemUserNotGrayed{
		ActiveAttemptID:     rawData.UserActiveAttemptID,
		Score:               rawData.UserScore,
		SubmissionsAttempts: rawData.UserSubmissionsAttempts,
		Validated:           rawData.UserValidated,
		Finished:            rawData.UserFinished,
		KeyObtained:         rawData.UserKeyObtained,
		HintsCached:         rawData.UserHintsCached,
		StartDate:           rawData.UserStartDate,
		ValidationDate:      rawData.UserValidationDate,
		FinishDate:          rawData.UserFinishDate,
		ContestStartDate:    rawData.UserContestStartDate,
	}
}

func fillItemCommonFieldsWithDBData(rawData *rawItem) *itemCommonFields {
	result := &itemCommonFields{
		ID:                     rawData.ID,
		Type:                   rawData.Type,
		DisplayDetailsInParent: rawData.DisplayDetailsInParent,
		ValidationType:         rawData.ValidationType,
		HasUnlockedItems:       rawData.HasUnlockedItems,
		ScoreMinUnlock:         rawData.ScoreMinUnlock,
		TeamMode:               rawData.TeamMode,
		TeamsEditable:          rawData.TeamsEditable,
		TeamMaxMembers:         rawData.TeamMaxMembers,
		HasAttempts:            rawData.HasAttempts,
		AccessOpenDate:         rawData.AccessOpenDate,
		Duration:               rawData.Duration,
		EndContestDate:         rawData.EndContestDate,
		NoScore:                rawData.NoScore,
		GroupCodeEnter:         rawData.GroupCodeEnter,
	}
	return result
}

func (srv *Service) fillItemResponseWithChildren(response *itemResponse, rawData *[]rawItem) {
	response.Children = make([]itemChildNode, 0, len(*rawData))
	for index := range *rawData {
		if index == 0 {
			continue
		}

		child := &itemChildNode{itemCommonFields: fillItemCommonFieldsWithDBData(&(*rawData)[index])}
		child.String.itemStringCommon = constructItemStringCommon(&(*rawData)[index])
		child.String.itemStringNotGrayed = constructStringNotGrayed(&(*rawData)[index])
		child.User.itemUserNotGrayed = constructUserNotGrayed(&(*rawData)[index])
		child.Order = (*rawData)[index].Order
		child.Category = (*rawData)[index].Category
		child.AlwaysVisible = (*rawData)[index].AlwaysVisible
		child.AccessRestricted = (*rawData)[index].AccessRestricted
		response.Children = append(response.Children, *child)
	}
}

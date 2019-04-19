package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type itemString struct {
	LanguageID  int64   `json:"language_id,string"`
	Title       string  `json:"title"`
	ImageURL    string  `json:"image_url"`
	Subtitle    *string `json:"subtitle,omitempty"`    // only if not grayed
	Description *string `json:"description,omitempty"` // only if not grayed
	EduComment  *string `json:"edu_comment,omitempty"` // if user has solution access (root node)
}

type itemUser struct {
	// from users_items for current user

	// only if not grayed
	ActiveAttemptID     *int64   `json:"active_attempt_id,omitempty,string"`
	Score               *float32 `json:"score,omitempty"`
	SubmissionsAttempts *int32   `json:"submissions_attempts,omitempty"`
	Validated           *bool    `json:"validated,omitempty"`
	Finished            *bool    `json:"finished,omitempty"`
	KeyObtained         *bool    `json:"key_obtained,omitempty"`
	HintsCached         *int32   `json:"hints_cached,omitempty"`
	StartDate           *string  `json:"start_date,omitempty"`         // iso8601 str
	ValidationDate      *string  `json:"validation_date,omitempty"`    // iso8601 str
	FinishDate          *string  `json:"finish_date,omitempty"`        // iso8601 str
	ContestStartDate    *string  `json:"contest_start_date,omitempty"` // iso8601 str

	// only if not a chapter
	State  *string `json:"state,omitempty"`
	Answer *string `json:"answer,omitempty"`
}

type itemCommonFields struct {
	// items
	ID                     int64  `json:"id,string"`
	Type                   string `json:"type"`
	DisplayDetailsInParent bool   `json:"display_details_in_parent"`
	ValidationType         string `json:"validation_type"`
	HasUnlockedItems       bool   `json:"has_unlocked_items"` // whether items.idItemUnlocked is empty
	ScoreMinUnlock         int32  `json:"score_min_unlock"`
	TeamMode               string `json:"team_mode"`
	TeamsEditable          bool   `json:"teams_editable"`
	TeamMaxMembers         int32  `json:"team_max_members"`
	HasAttempts            bool   `json:"has_attempts"`
	AccessOpenDate         string `json:"access_open_date"` // iso8601 str
	Duration               string `json:"duration"`
	EndContestDate         string `json:"end_contest_date"` // iso8601 str
	NoScore                bool   `json:"no_score"`
	GroupCodeEnter         bool   `json:"group_code_enter"`

	String itemString `json:"string"`
	User   itemUser   `json:"user,omitempty"`

	// root node only
	TitleBarVisible *bool   `json:"title_bar_visible,omitempty"`
	ReadOnly        *bool   `json:"read_only,omitempty"`
	FullScreen      *string `json:"full_screen,omitempty"`
	ShowSource      *bool   `json:"show_source,omitempty"`
	ValidationMin   *int32  `json:"validation_min,omitempty"`
	ShowUserInfos   *bool   `json:"show_user_infos,omitempty"`
	ContestPhase    *string `json:"contest_phase,omitempty"`
	URL             *string `json:"url,omitempty"`           // only if not a chapter
	UsesAPI         *bool   `json:"uses_API,omitempty"`      // only if not a chapter
	HintsAllowed    *bool   `json:"hints_allowed,omitempty"` // only if not a chapter

	// items_items (child nodes only)
	Order            *int32  `json:"order,omitempty"`
	Category         *string `json:"category,omitempty"`
	AlwaysVisible    *bool   `json:"always_visible,omitempty"`
	AccessRestricted *bool   `json:"access_restricted,omitempty"`
}

type itemResponse struct {
	*itemCommonFields
	Children []itemCommonFields `json:"children,omitempty"`
}

func (srv *Service) getItem(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	req := &GetItemRequest{}
	if err := req.Bind(httpReq); err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.GetUser(httpReq)
	err := user.Load() // check that the user exists
	if err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	rawData, err := getRawItemData(srv.Store.Items(), req.ID, user)
	if err != nil {
		return service.ErrUnexpected(err)
	}

	if len(rawData) == 0 || rawData[0].ID != req.ID {
		return service.ErrNotFound(errors.New("insufficient access rights on the given item id"))
	}

	if !rawData[0].FullAccess && !rawData[0].PartialAccess {
		return service.ErrForbidden(errors.New("the item is grayed"))
	}

	response := itemResponse{
		srv.fillItemCommonFieldsWithDBData(&rawData[0]),
		nil,
	}

	setItemResponseRootNodeFields(&response, &rawData)
	srv.fillItemResponseWithChildren(&response, &rawData)

	render.Respond(rw, httpReq, response)
	return service.NoError
}

// rawItem represents one row of the getItem service data returned from the DB
type rawItem struct {
	// items
	ID                     int64  `sql:"column:ID"`
	Type                   string `sql:"column:sType"`
	DisplayDetailsInParent bool   `sql:"column:bDisplayDetailsInParent"`
	ValidationType         string `sql:"column:sValidationType"`
	HasUnlockedItems       bool   `sql:"column:hasUnlockedItems"` // whether items.idItemUnlocked is empty
	ScoreMinUnlock         int32  `sql:"column:iScoreMinUnlock"`
	TeamMode               string `sql:"column:sTeamMode"`
	TeamsEditable          bool   `sql:"column:bTeamsEditable"`
	TeamMaxMembers         int32  `sql:"column:iTeamMaxMembers"`
	HasAttempts            bool   `sql:"column:bHasAttempts"`
	AccessOpenDate         string `sql:"column:sAccessOpenDate"` // iso8601 str
	Duration               string `sql:"column:sDuration"`
	EndContestDate         string `sql:"column:sEndContestDate"` // iso8601 str
	NoScore                bool   `sql:"column:bNoScore"`
	GroupCodeEnter         bool   `sql:"column:groupCodeEnter"`

	// root node only
	TitleBarVisible *bool   `sql:"column:bTitleBarVisible"`
	ReadOnly        *bool   `sql:"column:bReadOnly"`
	FullScreen      *string `sql:"column:sFullScreen"`
	ShowSource      *bool   `sql:"column:bShowSource"`
	ValidationMin   *int32  `sql:"column:iValidationMin"`
	ShowUserInfos   *bool   `sql:"column:bShowUserInfos"`
	ContestPhase    *string `sql:"column:sContestPhase"`
	URL             *string `sql:"column:sUrl"`          // only if not a chapter
	UsesAPI         *bool   `sql:"column:bUsesAPI"`      // only if not a chapter
	HintsAllowed    *bool   `sql:"column:bHintsAllowed"` // only if not a chapter

	// from items_strings: in the user’s default language or (if not available) default language of the item
	StringLanguageID  int64  `sql:"column:idLanguage"`
	StringTitle       string `sql:"column:sTitle"`
	StringImageURL    string `sql:"column:sImageUrl"`
	StringSubtitle    string `sql:"column:sSubtitle"`
	StringDescription string `sql:"column:sDescription"`
	StringEduComment  string `sql:"column:sEduComment"`

	// from users_items for current user
	UserActiveAttemptID     int64   `sql:"column:idAttemptActive"`
	UserScore               float32 `sql:"column:iScore"`
	UserSubmissionsAttempts int32   `sql:"column:nbSubmissionsAttempts"`
	UserValidated           bool    `sql:"column:bValidated"`
	UserFinished            bool    `sql:"column:bFinished"`
	UserKeyObtained         bool    `sql:"column:bKeyObtained"`
	UserHintsCached         int32   `sql:"column:nbHintsCached"`
	UserStartDate           string  `sql:"column:sStartDate"`        // iso8601 str
	UserValidationDate      string  `sql:"column:sValidationDate"`   // iso8601 str
	UserFinishDate          string  `sql:"column:sFinishDate"`       // iso8601 str
	UserContestStartDate    string  `sql:"column:sContestStartDate"` // iso8601 str
	UserState               *string `sql:"column:sState"`            // only if not a chapter
	UserAnswer              *string `sql:"column:sAnswer"`           // only if not a chapter

	// items_items
	Order            int32   `sql:"column:iChildOrder"`
	Category         *string `sql:"column:sCategory"`
	AlwaysVisible    *bool   `sql:"column:bAlwaysVisible"`
	AccessRestricted *bool   `sql:"column:bAccessRestricted"`

	*database.ItemAccessDetails
}

// getRawItemData reads data needed by the getItem service from the DB and returns an array of rawItem's
func getRawItemData(s *database.ItemStore, rootID int64, user *database.User) ([]rawItem, error) {
	var result []rawItem

	accessRights := s.AccessRights(user)
	service.MustNotBeError(accessRights.Error()) // we have already checked that the user exists in getItem()

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
		Joins("LEFT JOIN users_items ON users_items.idItem=items.ID AND users_items.idUser=?", user.UserID).
		Joins("JOIN ? accessRights on accessRights.idItem=items.ID AND (fullAccess>0 OR partialAccess>0 OR grayedAccess>0)",
			accessRights.SubQuery()).
		Order("iChildOrder")

	service.MustNotBeError(query.Scan(&result).Error())
	return result, nil
}

func setItemResponseRootNodeFields(response *itemResponse, rawData *[]rawItem) {
	if (*rawData)[0].AccessSolutions {
		response.String.EduComment = &((*rawData)[0].StringEduComment)
	}
	response.User.State = (*rawData)[0].UserState
	response.User.Answer = (*rawData)[0].UserAnswer
	response.TitleBarVisible = (*rawData)[0].TitleBarVisible
	response.ReadOnly = (*rawData)[0].ReadOnly
	response.FullScreen = (*rawData)[0].FullScreen
	response.ShowSource = (*rawData)[0].ShowSource
	response.ValidationMin = (*rawData)[0].ValidationMin
	response.ShowUserInfos = (*rawData)[0].ShowUserInfos
	response.ContestPhase = (*rawData)[0].ContestPhase
	response.URL = (*rawData)[0].URL
	response.UsesAPI = (*rawData)[0].UsesAPI
	response.HintsAllowed = (*rawData)[0].HintsAllowed
}

func (srv *Service) fillItemCommonFieldsWithDBData(rawData *rawItem) *itemCommonFields {
	result := itemCommonFields{
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

		String: itemString{
			LanguageID: rawData.StringLanguageID,
			Title:      rawData.StringTitle,
			ImageURL:   rawData.StringImageURL,
		},
	}
	if rawData.FullAccess || rawData.PartialAccess {
		result.String.Subtitle = &rawData.StringSubtitle
		result.String.Description = &rawData.StringDescription

		result.User.ActiveAttemptID = &rawData.UserActiveAttemptID
		result.User.Score = &rawData.UserScore
		result.User.SubmissionsAttempts = &rawData.UserSubmissionsAttempts
		result.User.Validated = &rawData.UserValidated
		result.User.Finished = &rawData.UserFinished
		result.User.KeyObtained = &rawData.UserKeyObtained
		result.User.HintsCached = &rawData.UserHintsCached
		result.User.StartDate = &rawData.UserStartDate
		result.User.ValidationDate = &rawData.UserValidationDate
		result.User.FinishDate = &rawData.UserFinishDate
		result.User.ContestStartDate = &rawData.UserContestStartDate
	}
	return &result
}

func (srv *Service) fillItemResponseWithChildren(response *itemResponse, rawData *[]rawItem) {
	for index := range *rawData {
		if index == 0 {
			continue
		}

		child := srv.fillItemCommonFieldsWithDBData(&(*rawData)[index])
		child.Order = &(*rawData)[index].Order
		child.Category = (*rawData)[index].Category
		child.AlwaysVisible = (*rawData)[index].AlwaysVisible
		child.AccessRestricted = (*rawData)[index].AccessRestricted
		response.Children = append(response.Children, *child)
	}
}

package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type itemString struct {
	LanguageId         		int64   `json:"language_id"`
	Title         				string  `json:"title"`
	ImageUrl         			string  `json:"image_url"`
	Subtitle         			*string `json:"subtitle,omitempty"`    // only if not grayed
	Description         	*string `json:"description,omitempty"` // only if not grayed
	EduComment           	*string `json:"edu_comment,omitempty"` // if user has solution access (root node)
}

type itemUser struct {
	// from users_items for current user

	// only if not grayed
	ActiveAttemptId     *int64   `json:"active_attempt_id,omitempty"`
	Score 						  *float32 `json:"score,omitempty"`
	SubmissionsAttempts *int64   `json:"submissions_attempts,omitempty"`
	Validated 				  *bool	   `json:"validated,omitempty"`
	Finished					  *bool	   `json:"finished,omitempty"`
	KeyObtained 			  *bool 	 `json:"key_obtained,omitempty"`
	HintsCached         *int64   `json:"hints_cached,omitempty"`
	StartDate           *string  `json:"start_date,omitempty"` // iso8601 str
	ValidationDate      *string  `json:"validation_date,omitempty"` // iso8601 str
	FinishDate          *string  `json:"finish_date,omitempty"` // iso8601 str
	ContestStartDate    *string  `json:"contest_start_date,omitempty"` // iso8601 str

	// only if not a chapter
	State               *string  `json:"state,omitempty"`
	Answer              *string  `json:"answer,omitempty"`
}

type itemCommonFields struct {
	// items
	ID                		 int64    `json:"id"`
	Type              		 string   `json:"type"`
	DisplayDetailsInParent bool	    `json:"display_details_in_parent"`
	ValidationType         string   `json:"validation_type"`
	HasUnlockedItems  		 bool     `json:"has_unlocked_items"` // whether items.idItemUnlocked is empty
	ScoreMinUnlock         int64    `json:"score_min_unlock"`
	TeamMode               string   `json:"team_mode"`
	TeamsEditable          bool	    `json:"teams_editable"`
	TeamMaxMembers         int64    `json:"team_max_members"`
	HasAttempts            bool	    `json:"has_attempts"`
	AccessOpenDate         string   `json:"access_open_date"` // iso8601 str
	Duration               string   `json:"duration"`
	EndContestDate         string   `json:"end_contest_date"` // iso8601 str
	NoScore                bool     `json:"no_score"`
	GroupCodeEnter         bool     `json:"group_code_enter"`

	String                 itemString `json:"string"`
	User                   itemUser   `json:"user,omitempty"`

	// root node only
	TitleBarVisible        *bool    `json:"title_bar_visible,omitempty"`
	ReadOnly               *bool    `json:"read_only,omitempty"`
	FullScreen             *string  `json:"full_screen,omitempty"`
	ShowSource             *bool    `json:"show_source,omitempty"`
	ValidationMin          *int64   `json:"validation_min,omitempty"`
	ShowUserInfos          *bool    `json:"show_user_infos,omitempty"`
	ContestPhase           *string  `json:"contest_phase,omitempty"`
	Url                    *string  `json:"url,omitempty"` // only if not a chapter
	UsesAPI                *bool    `json:"uses_API,omitempty"` // only if not a chapter
	HintsAllowed           *bool    `json:"hints_allowed,omitempty"` // only if not a chapter

	// items_items (child nodes only)
	Order 						    *int64 	 `json:"order,omitempty"`
	Category 						  *string  `json:"category,omitempty"`
	AlwaysVisible 				*bool    `json:"always_visible,omitempty"`
	AccessRestricted 			*bool    `json:"access_restricted,omitempty"`
}

type itemResponse struct {
	*itemCommonFields
	Children							[]itemCommonFields `json:"children,omitempty"`
}

func (srv *Service) getItem(rw http.ResponseWriter, httpReq *http.Request) service.APIError {
	req := &GetItemRequest{}
	if err := req.Bind(httpReq); err != nil {
		return service.ErrInvalidRequest(err)
	}

	user := srv.getUser(httpReq)
	rawData, err := srv.Store.Items().GetRawItemData(req.ID, user.UserID, user.DefaultLanguageID(), user)
	if err != nil {
		return service.ErrUnexpected(err)
	}

	if len(*rawData) == 0 || (*rawData)[0].ID != req.ID {
		return service.ErrNotFound(errors.New("insufficient access rights on the given item id"))
	}

	if !(*rawData)[0].FullAccess && !(*rawData)[0].PartialAccess {
		return service.ErrForbidden(errors.New("the item is grayed"))
	}

	response := itemResponse{
		srv.fillItemCommonFieldsWithDBData(&(*rawData)[0]),
		nil,
	}

	setItemResponseRootNodeFields(&response, rawData)
	srv.fillItemResponseWithChildren(&response, rawData)

	render.Respond(rw, httpReq, response)
	return service.NoError
}

func setItemResponseRootNodeFields(response *itemResponse, rawData *[]database.RawItem) {
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
	response.Url = (*rawData)[0].Url
	response.UsesAPI = (*rawData)[0].UsesAPI
	response.HintsAllowed = (*rawData)[0].HintsAllowed
}

func (srv *Service) getAccessDetailsForRawItems(rawData *[]database.RawItem, user *auth.User,
) (map[int64]database.ItemAccessDetails, error) {
	var ids []int64
	for _, row := range *rawData {
		ids = append(ids, row.ID)
	}
	accessDetailsMap, err := srv.Store.Items().GetAccessDetailsMapForIDs(user, ids)
	return accessDetailsMap, err
}

func (srv *Service) fillItemCommonFieldsWithDBData(rawData *database.RawItem) *itemCommonFields {
	result := itemCommonFields{
		ID: rawData.ID,
		Type: rawData.Type,
		DisplayDetailsInParent: rawData.DisplayDetailsInParent,
		ValidationType: rawData.ValidationType,
		HasUnlockedItems: rawData.HasUnlockedItems,
		ScoreMinUnlock: rawData.ScoreMinUnlock,
		TeamMode: rawData.TeamMode,
		TeamsEditable: rawData.TeamsEditable,
		TeamMaxMembers: rawData.TeamMaxMembers,
		HasAttempts: rawData.HasAttempts,
		AccessOpenDate: rawData.AccessOpenDate,
		Duration: rawData.Duration,
		EndContestDate: rawData.EndContestDate,
		NoScore: rawData.NoScore,
		GroupCodeEnter: rawData.GroupCodeEnter,

		String: itemString{
			LanguageId: rawData.StringLanguageId,
			Title: rawData.StringTitle,
			ImageUrl: rawData.StringImageUrl,
		},
	}
	if rawData.FullAccess || rawData.PartialAccess {
		result.String.Subtitle = &rawData.StringSubtitle
		result.String.Description = &rawData.StringDescription

		result.User.ActiveAttemptId = &rawData.UserActiveAttemptId
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

func (srv *Service) fillItemResponseWithChildren(response *itemResponse, rawData *[]database.RawItem) {
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

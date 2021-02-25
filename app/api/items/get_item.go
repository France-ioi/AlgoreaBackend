package items

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/structures"
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
type itemStringRoot struct {
	*itemStringCommon
	*itemStringNotInfo
	*itemStringRootNodeWithSolutionAccess
}

type commonItemFields struct {
	// items

	// required: true
	ID int64 `json:"id,string"`
	// required: true
	// enum: Chapter,Task,Course,Skill
	Type string `json:"type"`
	// required: true
	DisplayDetailsInParent bool `json:"display_details_in_parent"`
	// required: true
	// enum: None,All,AllButOne,Categories,One,Manual
	ValidationType string `json:"validation_type"`
	// required: true
	RequiresExplicitEntry bool `json:"requires_explicit_entry"`
	// required: true
	AllowsMultipleAttempts bool `json:"allows_multiple_attempts"`
	// required: true
	// enum: User,Team
	EntryParticipantType string `json:"entry_participant_type"`
	// pattern: ^\d{1,3}:[0-5]?\d:[0-5]?\d$
	// example: 838:59:59
	// Nullable
	// required: true
	Duration *string `json:"duration"`
	// required: true
	NoScore bool `json:"no_score"`
	// required: true
	DefaultLanguageTag string `json:"default_language_tag"`

	// required: true
	Permissions structures.ItemPermissions `json:"permissions"`
}

type itemRootNodeNotChapterFields struct {
	// Nullable; only if not a chapter
	URL *string `json:"url"`
	// only if not a chapter
	UsesAPI bool `json:"uses_api"`
	// only if not a chapter
	HintsAllowed bool `json:"hints_allowed"`
}

// swagger:model itemResponse
type itemResponse struct {
	*commonItemFields

	// required: true
	// enum: All,Half,One,None
	EntryMinAdmittedMembersRatio string `json:"entry_min_admitted_members_ratio"`
	// required: true
	EntryFrozenTeams bool `json:"entry_frozen_teams"`
	// required: true
	EntryMaxTeamSize int32 `json:"entry_max_team_size"`
	// required: true
	PromptToJoinGroupByCode bool `json:"prompt_to_join_group_by_code"`
	// required: true
	TitleBarVisible bool `json:"title_bar_visible"`
	// required: true
	ReadOnly bool `json:"read_only"`
	// required: true
	// enum: forceYes,forceNo,default
	FullScreen string `json:"full_screen"`
	// required: true
	ShowUserInfos bool `json:"show_user_infos"`
	// required: true
	EnteringTimeMin time.Time `json:"entering_time_min"`
	// required: true
	EnteringTimeMax time.Time `json:"entering_time_max"`

	// max among all attempts of the user (or of the team given in `{as_team_id}`)
	// required: true
	BestScore float32 `json:"best_score"`

	// required: true
	String itemStringRoot `json:"string"`

	*itemRootNodeNotChapterFields
}

// swagger:operation GET /items/{item_id} items itemView
// ---
// summary: Get an item
// description: Returns data related to the specified item,
//              and the current user's (or the team's given in `as_team_id`) permissions on it
//              (from tables `items`, `items_string`, `permissions_generated`).
//
//
//              * If the specified item is not visible by the current user (or the team given in `as_team_id`),
//                the 'not found' response is returned.
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
	participantID := service.ParticipantIDFromContext(httpReq.Context())

	rawData := getRawItemData(srv.Store.Items(), itemID, participantID, user)
	if rawData == nil {
		return service.ErrNotFound(errors.New("insufficient access rights on the given item id"))
	}

	permissionGrantedStore := srv.Store.PermissionsGranted()
	response := constructItemResponseFromDBData(rawData, permissionGrantedStore)

	render.Respond(rw, httpReq, response)
	return service.NoError
}

// rawItem represents the getItem service data returned from the DB
type rawItem struct {
	*RawCommonItemFields

	// items
	TitleBarVisible              bool
	ReadOnly                     bool
	FullScreen                   string
	ShowUserInfos                bool
	EntryMinAdmittedMembersRatio string
	EntryFrozenTeams             bool
	EntryMaxTeamSize             int32
	PromptToJoinGroupByCode      bool
	URL                          *string // only if not a chapter
	UsesAPI                      bool    // only if not a chapter
	HintsAllowed                 bool    // only if not a chapter
	BestScore                    float32

	// from items_strings: in the user’s default language or (if not available) default language of the item
	StringLanguageTag string  `sql:"column:language_tag"`
	StringTitle       *string `sql:"column:title"`
	StringImageURL    *string `sql:"column:image_url"`
	StringSubtitle    *string `sql:"column:subtitle"`
	StringDescription *string `sql:"column:description"`
	StringEduComment  *string `sql:"column:edu_comment"`
}

// getRawItemData reads data needed by the getItem service from the DB and returns an array of rawItem's
func getRawItemData(s *database.ItemStore, rootID, groupID int64, user *database.User) *rawItem {
	var result rawItem

	query := s.ByID(rootID).
		JoinsPermissionsForGroupToItemsWherePermissionAtLeast(groupID, "view", "info").
		JoinsUserAndDefaultItemStrings(user).
		Select(`
			items.id AS id,
			items.type,
			items.display_details_in_parent,
			items.validation_type,
			items.entry_min_admitted_members_ratio,
			items.entry_frozen_teams,
			items.entry_max_team_size,
			items.entering_time_min,
			items.entering_time_max,
			items.allows_multiple_attempts,
			items.entry_participant_type,
			items.duration,
			items.no_score,
			items.default_language_tag,
			items.prompt_to_join_group_by_code,
			items.title_bar_visible,
			items.read_only,
			items.full_screen,
			items.show_user_infos,
			items.url,
			items.requires_explicit_entry,
			IF(items.type <> 'Chapter', items.uses_api, NULL) AS uses_api,
			IF(items.type <> 'Chapter', items.hints_allowed, NULL) AS hints_allowed,
			COALESCE(user_strings.language_tag, default_strings.language_tag) AS language_tag,
			IF(user_strings.language_tag IS NULL, default_strings.title, user_strings.title) AS title,
			IF(user_strings.language_tag IS NULL, default_strings.image_url, user_strings.image_url) AS image_url,
			IF(user_strings.language_tag IS NULL, default_strings.subtitle, user_strings.subtitle) AS subtitle,
			IF(user_strings.language_tag IS NULL, default_strings.description, user_strings.description) AS description,
			IF(user_strings.language_tag IS NULL, default_strings.edu_comment, user_strings.edu_comment) AS edu_comment,
			can_view_generated_value, can_grant_view_generated_value, can_watch_generated_value, can_edit_generated_value, is_owner_generated,
			IFNULL(
					(SELECT MAX(results.score_computed) AS best_score
					FROM results
					WHERE results.item_id = items.id AND results.participant_id = ?), 0) AS best_score`, groupID)

	err := query.Scan(&result).Error()
	if gorm.IsRecordNotFoundError(err) {
		return nil
	}
	service.MustNotBeError(err)
	return &result
}

func constructItemResponseFromDBData(rawData *rawItem, permissionGrantedStore *database.PermissionGrantedStore) *itemResponse {
	result := &itemResponse{
		commonItemFields: rawData.asItemCommonFields(permissionGrantedStore),
		String: itemStringRoot{
			itemStringCommon: constructItemStringCommon(rawData),
		},
		EntryMinAdmittedMembersRatio: rawData.EntryMinAdmittedMembersRatio,
		EntryFrozenTeams:             rawData.EntryFrozenTeams,
		EntryMaxTeamSize:             rawData.EntryMaxTeamSize,
		PromptToJoinGroupByCode:      rawData.PromptToJoinGroupByCode,
		TitleBarVisible:              rawData.TitleBarVisible,
		ReadOnly:                     rawData.ReadOnly,
		FullScreen:                   rawData.FullScreen,
		ShowUserInfos:                rawData.ShowUserInfos,
		EnteringTimeMin:              time.Time(rawData.EnteringTimeMin),
		EnteringTimeMax:              time.Time(rawData.EnteringTimeMax),
		BestScore:                    rawData.BestScore,
	}
	result.String.itemStringNotInfo = constructStringNotInfo(rawData, permissionGrantedStore)

	if rawData.CanViewGeneratedValue == permissionGrantedStore.ViewIndexByName("solution") {
		result.String.itemStringRootNodeWithSolutionAccess = &itemStringRootNodeWithSolutionAccess{
			EduComment: rawData.StringEduComment,
		}
	}
	if rawData.Type != "Chapter" {
		result.itemRootNodeNotChapterFields = &itemRootNodeNotChapterFields{
			URL:          rawData.URL,
			UsesAPI:      rawData.UsesAPI,
			HintsAllowed: rawData.HintsAllowed,
		}
	}

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

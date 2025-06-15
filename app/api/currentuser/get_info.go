package currentuser

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model userData
type getInfoData struct {
	// required: true
	GroupID int64 `json:"group_id,string"`
	// required: true
	TempUser bool `json:"temp_user"`
	// required: true
	Login string `json:"login"`
	// required: true
	RegisteredAt *database.Time `json:"registered_at"`
	// required: true
	LatestProfileSyncAt *database.Time `json:"latest_profile_sync_at"`
	// required: true
	Email *string `json:"email"`
	// required: true
	EmailVerified bool `json:"email_verified"`
	// required: true
	FirstName *string `json:"first_name"`
	// required: true
	LastName *string `json:"last_name"`
	// required: true
	StudentID *string `json:"student_id"`
	// required: true
	CountryCode string `json:"country_code"`
	// required: true
	TimeZone *string `json:"time_zone"`
	// required: true
	BirthDate *string `json:"birth_date"`
	// required: true
	GraduationYear int32 `json:"graduation_year"`
	// required: true
	Grade *int32 `json:"grade"`
	// required: true
	// enum: Male,Female
	Sex *string `json:"sex"`
	// required: true
	Address *string `json:"address"`
	// required: true
	ZipCode *string `gorm:"column:zipcode" json:"zip_code"`
	// required: true
	City *string `json:"city"`
	// required: true
	LandLineNumber *string `json:"land_line_number"`
	// required: true
	CellPhoneNumber *string `json:"cell_phone_number"`
	// required: true
	DefaultLanguage string `json:"default_language"`
	// required: true
	PublicFirstName bool `json:"public_first_name"`
	// required: true
	PublicLastName bool `json:"public_last_name"`

	// required: true
	NotifyNews bool `json:"notify_news"`
	// required: true
	// enum: Never,Answers,Concerned
	Notify string `json:"notify"`
	// required: true
	FreeText *string `json:"free_text"`
	// required: true
	WebSite *string `json:"web_site"`
	// required: true
	PhotoAutoload bool `json:"photo_autoload"`
	// required: true
	LangProg *string `json:"lang_prog"`
	// required: true
	BasicEditorMode bool `json:"basic_editor_mode"`
	// required: true
	SpacesForTab int32 `json:"spaces_for_tab"`
	// required: true
	StepLevelInSite int32 `json:"step_level_in_site"`
	// required: true
	IsAdmin bool `json:"is_admin"`
	// required: true
	NoRanking bool `json:"no_ranking"`
}

// swagger:operation GET /current-user users userData
//
//	---
//	summary: Get profile info for the current user
//	description: Returns the data from the `users` table.
//	responses:
//		"200":
//				description: OK. Success response with user's data
//				schema:
//					"$ref": "#/definitions/userData"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getInfo(w http.ResponseWriter, r *http.Request) *service.APIError {
	user := srv.GetUser(r)

	var userInfo getInfoData
	err := srv.GetStore(r).Users().ByID(user.GroupID).
		Select(`group_id, temp_user, login, registered_at, email, email_verified, first_name, last_name,
			student_id, country_code, time_zone, latest_profile_sync_at,
			CONVERT(birth_date, char) AS birth_date, graduation_year, grade, sex, address, zipcode,
			city, land_line_number, cell_phone_number, default_language, public_first_name, public_last_name,
			notify_news, notify, free_text, web_site, photo_autoload, lang_prog, basic_editor_mode, spaces_for_tab,
			step_level_in_site, is_admin, no_ranking`).
		Scan(&userInfo).Error()

	// This is very unlikely since the user middleware has already checked that the user exists
	if err == gorm.ErrRecordNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	render.Respond(w, r, &userInfo)
	return service.NoError
}

package currentuser

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:model userData
type getInfoData struct {
	// required: true
	ID int64 `gorm:"column:ID" json:"id,string"`
	// required: true
	TempUser bool `gorm:"column:tempUser" json:"temp_user"`
	// required: true
	Login string `gorm:"column:sLogin" json:"login"`
	// Nullable
	// required: true
	RegistrationDate *string `gorm:"column:sRegistrationDate" json:"registration_date"`
	// Nullable
	// required: true
	Email *string `gorm:"column:sEmail" json:"email"`
	// required: true
	EmailVerified bool `gorm:"column:bEmailVerified" json:"email_verified"`
	// Nullable
	// required: true
	FirstName *string `gorm:"column:sFirstName" json:"first_name"`
	// Nullable
	// required: true
	LastName *string `gorm:"column:sLastName" json:"last_name"`
	// Nullable
	// required: true
	StudentID *string `gorm:"column:sStudentId" json:"student_id"`
	// required: true
	CountryCode string `gorm:"column:sCountryCode" json:"country_code"`
	// Nullable
	// required: true
	TimeZone *string `gorm:"column:sTimeZone" json:"time_zone"`
	// Nullable
	// required: true
	BirthDate *string `gorm:"column:sBirthDate" json:"birth_date"`
	// required: true
	GraduationYear int32 `gorm:"column:iGraduationYear" json:"graduation_year"`
	// Nullable
	// required: true
	Grade *int32 `gorm:"column:iGrade" json:"grade"`
	// Nullable
	// required: true
	// enum: Male,Female
	Sex *string `gorm:"column:sSex" json:"sex"`
	// Nullable
	// required: true
	Address *string `gorm:"column:sAddress" json:"address"`
	// Nullable
	// required: true
	ZipCode *string `gorm:"column:sZipcode" json:"zip_code"`
	// Nullable
	// required: true
	City *string `gorm:"column:sCity" json:"city"`
	// Nullable
	// required: true
	LandLineNumber *string `gorm:"column:sLandLineNumber" json:"land_line_number"`
	// Nullable
	// required: true
	CellPhoneNumber *string `gorm:"column:sCellPhoneNumber" json:"cell_phone_number"`
	// required: true
	DefaultLanguage string `gorm:"column:sDefaultLanguage" json:"default_language"`
	// required: true
	PublicFirstName bool `gorm:"column:bPublicFirstName" json:"public_first_name"`
	// required: true
	PublicLastName bool `gorm:"column:bPublicLastName" json:"public_last_name"`

	// required: true
	NotifyNews bool `gorm:"column:bNotifyNews" json:"notify_news"`
	// required: true
	// enum: Never,Answers,Concerned
	Notify string `gorm:"column:sNotify" json:"notify"`
	// Nullable
	// required: true
	FreeText *string `gorm:"column:sFreeText" json:"free_text"`
	// Nullable
	// required: true
	WebSite *string `gorm:"column:sWebSite" json:"web_site"`
	// required: true
	PhotoAutoload bool `gorm:"column:bPhotoAutoload" json:"photo_autoload"`
	// Nullable
	// required: true
	LangProg *string `gorm:"column:sLangProg" json:"lang_prog"`
	// required: true
	BasicEditorMode bool `gorm:"column:bBasicEditorMode" json:"basic_editor_mode"`
	// required: true
	SpacesForTab int32 `gorm:"column:nbSpacesForTab" json:"spaces_for_tab"`
	// required: true
	StepLevelInSite int32 `gorm:"column:iStepLevelInSite" json:"step_level_in_site"`
	// required: true
	IsAdmin bool `gorm:"column:bIsAdmin" json:"is_admin"`
	// required: true
	NoRanking bool `gorm:"column:bNoRanking" json:"no_ranking"`
	// Nullable
	// required: true
	LoginModulePrefix *string `gorm:"column:loginModulePrefix" json:"login_module_prefix"`
	// Nullable
	// required: true
	AllowSubgroups *bool `gorm:"column:allowSubgroups" json:"allow_subgroups"`
}

// swagger:operation GET /current-user users userData
// ---
// summary: Get profile info for the current user
// description: Returns the data from the `users` table.
// responses:
//   "200":
//     description: OK. Success response with user's data
//     schema:
//       "$ref": "#/definitions/userData"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getInfo(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	var userInfo getInfoData
	err := srv.Store.Users().ByID(user.ID).
		Select(`ID, tempUser, sLogin, sRegistrationDate, sEmail, bEmailVerified, sFirstName, sLastName,
			sStudentId, sCountryCode, sTimeZone,
			CONVERT(sBirthDate, char) AS sBirthDate, iGraduationYear, iGrade, sSex, sAddress, sZipcode,
			sCity, sLandLineNumber, sCellPhoneNumber, sDefaultLanguage, bPublicFirstName, bPublicLastName,
			bNotifyNews, sNotify, sFreeText, sWebSite, bPhotoAutoload, sLangProg, bBasicEditorMode, nbSpacesForTab,
			iStepLevelInSite, bIsAdmin, bNoRanking, loginModulePrefix, allowSubgroups`).
		Scan(&userInfo).Error()

	// This is very unlikely since the user middleware has already checked that the user exists
	if err == gorm.ErrRecordNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	render.Respond(w, r, &userInfo)
	return service.NoError
}

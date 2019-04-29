package currentuser

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getInfo(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	var userInfo struct {
		ID               int64   `gorm:"column:ID" json:"id,string"`
		TempUser         bool    `gorm:"column:tempUser" json:"temp_user"`
		Login            string  `gorm:"column:sLogin" json:"login"`
		RegistrationDate *string `gorm:"column:sRegistrationDate" json:"registration_date,omitempty"`
		Email            *string `gorm:"column:sEmail" json:"email,omitempty"`
		EmailVerified    bool    `gorm:"column:bEmailVerified" json:"email_verified,omitempty"`
		FirstName        *string `gorm:"column:sFirstName" json:"first_name,omitempty"`
		LastName         *string `gorm:"column:sLastName" json:"last_name,omitempty"`
		StudentID        *string `gorm:"column:sStudentId" json:"student_id,omitempty"`
		CountryCode      string  `gorm:"column:sCountryCode" json:"country_code"`
		TimeZone         *string `gorm:"column:sTimeZone" json:"time_zone,omitempty"`
		BirthDate        *string `gorm:"column:sBirthDate" json:"birth_date,omitempty"`
		GraduationYear   int32   `gorm:"column:iGraduationYear" json:"graduation_year"`
		Grade            *int32  `gorm:"column:iGrade" json:"grade,omitempty"`
		Sex              *string `gorm:"column:sSex" json:"sex,omitempty"`
		Address          *string `gorm:"column:sAddress" json:"address,omitempty"`
		ZipCode          *string `gorm:"column:sZipcode" json:"zip_code,omitempty"`
		City             *string `gorm:"column:sCity" json:"city,omitempty"`
		LandLineNumber   *string `gorm:"column:sLandLineNumber" json:"land_line_number,omitempty"`
		CellPhoneNumber  *string `gorm:"column:sCellPhoneNumber" json:"cell_phone_number,omitempty"`
		DefaultLanguage  string  `gorm:"column:sDefaultLanguage" json:"default_language"`
		PublicFirstName  bool    `gorm:"column:bPublicFirstName" json:"public_first_name"`
		PublicLastName   bool    `gorm:"column:bPublicLastName" json:"public_last_name"`

		NotifyNews        bool    `gorm:"column:bNotifyNews" json:"notify_news"`
		Notify            string  `gorm:"column:sNotify" json:"notify"`
		FreeText          *string `gorm:"column:sFreeText" json:"free_text,omitempty"`
		WebSite           *string `gorm:"column:sWebSite" json:"web_site,omitempty"`
		PhotoAutoload     bool    `gorm:"column:bPhotoAutoload" json:"photo_autoload"`
		LangProg          *string `gorm:"column:sLangProg" json:"lang_prog,omitempty"`
		BasicEditorMode   bool    `gorm:"column:bBasicEditorMode" json:"basic_editor_mode"`
		SpacesForTab      int32   `gorm:"column:nbSpacesForTab" json:"spaces_for_tab"`
		StepLevelInSite   int32   `gorm:"column:iStepLevelInSite" json:"step_level_in_site"`
		IsAdmin           bool    `gorm:"column:bIsAdmin" json:"is_admin"`
		NoRanking         bool    `gorm:"column:bNoRanking" json:"no_ranking"`
		LoginModulePrefix *string `gorm:"column:loginModulePrefix" json:"login_module_prefix,omitempty"`
		AllowSubgroups    *bool   `gorm:"column:allowSubgroups" json:"allow_subgroups,omitempty"`
	}

	err := srv.Store.Users().ByID(user.UserID).
		Select(`ID, tempUser, sLogin, sRegistrationDate, sEmail, bEmailVerified, sFirstName, sLastName,
			sStudentId, sCountryCode, sTimeZone,
			CONVERT(sBirthDate, char) AS sBirthDate, iGraduationYear, iGrade, sSex, sAddress, sZipcode,
			sCity, sLandLineNumber, sCellPhoneNumber, sDefaultLanguage, bPublicFirstName, bPublicLastName,
			bNotifyNews, sNotify, sFreeText, sWebSite, bPhotoAutoload, sLangProg, bBasicEditorMode, nbSpacesForTab,
			iStepLevelInSite, bIsAdmin, bNoRanking, loginModulePrefix, allowSubgroups`).
		Scan(&userInfo).Error()
	if err == gorm.ErrRecordNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

	render.Respond(w, r, &userInfo)
	return service.NoError
}

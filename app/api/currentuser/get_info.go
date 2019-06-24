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
		RegistrationDate *string `gorm:"column:sRegistrationDate" json:"registration_date"`
		Email            *string `gorm:"column:sEmail" json:"email"`
		EmailVerified    bool    `gorm:"column:bEmailVerified" json:"email_verified"`
		FirstName        *string `gorm:"column:sFirstName" json:"first_name"`
		LastName         *string `gorm:"column:sLastName" json:"last_name"`
		StudentID        *string `gorm:"column:sStudentId" json:"student_id"`
		CountryCode      string  `gorm:"column:sCountryCode" json:"country_code"`
		TimeZone         *string `gorm:"column:sTimeZone" json:"time_zone"`
		BirthDate        *string `gorm:"column:sBirthDate" json:"birth_date"`
		GraduationYear   int32   `gorm:"column:iGraduationYear" json:"graduation_year"`
		Grade            *int32  `gorm:"column:iGrade" json:"grade"`
		Sex              *string `gorm:"column:sSex" json:"sex"`
		Address          *string `gorm:"column:sAddress" json:"address"`
		ZipCode          *string `gorm:"column:sZipcode" json:"zip_code"`
		City             *string `gorm:"column:sCity" json:"city"`
		LandLineNumber   *string `gorm:"column:sLandLineNumber" json:"land_line_number"`
		CellPhoneNumber  *string `gorm:"column:sCellPhoneNumber" json:"cell_phone_number"`
		DefaultLanguage  string  `gorm:"column:sDefaultLanguage" json:"default_language"`
		PublicFirstName  bool    `gorm:"column:bPublicFirstName" json:"public_first_name"`
		PublicLastName   bool    `gorm:"column:bPublicLastName" json:"public_last_name"`

		NotifyNews        bool    `gorm:"column:bNotifyNews" json:"notify_news"`
		Notify            string  `gorm:"column:sNotify" json:"notify"`
		FreeText          *string `gorm:"column:sFreeText" json:"free_text"`
		WebSite           *string `gorm:"column:sWebSite" json:"web_site"`
		PhotoAutoload     bool    `gorm:"column:bPhotoAutoload" json:"photo_autoload"`
		LangProg          *string `gorm:"column:sLangProg" json:"lang_prog"`
		BasicEditorMode   bool    `gorm:"column:bBasicEditorMode" json:"basic_editor_mode"`
		SpacesForTab      int32   `gorm:"column:nbSpacesForTab" json:"spaces_for_tab"`
		StepLevelInSite   int32   `gorm:"column:iStepLevelInSite" json:"step_level_in_site"`
		IsAdmin           bool    `gorm:"column:bIsAdmin" json:"is_admin"`
		NoRanking         bool    `gorm:"column:bNoRanking" json:"no_ranking"`
		LoginModulePrefix *string `gorm:"column:loginModulePrefix" json:"login_module_prefix"`
		AllowSubgroups    *bool   `gorm:"column:allowSubgroups" json:"allow_subgroups"`
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

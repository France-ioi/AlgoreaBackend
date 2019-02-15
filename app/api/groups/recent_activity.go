package groups

import (
	"errors"
	"github.com/go-chi/render"
	"net/http"
	"strings"
	"unicode"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getRecentActivity(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	/*
		fromID, fromIDError := service.ResolveURLQueryGetInt64Field(r, "from.id")
		fromSubmissionDate, fromSubmissionDateError := service.ResolveURLQueryGetStringField(r, "from.submission_date")
		if (fromIDError != nil && fromSubmissionDateError == nil) || (fromIDError == nil && fromSubmissionDateError != nil) {
			return service.ErrInvalidRequest(errors.New(
				"both from.id and from.submission_date or none of them must be presented"))
		}
	*/
	itemID, err := service.ResolveURLQueryGetInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	groupID, err := service.ResolveURLQueryGetInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	var count int64
	if err := srv.Store.GroupAncestors().OwnedByUserID(user.UserID).
		Where("idGroupChild = ?", groupID).Count(&count).Error(); err != nil {
		return service.ErrUnexpected(err)
	}
	if count == 0 {
		return service.ErrForbidden(errors.New("insufficient access rights"))
	}

	query := srv.Store.UserAnswers().All().WithUsers().WithItems().
		Select(
			`users_answers.ID as ID, users_answers.sSubmissionDate, users_answers.bValidated, users_answers.iScore,
       items.ID AS Item_ID, items.sType AS Item_sType,
		   users.sLogin AS User_sLogin, users.sFirstName AS User_sFirstName, users.sLastName AS User_sLastName,
			 IF(user_strings.idLanguage IS NULL, default_strings.sTitle, user_strings.sTitle) AS Item_String_sTitle,
       COALESCE(user_strings.idLanguage, default_strings.idLanguage) AS Item_String_idLanguage`).
		Where("users_answers.idItem IN (?)",
			srv.Store.ItemAncestors().All().DescendantsOf(itemID).Select("idItemChild").SubQuery()).
		Where("users_answers.sType LIKE 'Submission'").
		Joins("LEFT JOIN items_strings default_strings FORCE INDEX (idItem) ON default_strings.idItem=items.ID AND default_strings.idLanguage=items.idDefaultLanguage").
		Joins("LEFT JOIN items_strings user_strings ON user_strings.idItem=items.ID AND user_strings.idLanguage=?", user.DefaultLanguageID())
	query = srv.Store.Items().KeepItemsVisibleBy(user, query)
	query = srv.Store.GroupAncestors().KeepUsersThatAreDescendantsOf(groupID, query)

	var result []map[string]interface{}
	if err := query.ScanIntoSliceOfMaps(&result).Error(); err != nil {
		return service.ErrUnexpected(err)
	}
	convertedResult := make([]map[string]interface{}, len(result))
	for index, _ := range result {
		convertedResult[index] = map[string]interface{}{}
		for key, value := range result[index] {
			currentMap := &convertedResult[index]

			subKeys := strings.Split(key, "_")
			for subKeyIndex, subKey := range subKeys {
				if subKeyIndex == len(subKeys)-1 {
					setConvertedValue(subKey, value, currentMap)
				} else {
					subKey = toSnakeCase(subKey)
					shouldCreateSubMap := true
					if subMap, hasSubMap := (*currentMap)[subKey]; hasSubMap {
						if subMap, ok := subMap.(*map[string]interface{}); ok {
							currentMap = subMap
							shouldCreateSubMap = false
						}
					}
					if shouldCreateSubMap {
						(*currentMap)[subKey] = &map[string]interface{}{}
						currentMap = (*currentMap)[subKey].(*map[string]interface{})
					}
				}
			}
		}
	}

	/*
		Filter by items_ansestors
		Filter: users are descendands of the group

		WHERE sType='Submission'
		AND bValidated // if validated=true
		AND sSubmissionDate <= ... // if from.submission_date
		AND ID > ... // if from.id
		ORDER BY users_answers.sSubmissionDate DESC, users_answers.ID
	*/
	render.Respond(w, r, convertedResult)
	return service.NoError
}

func setConvertedValue(valueName string, value interface{}, result *map[string]interface{}) {
	if value == nil {
		return
	}

	if valueName == "ID" {
		(*result)["id"] = value.(int64)
		return
	}

	if valueName[:2] == "id" {
		valueName = toSnakeCase(valueName[2:]) + "_id"
		(*result)[valueName] = value.(int64)
		return
	}

	switch valueName[0] {
	case 's':
		valueName = valueName[1:]
	case 'b':
		value = value == 1
		valueName = valueName[1:]
	case 'i':
		valueName = valueName[1:]
	}
	(*result)[toSnakeCase(valueName)] = value
}

// toSnakeCase convert the given string to snake case following the Golang format:
// acronyms are converted to lower-case and preceded by an underscore.
func toSnakeCase(in string) string {
	runes := []rune(in)

	var out []rune
	for i := 0; i < len(runes); i++ {
		if i > 0 && (unicode.IsUpper(runes[i]) || unicode.IsNumber(runes[i])) && ((i+1 < len(runes) && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

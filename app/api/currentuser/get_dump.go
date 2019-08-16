package currentuser

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /current-user/export-data users currentUserDataExport
// ---
// summary: Export the current user's data
// description: >
//   Returns a downloadable JSON file with all the current user's data.
//   The content returned is just the dump of raw entries of each table related to the user
//
//     * `current_user` (from `users`): all attributes except `iVersion`
//     * `sessions`, `refresh_token`: all attributes, but secrets replaced with “***”
//     * `owned_groups`: `ID` and `sName` for every descendant of user’s `idGroupOwned`;
//     * `joined_groups`: `ID` and `sName` for every ancestor of user’s `idGroupSelf`;
//     * `users_answers`: all attributes;
//     * `users_items`: all attributes except `iVersion`;
//     * `groups_attempts`: the user's or his teams' attempts, all attributes except `iVersion`;
//     * `groups_groups`: where the user’s `idGroupSelf` is the `idGroupChild`, all attributes except `iVersion` + `groups.sName`.
//
//   In case of unexpected error (e.g. a DB error), the response will be a malformed JSON like
//   ```{"current_user":{"success":false,"message":"Internal Server Error","error_text":"Some error"}```
// produces:
//   - application/json
// responses:
//   "200":
//     description: The returned data dump file
//     schema:
//       type: file
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) getDump(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=user_data.json")
	w.WriteHeader(200)

	databaseName := srv.Config.Database.Connection.DBName

	_, err := w.Write([]byte("{"))
	service.MustNotBeError(err)

	writeJSONObjectElement("current_user", w, func(writer io.Writer) {
		columns := getColumnsList(srv.Store, databaseName, "users", []string{"iVersion"})
		var userData []map[string]interface{}
		service.MustNotBeError(srv.Store.Users().ByID(user.ID).Select(columns).ScanIntoSliceOfMaps(&userData).Error())
		writeValue(w, userData[0])
	})
	writeComma(w)

	writeJSONObjectArrayElement("sessions", w, func(writer io.Writer) {
		columns := getColumnsList(srv.Store, databaseName, "sessions", []string{"sAccessToken"})
		service.MustNotBeError(srv.Store.Sessions().Where("idUser = ?", user.ID).
			Select(columns + ", '***' AS sAccessToken").ScanAndHandleMaps(streamerFunc(w)).Error())
	})
	writeComma(w)

	writeJSONObjectElement("refresh_token", w, func(writer io.Writer) {
		columns := getColumnsList(srv.Store, databaseName, "refresh_tokens", []string{"sRefreshToken"})
		var refreshTokens []map[string]interface{}
		service.MustNotBeError(srv.Store.RefreshTokens().Where("idUser = ?", user.ID).
			Select(columns + ", '***' AS sRefreshToken").ScanIntoSliceOfMaps(&refreshTokens).Error())
		if len(refreshTokens) > 0 {
			writeValue(w, refreshTokens[0])
		} else {
			writeValue(w, nil)
		}
	})
	writeComma(w)

	writeJSONObjectArrayElement("owned_groups", w, func(writer io.Writer) {
		service.MustNotBeError(srv.Store.GroupAncestors().OwnedByUser(user).
			Where("idGroupChild != idGroupAncestor").
			Joins("JOIN groups ON groups.ID = idGroupChild").
			Select("groups.ID, groups.sName").ScanAndHandleMaps(streamerFunc(w)).Error())
	})
	writeComma(w)

	writeJSONObjectArrayElement("joined_groups", w, func(writer io.Writer) {
		service.MustNotBeError(srv.Store.GroupAncestors().UserAncestors(user).
			Where("idGroupChild != idGroupAncestor").
			Joins("JOIN groups ON groups.ID = idGroupAncestor").
			Select("groups.ID, groups.sName").Order("groups.ID").ScanAndHandleMaps(streamerFunc(w)).Error())
	})
	writeComma(w)

	writeJSONObjectArrayElement("users_answers", w, func(writer io.Writer) {
		service.MustNotBeError(srv.Store.UserAnswers().Where("idUser = ?", user.ID).
			ScanAndHandleMaps(streamerFunc(w)).Error())
	})
	writeComma(w)

	writeJSONObjectArrayElement("users_items", w, func(writer io.Writer) {
		columns := getColumnsList(srv.Store, databaseName, "users_items", []string{"iVersion"})
		service.MustNotBeError(srv.Store.UserItems().Where("idUser = ?", user.ID).
			Select(columns).ScanAndHandleMaps(streamerFunc(w)).Error())
	})
	writeComma(w)

	writeJSONObjectArrayElement("groups_attempts", w, func(writer io.Writer) {
		columns := getColumnsList(srv.Store, databaseName, "groups_attempts", []string{"iVersion"})
		service.MustNotBeError(srv.Store.GroupAttempts().
			Select(columns).
			Where("idGroup = ?", user.SelfGroupID).
			UnionAll(
				srv.Store.GroupAttempts().
					Select(columns).
					Where("idGroup IN (?)",
						srv.Store.GroupGroups().WhereUserIsMember(user).
							Select("groups.ID").
							Joins("JOIN groups ON groups.ID = groups_groups.idGroupParent AND groups.sType = 'Team'").
							QueryExpr()).
					QueryExpr()).
			ScanAndHandleMaps(streamerFunc(w)).Error())
	})
	writeComma(w)

	writeJSONObjectArrayElement("groups_groups", w, func(writer io.Writer) {
		columns := getColumnsList(srv.Store, databaseName, "groups_groups", []string{"iVersion"})
		service.MustNotBeError(srv.Store.GroupGroups().
			Where("idGroupChild = ?", user.SelfGroupID).
			Joins("JOIN groups ON groups.ID = idGroupParent").
			Select(columns + ", groups.sName").
			ScanAndHandleMaps(streamerFunc(w)).Error())
	})

	_, err = w.Write([]byte("}"))
	service.MustNotBeError(err)

	return service.NoError
}

func getColumnsList(store *database.DataStore, databaseName, tableName string, excludeColumns []string) string {
	var columns []string
	service.MustNotBeError(store.Table("INFORMATION_SCHEMA.COLUMNS").
		Where("TABLE_SCHEMA = ?", databaseName).
		Where("TABLE_NAME = ?", tableName).
		Where("COLUMN_NAME NOT IN (?)", excludeColumns).
		Pluck("CONCAT('`', TABLE_NAME, '`.`', COLUMN_NAME, '`')", &columns).Error())
	return strings.Join(columns, ", ")
}

func streamerFunc(w io.Writer) func(map[string]interface{}) error {
	firstRow := true
	return func(row map[string]interface{}) error {
		if !firstRow {
			_, err := w.Write([]byte(","))
			service.MustNotBeError(err)
		}
		firstRow = false
		writeValue(w, row)
		return nil
	}
}

func writeJSONObjectElement(name string, w io.Writer, valueWriterFunc func(writer io.Writer)) {
	writeValue(w, name)
	_, err := w.Write([]byte(":"))
	service.MustNotBeError(err)
	valueWriterFunc(w)
}

func writeJSONObjectArrayElement(name string, w io.Writer, elementsWriterFunc func(writer io.Writer)) {
	writeJSONObjectElement(name, w, func(w io.Writer) {
		_, err := w.Write([]byte("["))
		service.MustNotBeError(err)
		elementsWriterFunc(w)
		_, err = w.Write([]byte("]"))
		service.MustNotBeError(err)
	})
}

func writeComma(w io.Writer) {
	_, err := w.Write([]byte(","))
	service.MustNotBeError(err)
}

func writeValue(w io.Writer, value interface{}) {
	if valueMap, ok := value.(map[string]interface{}); ok {
		for key := range valueMap {
			if int64Number, isInt64 := valueMap[key].(int64); isInt64 && len(key) >= 2 &&
				(strings.EqualFold(key[0:2], "id") || key[len(key)-2:] == "ID") {
				valueMap[key] = strconv.FormatInt(int64Number, 10)
			}
		}
	}
	data, err := json.Marshal(value)
	service.MustNotBeError(err)
	_, err = w.Write(data)
	service.MustNotBeError(err)
}

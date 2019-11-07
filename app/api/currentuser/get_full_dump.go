package currentuser

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /current-user/dump-full users currentFullUserDataExport
// ---
// summary: Export the current user's data
// description: >
//   Returns a downloadable JSON file with all the current user's data.
//   The content returned is just the dump of raw entries of each table related to the user
//
//     * `current_user` (from `users`): all attributes;
//     * `sessions`, `refresh_token`: all attributes, but secrets replaced with “***”;
//     * `owned_groups`: `id` and `name` for every descendant of user’s `owned_group_id`;
//     * `joined_groups`: `id` and `name` for every ancestor of user’s `group_id`;
//     * `users_answers`: all attributes;
//     * `users_items`: all attributes;
//     * `groups_attempts`: the user's or his teams' attempts, all attributes;
//     * `groups_groups`: where the user’s `group_id` is the `child_group_id`, all attributes + `groups.name`.
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
func (srv *Service) getFullDump(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.getDumpCommon(r, w, true)
}

func (srv *Service) getDumpCommon(r *http.Request, w http.ResponseWriter, full bool) service.APIError {
	user := srv.GetUser(r)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=user_data.json")
	w.WriteHeader(200)

	databaseName := srv.Config.Database.Connection.DBName

	_, err := w.Write([]byte("{"))
	service.MustNotBeError(err)

	writeJSONObjectElement("current_user", w, func(writer io.Writer) {
		columns := getColumnsList(srv.Store, databaseName, "users", nil)
		var userData []map[string]interface{}
		service.MustNotBeError(srv.Store.Users().Where("group_id = ?", user.GroupID).Select(columns).
			ScanIntoSliceOfMaps(&userData).Error())
		writeValue(w, userData[0])
	})

	if full {
		writeComma(w)
		writeJSONObjectArrayElement("sessions", w, func(writer io.Writer) {
			columns := getColumnsList(srv.Store, databaseName, "sessions", []string{"access_token"})
			service.MustNotBeError(srv.Store.Sessions().Where("user_id = ?", user.GroupID).
				Select(columns + ", '***' AS access_token").ScanAndHandleMaps(streamerFunc(w)).Error())
		})

		writeComma(w)
		writeJSONObjectElement("refresh_token", w, func(writer io.Writer) {
			columns := getColumnsList(srv.Store, databaseName, "refresh_tokens", []string{"refresh_token"})
			var refreshTokens []map[string]interface{}
			service.MustNotBeError(srv.Store.RefreshTokens().Where("user_id = ?", user.GroupID).
				Select(columns + ", '***' AS refresh_token").ScanIntoSliceOfMaps(&refreshTokens).Error())
			if len(refreshTokens) > 0 {
				writeValue(w, refreshTokens[0])
			} else {
				writeValue(w, nil)
			}
		})
	}

	writeComma(w)
	writeJSONObjectArrayElement("owned_groups", w, func(writer io.Writer) {
		service.MustNotBeError(srv.Store.GroupAncestors().OwnedByUser(user).
			Where("child_group_id != ancestor_group_id").
			Joins("JOIN `groups` ON `groups`.id = child_group_id").
			Order("`groups`.`id`").
			Select("`groups`.id, `groups`.name").ScanAndHandleMaps(streamerFunc(w)).Error())
	})

	writeComma(w)
	writeJSONObjectArrayElement("joined_groups", w, func(writer io.Writer) {
		service.MustNotBeError(srv.Store.GroupAncestors().UserAncestors(user).
			Where("child_group_id != ancestor_group_id").
			Joins("JOIN `groups` ON `groups`.id = ancestor_group_id").
			Select("`groups`.id, `groups`.name").Order("`groups`.id").ScanAndHandleMaps(streamerFunc(w)).Error())
	})

	if full {
		writeComma(w)
		writeJSONObjectArrayElement("users_answers", w, func(writer io.Writer) {
			service.MustNotBeError(srv.Store.UserAnswers().Where("user_id = ?", user.GroupID).
				ScanAndHandleMaps(streamerFunc(w)).Error())
		})

		writeComma(w)
		writeJSONObjectArrayElement("users_items", w, func(writer io.Writer) {
			columns := getColumnsList(srv.Store, databaseName, "users_items", nil)
			service.MustNotBeError(srv.Store.UserItems().Where("user_id = ?", user.GroupID).
				Select(columns).ScanAndHandleMaps(streamerFunc(w)).Error())
		})

		writeComma(w)
		writeJSONObjectArrayElement("groups_attempts", w, func(writer io.Writer) {
			columns := getColumnsList(srv.Store, databaseName, "groups_attempts", nil)
			service.MustNotBeError(srv.Store.GroupAttempts().
				Select(columns).
				Where("group_id = ?", user.GroupID).
				UnionAll(
					srv.Store.GroupAttempts().
						Select(columns).
						Where("group_id IN (?)",
							srv.Store.GroupGroups().WhereUserIsMember(user).
								Select("`groups`.id").
								Joins("JOIN `groups` ON `groups`.id = groups_groups.parent_group_id AND `groups`.type = 'Team'").
								QueryExpr()).
						QueryExpr()).
				ScanAndHandleMaps(streamerFunc(w)).Error())
		})
	}

	writeComma(w)
	writeJSONObjectArrayElement("groups_groups", w, func(writer io.Writer) {
		columns := getColumnsList(srv.Store, databaseName, "groups_groups", nil)
		service.MustNotBeError(srv.Store.GroupGroups().
			Where("child_group_id = ?", user.GroupID).
			Joins("JOIN `groups` ON `groups`.id = parent_group_id").
			Select(columns + ", `groups`.name").
			ScanAndHandleMaps(streamerFunc(w)).Error())
	})

	_, err = w.Write([]byte("}"))
	service.MustNotBeError(err)

	return service.NoError
}

func getColumnsList(store *database.DataStore, databaseName, tableName string, excludeColumns []string) string {
	query := store.Table("INFORMATION_SCHEMA.COLUMNS").
		Where("TABLE_SCHEMA = ?", databaseName).
		Where("TABLE_NAME = ?", tableName)
	if len(excludeColumns) > 0 {
		query = query.Where("COLUMN_NAME NOT IN (?)", excludeColumns)
	}
	var columns []string
	service.MustNotBeError(query.Pluck("CONCAT('`', TABLE_NAME, '`.`', COLUMN_NAME, '`')", &columns).Error())
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
			if int64Number, isInt64 := valueMap[key].(int64); isInt64 &&
				((len(key) > 3 && key[len(key)-3:] == "_id") || key == "id") {
				valueMap[key] = strconv.FormatInt(int64Number, 10)
			} else if stringValue, isString := valueMap[key].(string); isString && isDateColumnName(key) {
				parsedTime, _ := time.Parse("2006-01-02 15:04:05", stringValue)
				valueMap[key] = parsedTime.Format(time.RFC3339)
			}
		}
	}
	data, err := json.Marshal(value)
	service.MustNotBeError(err)
	_, err = w.Write(data)
	service.MustNotBeError(err)
}

func isDateColumnName(name string) bool {
	return strings.HasSuffix(name, "_at") || strings.HasSuffix(name, "_since") ||
		strings.HasSuffix(name, "_until")
}

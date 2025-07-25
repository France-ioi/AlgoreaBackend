package currentuser

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:operation GET /current-user/dump-full users currentUserFullDataExport
//
//	---
//	summary: Export the current user's data
//	description: >
//		Returns a downloadable JSON file with all the current user's data.
//		The content returned is just the dump of raw entries of each table related to the user
//
//		* `current_user` (from `users`): all attributes;
//
//		* `sessions`, `access_tokens`: all attributes, but secrets replaced with “***”;
//
//		* `managed_groups`: `id` and `name` for every descendant of groups managed by the user;
//
//		* `joined_groups`: `id` and `name` for every ancestor of user’s `group_id`;
//
//		* `answers`: all attributes;
//
//		* `attempts`: the user's or his teams' attempts, all attributes;
//
//		* `results`: the user's or his teams' attempt results, all attributes;
//
//		* `groups_groups`: where the user’s `group_id` is the `child_group_id`, all attributes + `groups.name`;
//
//		* `group_managers`: where the user’s `group_id` is the `manager_id`, all attributes + `groups.name`;
//
//		* `group_pending_requests`: where the user’s `group_id` is the `member_id`, all attributes + `groups.name`;
//
//		* `group_membership_changes`: where the user’s `group_id` is the `member_id`, all attributes + `groups.name`.
//
//
//		In case of unexpected error (e.g. a DB error), the response will be a malformed JSON like
//		```{"current_user":{"success":false,"message":"Internal Server Error","error_text":"Some error"}```
//	produces:
//		- application/json
//	responses:
//		"200":
//				description: The returned data dump file
//				schema:
//					type: file
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getFullDump(w http.ResponseWriter, r *http.Request) error {
	return srv.getDumpCommon(w, r, true)
}

func (srv *Service) getDumpCommon(responseWriter http.ResponseWriter, httpRequest *http.Request, full bool) error {
	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)
	responseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseWriter.Header().Set("Content-Disposition", "attachment; filename=user_data.json")
	responseWriter.WriteHeader(http.StatusOK)

	_, err := responseWriter.Write([]byte("{"))
	service.MustNotBeError(err)

	writeJSONObjectElement("current_user", responseWriter, func(_ io.Writer) {
		columns := getColumnsList(store, "users", nil)
		var userData []map[string]interface{}
		service.MustNotBeError(store.Users().ByID(user.GroupID).Select(columns).
			ScanIntoSliceOfMaps(&userData).Error())
		writeValue(responseWriter, userData[0])
	})

	if full {
		writeComma(responseWriter)
		writeJSONObjectArrayElement("sessions", responseWriter, func(_ io.Writer) {
			columns := getColumnsList(store, "sessions", []string{"refresh_token"})
			service.MustNotBeError(store.Sessions().
				Where("user_id = ?", user.GroupID).
				Select(columns + ", '***' AS refresh_token").
				Order("session_id").
				ScanAndHandleMaps(streamerFunc(responseWriter)).Error())
		})

		writeComma(responseWriter)
		writeJSONObjectArrayElement("access_tokens", responseWriter, func(_ io.Writer) {
			columns := getColumnsList(store, "access_tokens", []string{"token"})
			service.MustNotBeError(store.AccessTokens().
				Joins("JOIN sessions ON sessions.session_id = access_tokens.session_id").
				Where("sessions.user_id = ?", user.GroupID).
				Select(columns + ", '***' AS token").
				Order("session_id").
				ScanAndHandleMaps(streamerFunc(responseWriter)).Error())
		})
	}

	writeComma(responseWriter)
	writeJSONObjectArrayElement("managed_groups", responseWriter, func(_ io.Writer) {
		service.MustNotBeError(store.Groups().ManagedBy(user).
			Order("`groups`.`id`").
			Group("`groups`.`id`").
			Select("`groups`.id, `groups`.name").ScanAndHandleMaps(streamerFunc(responseWriter)).Error())
	})

	writeComma(responseWriter)
	writeJSONObjectArrayElement("joined_groups", responseWriter, func(_ io.Writer) {
		service.MustNotBeError(store.ActiveGroupGroups().
			Where("groups_groups_active.child_group_id = ?", user.GroupID).
			Joins("JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups_groups_active.parent_group_id").
			Joins("JOIN `groups` ON `groups`.id = groups_ancestors_active.ancestor_group_id").
			Group("groups.id").
			Select("`groups`.id, `groups`.name").Order("`groups`.id").ScanAndHandleMaps(streamerFunc(responseWriter)).Error())
	})

	if full {
		writeComma(responseWriter)
		writeJSONObjectArrayElement("answers", responseWriter, func(_ io.Writer) {
			service.MustNotBeError(store.Answers().Where("author_id = ?", user.GroupID).
				Order("id").
				ScanAndHandleMaps(streamerFunc(responseWriter)).Error())
		})

		writeComma(responseWriter)
		writeJSONObjectArrayElement("attempts", responseWriter, func(_ io.Writer) {
			service.MustNotBeError(buildQueryForGettingAttemptsOrResults(store.Attempts().DataStore, user, "attempts").
				Order("participant_id, id").ScanAndHandleMaps(streamerFunc(responseWriter)).Error())
		})

		writeComma(responseWriter)
		writeJSONObjectArrayElement("results", responseWriter, func(_ io.Writer) {
			service.MustNotBeError(buildQueryForGettingAttemptsOrResults(store.Results().DataStore, user, "results").
				Order("participant_id, attempt_id, item_id").ScanAndHandleMaps(streamerFunc(responseWriter)).Error())
		})
	}

	writeComma(responseWriter)
	writeJSONObjectArrayElement("groups_groups", responseWriter, func(_ io.Writer) {
		columns := getColumnsList(store, "groups_groups", nil)
		service.MustNotBeError(store.GroupGroups().
			Where("child_group_id = ?", user.GroupID).
			Joins("JOIN `groups` ON `groups`.id = parent_group_id").
			Select(columns + ", `groups`.name").
			Order("parent_group_id").
			ScanAndHandleMaps(streamerFunc(responseWriter)).Error())
	})

	writeComma(responseWriter)
	writeJSONObjectArrayElement("group_managers", responseWriter, func(_ io.Writer) {
		columns := getColumnsList(store, "group_managers", nil)
		service.MustNotBeError(store.GroupManagers().
			Where("manager_id = ?", user.GroupID).
			Joins("JOIN `groups` ON `groups`.id = group_id").
			Select(columns + ", `groups`.name").
			Order("group_id").
			ScanAndHandleMaps(streamerFunc(responseWriter)).Error())
	})

	if full {
		writeComma(responseWriter)
		writeJSONObjectArrayElement("group_membership_changes", responseWriter, func(_ io.Writer) {
			columns := getColumnsList(store, "group_membership_changes", nil)
			service.MustNotBeError(store.GroupMembershipChanges().
				Where("member_id = ?", user.GroupID).
				Joins("JOIN `groups` ON `groups`.id = group_id").
				Select(columns + ", `groups`.name").
				Order("at DESC, group_id").
				ScanAndHandleMaps(streamerFunc(responseWriter)).Error())
		})

		writeComma(responseWriter)
		writeJSONObjectArrayElement("group_pending_requests", responseWriter, func(_ io.Writer) {
			columns := getColumnsList(store, "group_pending_requests", nil)
			service.MustNotBeError(store.GroupPendingRequests().
				Where("member_id = ?", user.GroupID).
				Joins("JOIN `groups` ON `groups`.id = group_id").
				Select(columns + ", `groups`.name").
				Order("group_id").
				ScanAndHandleMaps(streamerFunc(responseWriter)).Error())
		})
	}

	_, err = responseWriter.Write([]byte("}"))
	service.MustNotBeError(err)

	return nil
}

func getColumnsList(store *database.DataStore, tableName string, excludeColumns []string) string {
	query := store.Table("INFORMATION_SCHEMA.COLUMNS").
		Where("TABLE_SCHEMA = DATABASE()").
		Where("TABLE_NAME = ?", tableName)
	if len(excludeColumns) > 0 {
		query = query.Where("COLUMN_NAME NOT IN (?)", excludeColumns)
	}
	var columns []string
	service.MustNotBeError(query.Pluck("CONCAT('`', TABLE_NAME, '`.`', COLUMN_NAME, '`')", &columns).Error())
	return strings.Join(columns, ", ")
}

func streamerFunc(writer io.Writer) func(map[string]interface{}) error {
	firstRow := true
	return func(row map[string]interface{}) error {
		if !firstRow {
			_, err := writer.Write([]byte(","))
			service.MustNotBeError(err)
		}
		firstRow = false
		writeValue(writer, row)
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

func writeValue(writer io.Writer, value interface{}) {
	if valueMap, ok := value.(map[string]interface{}); ok {
		for key := range valueMap {
			if int64Number, isInt64 := valueMap[key].(int64); isInt64 &&
				((len(key) > 3 && key[len(key)-3:] == "_id") || key == "id") {
				valueMap[key] = strconv.FormatInt(int64Number, 10)
			} else if stringValue, isString := valueMap[key].(string); isString && isDateColumnName(key) {
				parsedTime, _ := time.Parse("2006-01-02 15:04:05.999999999", stringValue)
				valueMap[key] = parsedTime.Format(time.RFC3339Nano)
			}
		}
	}
	data, err := json.Marshal(value)
	service.MustNotBeError(err)
	_, err = writer.Write(data)
	service.MustNotBeError(err)
}

func isDateColumnName(name string) bool {
	return strings.HasSuffix(name, "_at") || strings.HasSuffix(name, "_since") ||
		strings.HasSuffix(name, "_until") || name == "at"
}

func buildQueryForGettingAttemptsOrResults(store *database.DataStore, user *database.User, tableName string) *database.DB {
	columns := getColumnsList(store, tableName, nil)
	return store.
		Select(columns).
		Where("participant_id = ?", user.GroupID).
		UnionAll(
			store.
				Select(columns).
				Where("participant_id IN (?)",
					store.GroupGroups().WhereUserIsMember(user).
						Where("groups_groups.is_team_membership = 1").
						Select("groups_groups.parent_group_id AS id").
						QueryExpr()))
}

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
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getFullDump(w http.ResponseWriter, r *http.Request) service.APIError {
	return srv.getDumpCommon(r, w, true)
}

func (srv *Service) getDumpCommon(r *http.Request, w http.ResponseWriter, full bool) service.APIError {
	user := srv.GetUser(r)
	store := srv.GetStore(r)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=user_data.json")
	w.WriteHeader(200)

	_, err := w.Write([]byte("{"))
	service.MustNotBeError(err)

	writeJSONObjectElement("current_user", w, func(writer io.Writer) {
		columns := getColumnsList(store, "users", nil)
		var userData []map[string]interface{}
		service.MustNotBeError(store.Users().ByID(user.GroupID).Select(columns).
			ScanIntoSliceOfMaps(&userData).Error())
		writeValue(w, userData[0])
	})

	if full {
		writeComma(w)
		writeJSONObjectArrayElement("sessions", w, func(writer io.Writer) {
			columns := getColumnsList(store, "sessions", []string{"refresh_token"})
			service.MustNotBeError(store.Sessions().
				Where("user_id = ?", user.GroupID).
				Select(columns + ", '***' AS refresh_token").
				ScanAndHandleMaps(streamerFunc(w)).Error())
		})

		writeComma(w)
		writeJSONObjectArrayElement("access_tokens", w, func(writer io.Writer) {
			columns := getColumnsList(store, "access_tokens", []string{"token"})
			service.MustNotBeError(store.AccessTokens().
				Joins("JOIN sessions ON sessions.session_id = access_tokens.session_id").
				Where("sessions.user_id = ?", user.GroupID).
				Select(columns + ", '***' AS token").
				ScanAndHandleMaps(streamerFunc(w)).Error())
		})
	}

	writeComma(w)
	writeJSONObjectArrayElement("managed_groups", w, func(writer io.Writer) {
		service.MustNotBeError(store.Groups().ManagedBy(user).
			Order("`groups`.`id`").
			Group("`groups`.`id`").
			Select("`groups`.id, `groups`.name").ScanAndHandleMaps(streamerFunc(w)).Error())
	})

	writeComma(w)
	writeJSONObjectArrayElement("joined_groups", w, func(writer io.Writer) {
		service.MustNotBeError(store.ActiveGroupGroups().
			Where("groups_groups_active.child_group_id = ?", user.GroupID).
			Joins("JOIN groups_ancestors_active ON groups_ancestors_active.child_group_id = groups_groups_active.parent_group_id").
			Joins("JOIN `groups` ON `groups`.id = groups_ancestors_active.ancestor_group_id").
			Group("groups.id").
			Select("`groups`.id, `groups`.name").Order("`groups`.id").ScanAndHandleMaps(streamerFunc(w)).Error())
	})

	if full {
		writeComma(w)
		writeJSONObjectArrayElement("answers", w, func(writer io.Writer) {
			service.MustNotBeError(store.Answers().Where("author_id = ?", user.GroupID).
				ScanAndHandleMaps(streamerFunc(w)).Error())
		})

		writeComma(w)
		writeJSONObjectArrayElement("attempts", w, func(writer io.Writer) {
			service.MustNotBeError(buildQueryForGettingAttemptsOrResults(store.Attempts().DataStore, user, "attempts").
				Order("participant_id, id").ScanAndHandleMaps(streamerFunc(w)).Error())
		})

		writeComma(w)
		writeJSONObjectArrayElement("results", w, func(writer io.Writer) {
			service.MustNotBeError(buildQueryForGettingAttemptsOrResults(store.Results().DataStore, user, "results").
				Order("participant_id, attempt_id, item_id").ScanAndHandleMaps(streamerFunc(w)).Error())
		})
	}

	writeComma(w)
	writeJSONObjectArrayElement("groups_groups", w, func(writer io.Writer) {
		columns := getColumnsList(store, "groups_groups", nil)
		service.MustNotBeError(store.GroupGroups().
			Where("child_group_id = ?", user.GroupID).
			Joins("JOIN `groups` ON `groups`.id = parent_group_id").
			Select(columns + ", `groups`.name").
			ScanAndHandleMaps(streamerFunc(w)).Error())
	})

	writeComma(w)
	writeJSONObjectArrayElement("group_managers", w, func(writer io.Writer) {
		columns := getColumnsList(store, "group_managers", nil)
		service.MustNotBeError(store.GroupManagers().
			Where("manager_id = ?", user.GroupID).
			Joins("JOIN `groups` ON `groups`.id = group_id").
			Select(columns + ", `groups`.name").
			ScanAndHandleMaps(streamerFunc(w)).Error())
	})

	if full {
		writeComma(w)
		writeJSONObjectArrayElement("group_membership_changes", w, func(writer io.Writer) {
			columns := getColumnsList(store, "group_membership_changes", nil)
			service.MustNotBeError(store.GroupMembershipChanges().
				Where("member_id = ?", user.GroupID).
				Joins("JOIN `groups` ON `groups`.id = group_id").
				Select(columns + ", `groups`.name").
				Order("at DESC, group_id").
				ScanAndHandleMaps(streamerFunc(w)).Error())
		})

		writeComma(w)
		writeJSONObjectArrayElement("group_pending_requests", w, func(writer io.Writer) {
			columns := getColumnsList(store, "group_pending_requests", nil)
			service.MustNotBeError(store.GroupPendingRequests().
				Where("member_id = ?", user.GroupID).
				Joins("JOIN `groups` ON `groups`.id = group_id").
				Select(columns + ", `groups`.name").
				ScanAndHandleMaps(streamerFunc(w)).Error())
		})
	}

	_, err = w.Write([]byte("}"))
	service.MustNotBeError(err)

	return service.NoError
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
						Select("`groups`.id").
						Joins("JOIN `groups` ON `groups`.id = groups_groups.parent_group_id AND `groups`.type = 'Team'").
						QueryExpr()).
				QueryExpr())
}

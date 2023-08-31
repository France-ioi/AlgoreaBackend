package groups

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// swagger:operation GET /groups/{group_id}/group-progress-with-answers-zip groups groupGroupProgressWithAnswersZIP
//
//	---
//	summary: Get group progress with answers as a ZIP file
//	description: >
//		Returns the current progress of a group with answers.
//
//		Content of the ZIP file
//			* the csv returned by `groupGroupProgressCSV` at its root in `group_progress.csv`
//			* directories reflecting the items subtree with directory names in the format
//				`{items_items.child_order}-{items[i].title}-{items[i].id}`
//			* Each item directory contains a subdirectory submissions which contains a directory for each user
//				with user login as name. Each of these user-item directory contains:
//				- `data.json` file with `hints_requested`, `latest_activity_at`, `score`, `submissions`, `time_spent`,
//					`validated` (as for [groupUserProgress](https://france-ioi.github.io/algorea-devdoc/api/#operation/groupUserProgress),
//					in the JSON format
//				- for each answer of type Submission of this user-item pair, there is a file named
//					`{number}-{attempt_id}-{score}-{answers.id}.txt` with:
//					* `number` is the answer order among the user-item answers (ordered by submission time)
//					* as content, `answer` field of the answer entry
//
//	parameters:
//		- name: group_id
//			in: path
//			type: integer
//			required: true
//		- name: parent_item_ids
//			in: query
//			type: array
//			required: true
//			items:
//				type: integer
//	responses:
//		"200":
//			description: OK. Success response with users progress and answers on items
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getGroupProgressWithAnswersZIP(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)
	store := srv.GetStore(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if !user.CanWatchGroupMembers(store, groupID) {
		return service.ErrForbidden(errors.New("no rights to watch group members"))
	}

	itemParentIDs, apiError := resolveAndCheckParentIDs(store, r, user, "answer")
	if apiError != service.NoError {
		return apiError
	}

	w.Header().Set("Content-Type", "application/zip")
	itemParentIDsString := make([]string, len(itemParentIDs))
	for i, id := range itemParentIDs {
		itemParentIDsString[i] = strconv.FormatInt(id, 10)
	}
	w.Header().Set(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=groups_progress_with_answers_for_group-%d-and_child_items_of-%s.zip",
			groupID, strings.Join(itemParentIDsString, "-")),
	)

	zipWriter := zip.NewWriter(w)

	csvBuffer := new(bytes.Buffer)

	groupProgressFile, err := zipWriter.Create("group_progress.csv")
	service.MustNotBeError(err)

	// For now, group_progress.csv is empty.
	csvWriter := csv.NewWriter(csvBuffer)
	csvWriter.Flush()

	_, err = groupProgressFile.Write(csvBuffer.Bytes())
	service.MustNotBeError(err)

	defer func(zipWriter *zip.Writer) {
		err = zipWriter.Close()
		service.MustNotBeError(err)

		err = zipWriter.Flush()
		service.MustNotBeError(err)
	}(zipWriter)

	return service.NoError
}

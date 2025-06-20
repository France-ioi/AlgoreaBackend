package items

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

// swagger:model breadcrumbPath
type breadcrumbPath struct {
	// required: true
	Path []breadcrumbElement `json:"path"`
	// Whether the path is already started by the participant
	// (true when the participant has at least one result with `started_at` set for each item in the path).
	// required: true
	IsStarted bool `json:"is_started"`
}

// swagger:model breadcrumbElement
type breadcrumbElement struct {
	// required: true
	ID int64 `json:"id,string"`
	// required: true
	Title *string `json:"title"`
	// required: true
	// enum: Chapter,Task,Course,Skill
	Type string `json:"type"`
	// required: true
	LanguageTag string `json:"language_tag"`
}

// swagger:operation GET /items/{item_id}/breadcrumbs-from-roots items itemBreadcrumbsFromRootsGet
//
//	---
//	summary: List all possible breadcrumbs for an item using `item_id`
//	description: >
//		Lists all paths from a root (`root_activity_id`|`root_skill_id` of groups the participant is descendant of or manages)
//		to the given item that the participant may have used to access this item.
//
//
//		Each path consists only of the items visible to both the participant and the current user
//		(`can_view`>='content' for all the items except for the final one and `can_view`>='info' for the final one).
//		The chain of attempts in the path cannot have missing results for non-final items that require explicit entry.
//		It also cannot have not-started results within or below ended or non-submission-allowing attempts for non-final items.
//
//
//		Note that paths may contain items without results for final items or non-final items not requiring explicit entry.
//		Also, paths may contain not-started results for final items even within or below ended or non-submission-allowing attempts.
//		It is even possible that the final item has no linked attempt at all while the final item requires explicit entry.
//
//
//		The service sorts the paths in the following order:
//			* Paths with an attempt linked to the final item come first.
//			* Paths where missing or not-started results are closer to the end of the path are prioritized.
//			* Paths with fewer missing or not-started results are preferred.
//			* Paths with higher `attempt_id` values are ranked higher.
//
//		The participant is `participant_id` (if given) or the current user (otherwise).
//
//
//		The item info (`title` and `language_tag`) in the paths is in the current user's language,
//		or the item's default language (if not available).
//
//
//		Restrictions:
//
//			* if `participant_id` is given, it should be a descendant of a group the current user can manage with `can_watch_members`,
//			* at least one path should exist,
//
//		otherwise the 'forbidden' error is returned.
//	parameters:
//		- name: item_id
//			in: path
//			type: integer
//			format: int64
//			required: true
//		- name: participant_id
//			in: query
//			type: integer
//			format: int64
//	responses:
//		"200":
//			description: OK. Success response with the found item path
//			schema:
//				type: array
//				items:
//					"$ref": "#/definitions/breadcrumbPath"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getBreadcrumbsFromRootsByItemID(w http.ResponseWriter, r *http.Request) error {
	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	return srv.getBreadcrumbsFromRoots(w, r, itemID)
}

// swagger:operation GET /items/by-text-id/{text_id}/breadcrumbs-from-roots items itemBreadcrumbsFromRootsByTextIdGet
//
//	---
//	summary: List all possible breadcrumbs for an item using `text_id`
//	description: >
//
//		Same as [/items/{item_id}/breadcrumbs-from-roots](#tag/items/operation/itemBreadcrumbsFromRootsGet)
//		but using `text_id`.
//
//		* `text_id` must be URL-encoded.
//
//	parameters:
//		- name: text_id
//			in: path
//			type: string
//			required: true
//		- name: participant_id
//			in: query
//			type: integer
//			format: int64
//
//	responses:
//		"200":
//			description: OK. Success response with the found item path
//			schema:
//				type: array
//				items:
//					type: array
//					items:
//						"$ref": "#/definitions/breadcrumbElement"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"403":
//			"$ref": "#/responses/forbiddenResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getBreadcrumbsFromRootsByTextID(w http.ResponseWriter, r *http.Request) error {
	textID := chi.URLParam(r, "text_id")

	// we wouldn't be here if the url weren't valid.
	decodedTextID, _ := url.QueryUnescape(textID)

	store := srv.GetStore(r)
	itemID, err := store.Items().GetItemIDFromTextID(decodedTextID)
	if err != nil {
		return service.ErrInvalidRequest(errors.New("no item found with text_id"))
	}

	return srv.getBreadcrumbsFromRoots(w, r, itemID)
}

func (srv *Service) getBreadcrumbsFromRoots(w http.ResponseWriter, r *http.Request, itemID int64) error {
	store := srv.GetStore(r)
	user := srv.GetUser(r)

	participantID := user.GroupID
	if len(r.URL.Query()["participant_id"]) != 0 {
		var err error
		participantID, err = service.ResolveURLQueryGetInt64Field(r, "participant_id")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}

		if !user.CanWatchGroupMembers(store, participantID) {
			return service.ErrAPIInsufficientAccessRights
		}
	}

	breadcrumbs := findItemBreadcrumbs(store, participantID, user, itemID)
	if len(breadcrumbs) == 0 {
		return service.ErrAPIInsufficientAccessRights
	}
	render.Respond(w, r, breadcrumbs)
	return nil
}

func findItemBreadcrumbs(store *database.DataStore, participantID int64, user *database.User, itemID int64) []breadcrumbPath {
	itemPaths := findItemPaths(store, participantID, itemID, 0)
	if len(itemPaths) == 0 {
		return nil
	}

	itemIDsSet := golang.NewSet[int64]()
	breadcrumbPaths := make([]breadcrumbPath, 0, len(itemPaths))
	for _, itemPath := range itemPaths {
		breadcrumb := make([]breadcrumbElement, 0, len(itemPath.Path))
		for _, id := range itemPath.Path {
			idInt64, _ := strconv.ParseInt(id, 10, 64)
			itemIDsSet.Add(idInt64)
			breadcrumb = append(breadcrumb, breadcrumbElement{ID: idInt64})
		}
		breadcrumbPaths = append(breadcrumbPaths, breadcrumbPath{
			IsStarted: itemPath.IsStarted,
			Path:      breadcrumb,
		})
	}

	itemInfoMap := getItemInfoMapForVisibleItems(store, itemIDsSet.Values(), user, participantID)

	contentViewPermissionIndex := store.PermissionsGranted().ViewIndexByName("content")

	for breadcrumbPathsIndex := 0; breadcrumbPathsIndex < len(breadcrumbPaths); breadcrumbPathsIndex++ {
		bcPath := &breadcrumbPaths[breadcrumbPathsIndex]
		for pathIndex := range bcPath.Path {
			id := bcPath.Path[pathIndex].ID
			if participantID != user.GroupID &&
				(itemInfoMap[id] == nil || // if the item is not visible to the current user
					// or a non-final item's content is not visible to the current user
					pathIndex != len(bcPath.Path)-1 && itemInfoMap[id].CanViewGeneratedValue < contentViewPermissionIndex) {
				// remove the path
				breadcrumbPaths = append(breadcrumbPaths[:breadcrumbPathsIndex], breadcrumbPaths[breadcrumbPathsIndex+1:]...)
				breadcrumbPathsIndex--
				break
			}
			bcPath.Path[pathIndex].Title = itemInfoMap[id].Title
			bcPath.Path[pathIndex].LanguageTag = itemInfoMap[id].LanguageTag
			bcPath.Path[pathIndex].Type = itemInfoMap[id].Type
		}
	}

	return breadcrumbPaths
}

type itemInfo struct {
	ID                    int64
	Title                 *string
	Type                  string
	LanguageTag           string
	CanViewGeneratedValue int
}

func getItemInfoMapForVisibleItems(
	store *database.DataStore, idsList []int64, user *database.User, participantID int64,
) map[int64]*itemInfo {
	fieldsToSelect := `
			id,
			COALESCE(user_strings.title, default_strings.title) AS title,
			COALESCE(user_strings.language_tag, default_strings.language_tag) AS language_tag,
			type`
	itemInfoQuery := store.Items().Where("id IN(?)", idsList).JoinsUserAndDefaultItemStrings(user)
	if participantID != user.GroupID {
		fieldsToSelect += `,
			can_view_generated_value`
		itemInfoQuery = itemInfoQuery.JoinsPermissionsForGroupToItemsWherePermissionAtLeast(user.GroupID, "view", "info")
	}

	var itemInfos []itemInfo
	service.MustNotBeError(itemInfoQuery.Select(fieldsToSelect).Scan(&itemInfos).Error())

	itemInfoMap := make(map[int64]*itemInfo, len(itemInfos))
	for index, itemInfo := range itemInfos {
		itemInfoMap[itemInfo.ID] = &itemInfos[index]
	}

	return itemInfoMap
}

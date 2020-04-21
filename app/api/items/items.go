// Package items provides API services for items managing
package items

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

// Service is the mount point for services related to `items`
type Service struct {
	service.Base
}

const undefined = "Undefined"
const course = "Course"
const task = "Task"
const skill = "Skill"

// SetRoutes defines the routes for this package in a route group
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Store.Sessions()))
	router.Post("/items", service.AppHandler(srv.createItem).ServeHTTP)
	router.Get(`/items/{ids:(\d+/)+}breadcrumbs`, service.AppHandler(srv.getBreadcrumbs).ServeHTTP)
	router.Get("/items/{item_id}", service.AppHandler(srv.getItem).ServeHTTP)
	router.Put("/items/{item_id}", service.AppHandler(srv.updateItem).ServeHTTP)
	router.Get("/items/{item_id}/nav-tree", service.AppHandler(srv.getItemNavigationTree).ServeHTTP)
	router.Get("/items/{item_id}/attempts/{attempt_id}/task-token", service.AppHandler(srv.getTaskToken).ServeHTTP)
	router.Get("/items/{item_id}/attempts", service.AppHandler(srv.listAttempts).ServeHTTP)
	router.Post("/items/{item_id}/attempts", service.AppHandler(srv.createAttempt).ServeHTTP)
	router.Put("/items/{item_id}/strings/{language_tag}", service.AppHandler(srv.updateItemString).ServeHTTP)
	router.Post("/items/ask-hint", service.AppHandler(srv.askHint).ServeHTTP)
	router.Post("/items/save-grade", service.AppHandler(srv.saveGrade).ServeHTTP)
}

func checkHintOrScoreTokenRequiredFields(user *database.User, taskToken *token.Task, otherTokenFieldName string,
	otherTokenConvertedUserID int64,
	otherTokenLocalItemID, otherTokenItemURL, otherTokenAttemptID string) service.APIError {
	if user.GroupID != taskToken.Converted.UserID {
		return service.ErrInvalidRequest(fmt.Errorf(
			"token in task_token doesn't correspond to user session: got idUser=%d, expected %d",
			taskToken.Converted.UserID, user.GroupID))
	}
	if user.GroupID != otherTokenConvertedUserID {
		return service.ErrInvalidRequest(fmt.Errorf(
			"token in %s doesn't correspond to user session: got idUser=%d, expected %d",
			otherTokenFieldName, otherTokenConvertedUserID, user.GroupID))
	}
	if taskToken.LocalItemID != otherTokenLocalItemID {
		return service.ErrInvalidRequest(fmt.Errorf("wrong idItemLocal in %s token", otherTokenFieldName))
	}
	if taskToken.ItemURL != otherTokenItemURL {
		return service.ErrInvalidRequest(fmt.Errorf("wrong itemUrl in %s token", otherTokenFieldName))
	}
	if taskToken.AttemptID != otherTokenAttemptID {
		return service.ErrInvalidRequest(fmt.Errorf("wrong idAttempt in %s token", otherTokenFieldName))
	}
	return service.NoError
}

// Permission represents item permissions + ItemID
type Permission struct {
	ItemID                     int64
	CanViewGeneratedValue      int
	CanGrantViewGeneratedValue int
	CanWatchGeneratedValue     int
	CanEditGeneratedValue      int
}

type permissionAndType struct {
	*Permission
	Type string
}

type itemChild struct {
	// required: true
	ItemID int64 `json:"item_id,string" sql:"column:child_item_id" validate:"set"`
	// default: 0
	Order int32 `json:"order" sql:"column:child_order"`
	// enum: Undefined,Discovery,Application,Validation,Challenge
	// default: Undefined
	// required: true
	Category string `json:"category" validate:"oneof=Undefined Discovery Application Validation Challenge"`
	// default: 1
	ScoreWeight int8 `json:"score_weight"`
	// Can be set to 'as_info' if `can_grant_view` != 'none' or to 'as_content' if `can_grant_view` >= 'content'.
	// Defaults to 'as_info' (if `can_grant_view` != 'none') or 'none'.
	// enum: none,as_info,as_content
	ContentViewPropagation string `json:"content_view_propagation" validate:"oneof=none as_info as_content"`
	// Can be set to 'as_is' (if `can_grant_view` >= 'solution') or 'as_content_with_descendants'
	// (if `can_grant_view` >= 'content_with_descendants') or 'use_content_view_propagation'.
	// Defaults to 'as_is' (if `can_grant_view` >= 'solution') or 'as_content_with_descendants'
	// (if `can_grant_view` >= 'content_with_descendants') or 'use_content_view_propagation' (otherwise).
	// enum: use_content_view_propagation,as_content_with_descendants,as_is
	UpperViewLevelsPropagation string `json:"upper_view_levels_propagation" validate:"oneof=use_content_view_propagation as_content_with_descendants as_is"` // nolint:lll
	// Can be set to true if `can_grant_view` >= 'solution_with_grant'.
	// Defaults to true  if `can_grant_view` >= 'solution_with_grant', false otherwise.
	GrantViewPropagation bool `json:"grant_view_propagation"`
	// Can be set to true if `can_watch` >= 'answer_with_grant'.
	// Defaults to true  if `can_watch` >= 'answer_with_grant', false otherwise.
	WatchPropagation bool `json:"watch_propagation"`
	// Can be set to true if `can_edit` >= 'all_with_grant'.
	// Defaults to true  if `can_edit` >= 'all_with_grant', false otherwise.
	EditPropagation bool `json:"edit_propagation"`
}

type insertItemItemsSpec struct {
	ParentItemID               int64
	ChildItemID                int64
	Order                      int32
	Category                   string
	ScoreWeight                int8
	ContentViewPropagation     string
	UpperViewLevelsPropagation string
	GrantViewPropagation       bool
	WatchPropagation           bool
	EditPropagation            bool
}

// constructItemsItemsForChildren constructs items_items rows to be inserted by itemCreate/itemEdit services.
func constructItemsItemsForChildren(children []itemChild, itemID int64) []*insertItemItemsSpec {
	parentChildSpec := make([]*insertItemItemsSpec, 0, len(children))
	for _, child := range children {
		parentChildSpec = append(parentChildSpec,
			&insertItemItemsSpec{
				ParentItemID:               itemID,
				ChildItemID:                child.ItemID,
				Order:                      child.Order,
				Category:                   child.Category,
				ScoreWeight:                child.ScoreWeight,
				ContentViewPropagation:     child.ContentViewPropagation,
				UpperViewLevelsPropagation: child.UpperViewLevelsPropagation,
				GrantViewPropagation:       child.GrantViewPropagation,
				WatchPropagation:           child.WatchPropagation,
				EditPropagation:            child.EditPropagation,
			})
	}
	return parentChildSpec
}

const (
	asInfo                    = "as_info"
	asContent                 = "as_content"
	asIs                      = "as_is"
	none                      = "none"
	asContentWithDescendants  = "as_content_with_descendants"
	useContentViewPropagation = "use_content_view_propagation"
)

func defaultContentViewPropagationForNewItemItems(canGrantViewGeneratedValue int, store *database.DataStore) string {
	if canGrantViewGeneratedValue > store.PermissionsGranted().PermissionIndexByKindAndName("grant_view", "none") {
		return asInfo
	}
	return none
}

func defaultUpperViewLevelsPropagationForNewItemItems(canGrantViewGeneratedValue int, store *database.DataStore) string {
	if canGrantViewGeneratedValue >= store.PermissionsGranted().PermissionIndexByKindAndName("grant_view", "solution") {
		return asIs
	}
	if canGrantViewGeneratedValue >= store.PermissionsGranted().PermissionIndexByKindAndName("grant_view", "content_with_descendants") {
		return asContentWithDescendants
	}
	return useContentViewPropagation
}

func defaultGrantViewPropagationForNewItemItems(canGrantViewGeneratedValue int, store *database.DataStore) bool {
	return canGrantViewGeneratedValue >= store.PermissionsGranted().PermissionIndexByKindAndName("grant_view", "solution_with_grant")
}

func defaultWatchPropagationForNewItemItems(canWatchGeneratedValue int, store *database.DataStore) bool {
	return canWatchGeneratedValue >= store.PermissionsGranted().PermissionIndexByKindAndName("watch", "answer_with_grant")
}

func defaultEditPropagationForNewItemItems(canEditGeneratedValue int, store *database.DataStore) bool {
	return canEditGeneratedValue >= store.PermissionsGranted().PermissionIndexByKindAndName("edit", "all_with_grant")
}

func validateChildrenFieldsAndApplyDefaults(childrenInfoMap map[int64]permissionAndType, children []itemChild,
	formData *formdata.FormData, store *database.DataStore) service.APIError {
	for index := range children {
		prefix := fmt.Sprintf("children[%d].", index)
		if !formData.IsSet(prefix + "category") {
			children[index].Category = undefined
		}
		if !formData.IsSet(prefix + "score_weight") {
			children[index].ScoreWeight = 1
		}

		childPermissions := childrenInfoMap[children[index].ItemID]
		apiError := validateChildContentViewPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childPermissions.Permission, store)
		if apiError != service.NoError {
			return apiError
		}

		apiError = validateChildUpperViewLevelsPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childPermissions.Permission, store)
		if apiError != service.NoError {
			return apiError
		}

		apiError = validateChildGrantViewPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childPermissions.Permission, store)
		if apiError != service.NoError {
			return apiError
		}

		apiError = validateChildWatchPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childPermissions.Permission, store)
		if apiError != service.NoError {
			return apiError
		}

		apiError = validateChildEditPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childPermissions.Permission, store)
		if apiError != service.NoError {
			return apiError
		}
	}
	return service.NoError
}

func validateChildEditPropagationAndApplyDefaultValue(formData *formdata.FormData, prefix string, child *itemChild,
	childPermissions *Permission, store *database.DataStore) service.APIError {
	if formData.IsSet(prefix + "edit_propagation") {
		if child.EditPropagation &&
			childPermissions.CanEditGeneratedValue < store.PermissionsGranted().PermissionIndexByKindAndName("edit", "all_with_grant") {
			return service.ErrForbidden(errors.New("not enough permissions for setting edit_propagation"))
		}
	} else {
		child.EditPropagation =
			defaultEditPropagationForNewItemItems(childPermissions.CanEditGeneratedValue, store)
	}
	return service.NoError
}

func validateChildWatchPropagationAndApplyDefaultValue(formData *formdata.FormData, prefix string, child *itemChild,
	childPermissions *Permission, store *database.DataStore) service.APIError {
	if formData.IsSet(prefix + "watch_propagation") {
		if child.WatchPropagation &&
			childPermissions.CanWatchGeneratedValue < store.PermissionsGranted().PermissionIndexByKindAndName("watch", "answer_with_grant") {
			return service.ErrForbidden(errors.New("not enough permissions for setting watch_propagation"))
		}
	} else {
		child.WatchPropagation =
			defaultWatchPropagationForNewItemItems(childPermissions.CanWatchGeneratedValue, store)
	}
	return service.NoError
}

func validateChildGrantViewPropagationAndApplyDefaultValue(formData *formdata.FormData, prefix string, child *itemChild,
	childPermissions *Permission, store *database.DataStore) service.APIError {
	if formData.IsSet(prefix + "grant_view_propagation") {
		if child.GrantViewPropagation &&
			childPermissions.CanGrantViewGeneratedValue <
				store.PermissionsGranted().PermissionIndexByKindAndName("grant_view", "solution_with_grant") {
			return service.ErrForbidden(errors.New("not enough permissions for setting grant_view_propagation"))
		}
	} else {
		child.GrantViewPropagation =
			defaultGrantViewPropagationForNewItemItems(childPermissions.CanGrantViewGeneratedValue, store)
	}
	return service.NoError
}

// nolint:dupl
func validateChildUpperViewLevelsPropagationAndApplyDefaultValue(formData *formdata.FormData, prefix string,
	child *itemChild, childPermissions *Permission, store *database.DataStore) service.APIError {
	if formData.IsSet(prefix + "upper_view_levels_propagation") {
		var failed bool
		switch child.UpperViewLevelsPropagation {
		case asContentWithDescendants:
			failed = childPermissions.CanGrantViewGeneratedValue <
				store.PermissionsGranted().PermissionIndexByKindAndName("grant_view", "content_with_descendants")
		case asIs:
			failed = childPermissions.CanGrantViewGeneratedValue <
				store.PermissionsGranted().PermissionIndexByKindAndName("grant_view", "solution")
		}
		if failed {
			return service.ErrForbidden(errors.New("not enough permissions for setting upper_view_levels_propagation"))
		}
	} else {
		child.UpperViewLevelsPropagation =
			defaultUpperViewLevelsPropagationForNewItemItems(childPermissions.CanGrantViewGeneratedValue, store)
	}
	return service.NoError
}

// nolint:dupl
func validateChildContentViewPropagationAndApplyDefaultValue(formData *formdata.FormData, prefix string,
	child *itemChild, childPermissions *Permission, store *database.DataStore) service.APIError {
	if formData.IsSet(prefix + "content_view_propagation") {
		var failed bool
		switch child.ContentViewPropagation {
		case asInfo:
			failed =
				childPermissions.CanGrantViewGeneratedValue == store.PermissionsGranted().PermissionIndexByKindAndName("grant_view", "none")
		case asContent:
			failed = childPermissions.CanGrantViewGeneratedValue < store.PermissionsGranted().PermissionIndexByKindAndName("grant_view", "content")
		}
		if failed {
			return service.ErrForbidden(errors.New("not enough permissions for setting content_view_propagation"))
		}
	} else {
		child.ContentViewPropagation =
			defaultContentViewPropagationForNewItemItems(childPermissions.CanGrantViewGeneratedValue, store)
	}
	return service.NoError
}

// insertItemsItems is used by itemCreate/itemEdit services to insert data constructed by
// constructItemsItemsForChildren() into the DB
func insertItemItems(store *database.DataStore, spec []*insertItemItemsSpec) {
	if len(spec) == 0 {
		return
	}

	var values = make([]interface{}, 0, len(spec)*9)

	for index := range spec {
		values = append(values,
			spec[index].ParentItemID, spec[index].ChildItemID, spec[index].Order, spec[index].Category,
			spec[index].ScoreWeight, spec[index].ContentViewPropagation,
			spec[index].UpperViewLevelsPropagation, spec[index].GrantViewPropagation, spec[index].WatchPropagation,
			spec[index].EditPropagation)
	}

	valuesMarks := strings.Repeat("(?, ?, ?, ?, ?, ?, ?, ?, ?, ?), ", len(spec)-1) + "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	// nolint:gosec
	query :=
		`INSERT INTO items_items (
			parent_item_id, child_item_id, child_order, category, score_weight,
			content_view_propagation, upper_view_levels_propagation, grant_view_propagation,
			watch_propagation, edit_propagation) VALUES ` + valuesMarks
	service.MustNotBeError(store.Exec(query, values...).Error())
}

// createContestParticipantsGroup creates a new contest participants group for the given item and
// gives "can_manage:content" permission on the item to this new group.
// The method doesn't update `items.participants_group_id` or run ItemItemStore.After()
// (a caller should do both on their own).
func createContestParticipantsGroup(store *database.DataStore, itemID int64) int64 {
	var participantsGroupID int64
	service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(s *database.DataStore) error {
		participantsGroupID = s.NewID()
		return s.Groups().InsertMap(map[string]interface{}{
			"id": participantsGroupID, "type": "ContestParticipants",
			"name": fmt.Sprintf("%d-participants", itemID),
		})
	}))
	service.MustNotBeError(store.PermissionsGranted().InsertMap(map[string]interface{}{
		"group_id":        participantsGroupID,
		"item_id":         itemID,
		"source_group_id": participantsGroupID,
		"origin":          "group_membership",
		"can_view":        "content",
	}))
	return participantsGroupID
}

func (srv *Service) getParticipantIDFromRequest(httpReq *http.Request, user *database.User) (int64, service.APIError) {
	groupID := user.GroupID
	var err error
	if len(httpReq.URL.Query()["as_team_id"]) != 0 {
		groupID, err = service.ResolveURLQueryGetInt64Field(httpReq, "as_team_id")
		if err != nil {
			return 0, service.ErrInvalidRequest(err)
		}

		var found bool
		found, err = srv.Store.Groups().ByID(groupID).Where("type = 'Team'").
			Joins("JOIN groups_groups_active ON groups_groups_active.parent_group_id = groups.id").
			Where("groups_groups_active.child_group_id = ?", user.GroupID).HasRows()
		service.MustNotBeError(err)
		if !found {
			return 0, service.ErrForbidden(errors.New("can't use given as_team_id as a user's team"))
		}
	}
	return groupID, service.NoError
}

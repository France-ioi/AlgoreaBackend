// Package items provides API services for items managing
package items

import (
	"errors"
	"fmt"

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
	*service.Base
}

const undefined = "Undefined"
const course = "Course"
const task = "Task"
const skill = "Skill"

// SetRoutes defines the routes for this package in a route group
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Store.Sessions()))
	routerWithParticipant := router.With(service.ParticipantMiddleware(srv.Store))

	router.Post("/items", service.AppHandler(srv.createItem).ServeHTTP)
	routerWithParticipant.Get(`/items/{ids:(\d+/)+}breadcrumbs`, service.AppHandler(srv.getBreadcrumbs).ServeHTTP)
	router.Put("/items/{item_id}", service.AppHandler(srv.updateItem).ServeHTTP)
	router.Delete("/items/{item_id}", service.AppHandler(srv.deleteItem).ServeHTTP)

	routerWithParticipant.Get("/items/{item_id}/children", service.AppHandler(srv.getItemChildren).ServeHTTP)
	routerWithParticipant.Get("/items/{item_id}/parents", service.AppHandler(srv.getItemParents).ServeHTTP)
	routerWithParticipant.Get("/items/{item_id}", service.AppHandler(srv.getItem).ServeHTTP)
	routerWithParticipant.Get("/items/{item_id}/navigation", service.AppHandler(srv.getItemNavigation).ServeHTTP)
	routerWithParticipant.Get("/items/{item_id}/prerequisites", service.AppHandler(srv.getItemPrerequisites).ServeHTTP)
	routerWithParticipant.Post("/items/{dependent_item_id}/prerequisites/{prerequisite_item_id}",
		service.AppHandler(srv.createDependency).ServeHTTP)
	routerWithParticipant.Delete("/items/{dependent_item_id}/prerequisites/{prerequisite_item_id}",
		service.AppHandler(srv.deleteDependency).ServeHTTP)
	routerWithParticipant.Post("/items/{dependent_item_id}/prerequisites/{prerequisite_item_id}/apply",
		service.AppHandler(srv.applyDependency).ServeHTTP)
	routerWithParticipant.Get("/items/{item_id}/dependencies", service.AppHandler(srv.getItemDependencies).ServeHTTP)
	router.Get("/items/search", service.AppHandler(srv.searchForItems).ServeHTTP)

	routerWithParticipant.Post("/items/{item_id}/attempts/{attempt_id}/generate-task-token",
		service.AppHandler(srv.generateTaskToken).ServeHTTP)
	routerWithParticipant.Post("/items/{item_id}/attempts/{attempt_id}/publish", service.AppHandler(srv.publishResult).ServeHTTP)
	routerWithParticipant.Get("/items/{item_id}/attempts", service.AppHandler(srv.listAttempts).ServeHTTP)
	routerWithParticipant.Post("/items/{ids:(\\d+/)+}attempts", service.AppHandler(srv.createAttempt).ServeHTTP)
	routerWithParticipant.Get("/items/{item_id}/log", service.AppHandler(srv.getActivityLogForItem).ServeHTTP)
	routerWithParticipant.Get("/items/log", service.AppHandler(srv.getActivityLogForAllItems).ServeHTTP)
	router.Get("/items/{item_id}/official-sessions", service.AppHandler(srv.listOfficialSessions).ServeHTTP)
	router.Put("/items/{item_id}/strings/{language_tag}", service.AppHandler(srv.updateItemString).ServeHTTP)
	router.Post("/items/ask-hint", service.AppHandler(srv.askHint).ServeHTTP)
	router.Post("/items/save-grade", service.AppHandler(srv.saveGrade).ServeHTTP)
	router.Get("/items/{item_id}/entry-state",
		service.AppHandler(srv.getEntryState).ServeHTTP)
	routerWithParticipant.Post("/items/{ids:(\\d+/)+}enter", service.AppHandler(srv.enter).ServeHTTP)
	routerWithParticipant.Post("/attempts/{attempt_id}/end", service.AppHandler(srv.endAttempt).ServeHTTP)
	routerWithParticipant.Post("/items/{ids:(\\d+/)+}start-result", service.AppHandler(srv.startResult).ServeHTTP)
	routerWithParticipant.Post("/items/{ids:(\\d+/)+}start-result-path", service.AppHandler(srv.startResultPath).ServeHTTP)
	routerWithParticipant.Get("/items/{item_id}/path-from-root", service.AppHandler(srv.getPathFromRoot).ServeHTTP)
	router.Get("/items/{item_id}/breadcrumbs-from-roots", service.AppHandler(srv.getBreadcrumbsFromRoots).ServeHTTP)
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
	Category string `json:"category" validate:"oneof=Undefined Discovery Application Validation Challenge"`
	// default: 1
	ScoreWeight int8 `json:"score_weight"`
	// Can be set to 'as_info' if `can_grant_view` != 'none' or to 'as_content' if `can_grant_view` >= 'content'.
	// Defaults to 'as_info' (if `can_grant_view` != 'none') or 'none'.
	// It is always possible to set this field to the same or a lower value, `can_grant_view` doesn't matter in this case.
	// enum: none,as_info,as_content
	ContentViewPropagation string `json:"content_view_propagation" validate:"oneof=none as_info as_content"`
	// Can be set to 'as_is' (if `can_grant_view` >= 'solution') or 'as_content_with_descendants'
	// (if `can_grant_view` >= 'content_with_descendants') or 'use_content_view_propagation'.
	// Defaults to 'as_is' (if `can_grant_view` >= 'solution') or 'as_content_with_descendants'
	// (if `can_grant_view` >= 'content_with_descendants') or 'use_content_view_propagation' (otherwise).
	// It is always possible to set this field to the same or a lower value, `can_grant_view` doesn't matter in this case.
	// enum: use_content_view_propagation,as_content_with_descendants,as_is
	UpperViewLevelsPropagation string `json:"upper_view_levels_propagation" validate:"oneof=use_content_view_propagation as_content_with_descendants as_is"` // nolint:lll
	// Can be set to true if `can_grant_view` >= 'solution_with_grant'.
	// Defaults to true  if `can_grant_view` >= 'solution_with_grant', false otherwise.
	// It is always possible to set this field to the same or a lower value, `can_grant_view` doesn't matter in this case.
	GrantViewPropagation bool `json:"grant_view_propagation"`
	// Can be set to true if `can_watch` >= 'answer_with_grant'.
	// Defaults to true  if `can_watch` >= 'answer_with_grant', false otherwise.
	// It is always possible to set this field to the same or a lower value, `can_watch` doesn't matter in this case.
	WatchPropagation bool `json:"watch_propagation"`
	// Can be set to true if `can_edit` >= 'all_with_grant'.
	// Defaults to true  if `can_edit` >= 'all_with_grant', false otherwise.
	// It is always possible to set this field to the same or a lower value, `can_edit` doesn't matter in this case.
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

type propagationLevels struct {
	ItemID                          int64
	ContentViewPropagationValue     int
	UpperViewLevelsPropagationValue int
	GrantViewPropagation            bool
	WatchPropagation                bool
	EditPropagation                 bool
}

func validateChildrenFieldsAndApplyDefaults(childrenInfoMap map[int64]permissionAndType, children []itemChild,
	formData *formdata.FormData, oldPropagationLevelsMap map[int64]*propagationLevels, store *database.DataStore) service.APIError {
	for index := range children {
		prefix := fmt.Sprintf("children[%d].", index)
		if !formData.IsSet(prefix + "category") {
			children[index].Category = undefined
		}
		if !formData.IsSet(prefix + "score_weight") {
			children[index].ScoreWeight = 1
		}

		childPermissions := childrenInfoMap[children[index].ItemID]
		oldPropagationLevels := oldPropagationLevelsMap[children[index].ItemID]
		apiError := validateChildContentViewPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childPermissions.Permission, oldPropagationLevels, store)
		if apiError != service.NoError {
			return apiError
		}

		apiError = validateChildUpperViewLevelsPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childPermissions.Permission, oldPropagationLevels, store)
		if apiError != service.NoError {
			return apiError
		}

		apiError = validateChildGrantViewPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childPermissions.Permission, oldPropagationLevels, store)
		if apiError != service.NoError {
			return apiError
		}

		apiError = validateChildWatchPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childPermissions.Permission, oldPropagationLevels, store)
		if apiError != service.NoError {
			return apiError
		}

		apiError = validateChildEditPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childPermissions.Permission, oldPropagationLevels, store)
		if apiError != service.NoError {
			return apiError
		}
	}
	return service.NoError
}

func validateChildBooleanPropagationAndApplyDefaultValue(formData *formdata.FormData, fieldName, prefix string,
	propagationValue, oldPropagationValue *bool, permissionValue, requiredPermissionValue int) service.APIError {
	if formData.IsSet(prefix + fieldName) {
		// allow setting the propagation to the same or a lower value
		if oldPropagationValue != nil && (!*propagationValue || *oldPropagationValue) {
			return service.NoError
		}
		if *propagationValue && permissionValue < requiredPermissionValue {
			return service.ErrForbidden(fmt.Errorf("not enough permissions for setting %s", fieldName))
		}
	} else {
		*propagationValue = permissionValue >= requiredPermissionValue
	}
	return service.NoError
}

func validateChildEditPropagationAndApplyDefaultValue(formData *formdata.FormData, prefix string, child *itemChild,
	childPermissions *Permission, oldPropagationLevels *propagationLevels, store *database.DataStore) service.APIError {
	var oldPropagationValue *bool
	if oldPropagationLevels != nil {
		oldPropagationValue = &oldPropagationLevels.EditPropagation
	}
	return validateChildBooleanPropagationAndApplyDefaultValue(formData, "edit_propagation", prefix,
		&child.EditPropagation, oldPropagationValue,
		childPermissions.CanEditGeneratedValue, store.PermissionsGranted().PermissionIndexByKindAndName("edit", "all_with_grant"))
}

func validateChildWatchPropagationAndApplyDefaultValue(formData *formdata.FormData, prefix string, child *itemChild,
	childPermissions *Permission, oldPropagationLevels *propagationLevels, store *database.DataStore) service.APIError {
	var oldPropagationValue *bool
	if oldPropagationLevels != nil {
		oldPropagationValue = &oldPropagationLevels.WatchPropagation
	}
	return validateChildBooleanPropagationAndApplyDefaultValue(formData, "watch_propagation", prefix,
		&child.WatchPropagation, oldPropagationValue,
		childPermissions.CanWatchGeneratedValue, store.PermissionsGranted().PermissionIndexByKindAndName("watch", "answer_with_grant"))
}

func validateChildGrantViewPropagationAndApplyDefaultValue(formData *formdata.FormData, prefix string, child *itemChild,
	childPermissions *Permission, oldPropagationLevels *propagationLevels, store *database.DataStore) service.APIError {
	var oldPropagationValue *bool
	if oldPropagationLevels != nil {
		oldPropagationValue = &oldPropagationLevels.GrantViewPropagation
	}
	return validateChildBooleanPropagationAndApplyDefaultValue(formData, "grant_view_propagation", prefix,
		&child.GrantViewPropagation, oldPropagationValue,
		childPermissions.CanGrantViewGeneratedValue, store.PermissionsGranted().PermissionIndexByKindAndName("grant_view", "solution_with_grant"))
}

// nolint:dupl
func validateChildUpperViewLevelsPropagationAndApplyDefaultValue(formData *formdata.FormData, prefix string,
	child *itemChild, childPermissions *Permission, oldPropagationLevels *propagationLevels, store *database.DataStore) service.APIError {
	if formData.IsSet(prefix + "upper_view_levels_propagation") {
		if oldPropagationLevels != nil &&
			store.ItemItems().UpperViewLevelsPropagationIndexByName(child.UpperViewLevelsPropagation) <=
				oldPropagationLevels.UpperViewLevelsPropagationValue {
			return service.NoError
		}
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
	child *itemChild, childPermissions *Permission, oldPropagationLevels *propagationLevels, store *database.DataStore) service.APIError {
	if formData.IsSet(prefix + "content_view_propagation") {
		if oldPropagationLevels != nil &&
			store.ItemItems().ContentViewPropagationIndexByName(child.ContentViewPropagation) <= oldPropagationLevels.ContentViewPropagationValue {
			return service.NoError
		}
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

	values := make([]map[string]interface{}, 0, len(spec))
	for index := range spec {
		values = append(values, map[string]interface{}{
			"parent_item_id":                spec[index].ParentItemID,
			"child_item_id":                 spec[index].ChildItemID,
			"child_order":                   spec[index].Order,
			"category":                      spec[index].Category,
			"score_weight":                  spec[index].ScoreWeight,
			"content_view_propagation":      spec[index].ContentViewPropagation,
			"upper_view_levels_propagation": spec[index].UpperViewLevelsPropagation,
			"grant_view_propagation":        spec[index].GrantViewPropagation,
			"watch_propagation":             spec[index].WatchPropagation,
			"edit_propagation":              spec[index].EditPropagation,
		})
	}

	service.MustNotBeError(store.ItemItems().InsertOrUpdateMaps(values,
		[]string{"child_order", "category", "score_weight", "content_view_propagation",
			"upper_view_levels_propagation", "grant_view_propagation", "watch_propagation", "edit_propagation"}))
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

type entryState string

const (
	alreadyStarted entryState = "already_started"
	notReady       entryState = "not_ready"
	ready          entryState = "ready"
)

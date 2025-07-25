// Package items provides API services for items managing.
package items

import (
	"errors"
	"fmt"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/auth"
	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloads"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// Service is the mount point for services related to `items`.
type Service struct {
	*service.Base
}

const (
	undefined = "Undefined"
	task      = "Task"
	skill     = "Skill"
)

// SetRoutes defines the routes for this package in a route group.
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Post("/items/ask-hint", service.AppHandler(srv.askHint).ServeHTTP)
	router.Post("/items/save-grade", service.AppHandler(srv.saveGrade).ServeHTTP)

	routerWithAuth := router.With(auth.UserMiddleware(srv.Base))
	routerWithAuthAndParticipant := routerWithAuth.With(service.ParticipantMiddleware(srv.Base))

	routerWithAuth.Post("/items", service.AppHandler(srv.createItem).ServeHTTP)
	routerWithAuthAndParticipant.Get(`/items/{ids:(\d+/)+}breadcrumbs`, service.AppHandler(srv.getBreadcrumbs).ServeHTTP)
	routerWithAuth.Put("/items/{item_id}", service.AppHandler(srv.updateItem).ServeHTTP)
	routerWithAuth.Delete("/items/{item_id}", service.AppHandler(srv.deleteItem).ServeHTTP)

	routerWithAuthAndParticipant.Get("/items/{item_id}/children", service.AppHandler(srv.getItemChildren).ServeHTTP)
	routerWithAuthAndParticipant.Get("/items/{item_id}/parents", service.AppHandler(srv.getItemParents).ServeHTTP)
	routerWithAuthAndParticipant.Get("/items/{item_id}", service.AppHandler(srv.getItem).ServeHTTP)
	routerWithAuthAndParticipant.Get("/items/{item_id}/navigation", service.AppHandler(srv.getItemNavigation).ServeHTTP)
	routerWithAuthAndParticipant.Get("/items/{item_id}/prerequisites", service.AppHandler(srv.getItemPrerequisites).ServeHTTP)
	routerWithAuthAndParticipant.Post("/items/{dependent_item_id}/prerequisites/{prerequisite_item_id}",
		service.AppHandler(srv.createDependency).ServeHTTP)
	routerWithAuthAndParticipant.Delete("/items/{dependent_item_id}/prerequisites/{prerequisite_item_id}",
		service.AppHandler(srv.deleteDependency).ServeHTTP)
	routerWithAuthAndParticipant.Post("/items/{dependent_item_id}/prerequisites/{prerequisite_item_id}/apply",
		service.AppHandler(srv.applyDependency).ServeHTTP)
	routerWithAuthAndParticipant.Get("/items/{item_id}/dependencies", service.AppHandler(srv.getItemDependencies).ServeHTTP)
	routerWithAuth.Get("/items/search", service.AppHandler(srv.searchForItems).ServeHTTP)

	routerWithAuthAndParticipant.Post("/items/{item_id}/attempts/{attempt_id}/generate-task-token",
		service.AppHandler(srv.generateTaskToken).ServeHTTP)
	routerWithAuthAndParticipant.Post("/items/{item_id}/attempts/{attempt_id}/publish", service.AppHandler(srv.publishResult).ServeHTTP)
	routerWithAuthAndParticipant.Put("/items/{item_id}/attempts/{attempt_id}", service.AppHandler(srv.updateResult).ServeHTTP)
	routerWithAuthAndParticipant.Get("/items/{item_id}/attempts", service.AppHandler(srv.listAttempts).ServeHTTP)
	routerWithAuthAndParticipant.Post("/items/{ids:(\\d+/)+}attempts", service.AppHandler(srv.createAttempt).ServeHTTP)
	routerWithAuthAndParticipant.Get("/items/{ancestor_item_id}/log", service.AppHandler(srv.getActivityLogForItem).ServeHTTP)
	routerWithAuthAndParticipant.Get("/items/log", service.AppHandler(srv.getActivityLogForAllItems).ServeHTTP)
	routerWithAuth.Get("/items/{item_id}/official-sessions", service.AppHandler(srv.listOfficialSessions).ServeHTTP)
	routerWithAuth.Put("/items/{item_id}/strings/{language_tag}", service.AppHandler(srv.updateItemString).ServeHTTP)
	routerWithAuth.Get("/items/{item_id}/entry-state",
		service.AppHandler(srv.getEntryState).ServeHTTP)
	routerWithAuthAndParticipant.Post("/items/{ids:(\\d+/)+}enter", service.AppHandler(srv.enter).ServeHTTP)
	routerWithAuthAndParticipant.Post("/attempts/{attempt_id}/end", service.AppHandler(srv.endAttempt).ServeHTTP)
	routerWithAuthAndParticipant.Post("/items/{ids:(\\d+/)+}start-result", service.AppHandler(srv.startResult).ServeHTTP)
	routerWithAuthAndParticipant.Post("/items/{ids:(\\d+/)+}start-result-path", service.AppHandler(srv.startResultPath).ServeHTTP)
	routerWithAuthAndParticipant.Get("/items/{item_id}/path-from-root", service.AppHandler(srv.getPathFromRoot).ServeHTTP)
	routerWithAuth.Get("/items/{item_id}/breadcrumbs-from-roots", service.AppHandler(srv.getBreadcrumbsFromRootsByItemID).ServeHTTP)
	routerWithAuth.Get("/items/by-text-id/{text_id}/breadcrumbs-from-roots", service.AppHandler(srv.getBreadcrumbsFromRootsByTextID).ServeHTTP)

	routerWithAuth.Put("/items/{item_id}/groups/{group_id}/additional-times", service.AppHandler(srv.setAdditionalTime).ServeHTTP)
	routerWithAuth.Get("/items/{item_id}/groups/{group_id}/additional-times", service.AppHandler(srv.getGroupAdditionalTimes).ServeHTTP)
	routerWithAuth.Get("/items/{item_id}/groups/{group_id}/members/additional-times",
		service.AppHandler(srv.getMembersAdditionalTimes).ServeHTTP)
	routerWithAuth.Get("/items/time-limited/administered", service.AppHandler(srv.getAdministeredList).ServeHTTP)
}

func checkHintOrScoreTokenRequiredFields(taskToken *payloads.TaskToken, otherTokenFieldName string,
	otherTokenConvertedUserID int64,
	otherTokenLocalItemID, otherTokenItemURL, otherTokenAttemptID string,
) error {
	if taskToken.Converted.UserID != otherTokenConvertedUserID {
		return service.ErrInvalidRequest(fmt.Errorf(
			"token in %s doesn't correspond to user session: got idUser=%d, expected %d",
			otherTokenFieldName, otherTokenConvertedUserID, taskToken.Converted.UserID))
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
	return nil
}

// Permission represents item permissions + ItemID.
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
	// required: true
	Order int32 `json:"order" sql:"column:child_order" validate:"set"`
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
	UpperViewLevelsPropagation string `json:"upper_view_levels_propagation" validate:"oneof=use_content_view_propagation as_content_with_descendants as_is"` //nolint:lll
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
	// Can be set to true if `can_request_help_to` can be set (i.e. if user have “can_grant_view ≥ content” on item)
	// Defaults to true if `can_request_help_to` can be set, false otherwise.
	// It is always possible to set this field to the same or a lower value, `can_request_help_to` doesn't matter in this case.
	RequestHelpPropagation bool `json:"request_help_propagation"`
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
	RequestHelpPropagation     bool
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
				RequestHelpPropagation:     child.RequestHelpPropagation,
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

type itemsRelationData struct {
	ItemID                          int64
	Category                        string
	ScoreWeight                     int8
	ContentViewPropagationValue     int
	UpperViewLevelsPropagationValue int
	GrantViewPropagation            bool
	WatchPropagation                bool
	EditPropagation                 bool
	RequestHelpPropagation          bool
}

func validateChildrenFieldsAndApplyDefaults(childrenInfoMap map[int64]permissionAndType, children []itemChild,
	formData *formdata.FormData, oldRelationsMap map[int64]*itemsRelationData, store *database.DataStore,
) error {
	for index := range children {
		childRelation := childrenInfoMap[children[index].ItemID]
		oldChildRelation := oldRelationsMap[children[index].ItemID]
		prefix := fmt.Sprintf("children[%d].", index)

		applyCategoryDefaultValue(formData, prefix, &children[index], oldChildRelation)
		applyScoreWeightDefaultValue(formData, prefix, &children[index], oldChildRelation)

		err := validateChildContentViewPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childRelation.Permission, oldChildRelation, store)
		if err != nil {
			return err
		}

		err = validateChildUpperViewLevelsPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childRelation.Permission, oldChildRelation, store)
		if err != nil {
			return err
		}

		err = validateChildGrantViewPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childRelation.Permission, oldChildRelation, store)
		if err != nil {
			return err
		}

		err = validateChildWatchPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childRelation.Permission, oldChildRelation, store)
		if err != nil {
			return err
		}

		err = validateChildEditPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childRelation.Permission, oldChildRelation, store)
		if err != nil {
			return err
		}

		err = validateChildRequestHelpPropagationAndApplyDefaultValue(
			formData, prefix, &children[index], childRelation.Permission, oldChildRelation, store)
		if err != nil {
			return err
		}
	}
	return nil
}

func applyCategoryDefaultValue(formData *formdata.FormData, prefix string, child *itemChild, oldRelationData *itemsRelationData) {
	if !formData.IsSet(prefix + "category") {
		if oldRelationData != nil {
			child.Category = oldRelationData.Category
		} else {
			child.Category = undefined
		}
	}
}

func applyScoreWeightDefaultValue(formData *formdata.FormData, prefix string, child *itemChild, oldRelationData *itemsRelationData) {
	if !formData.IsSet(prefix + "score_weight") {
		if oldRelationData != nil {
			child.ScoreWeight = oldRelationData.ScoreWeight
		} else {
			child.ScoreWeight = 1
		}
	}
}

func validateChildBooleanPropagationAndApplyDefaultValue(formData *formdata.FormData, fieldName, prefix string,
	propagationValue, oldPropagationValue *bool, permissionValue, requiredPermissionValue int,
) error {
	switch {
	case formData.IsSet(prefix + fieldName):
		// allow setting the propagation to the same or a lower value
		if oldPropagationValue != nil && (!*propagationValue || *oldPropagationValue) {
			return nil
		}
		if *propagationValue && permissionValue < requiredPermissionValue {
			return service.ErrForbidden(fmt.Errorf("not enough permissions for setting %s", fieldName))
		}
	case oldPropagationValue != nil:
		*propagationValue = *oldPropagationValue
	default:
		*propagationValue = permissionValue >= requiredPermissionValue
	}
	return nil
}

func validateChildEditPropagationAndApplyDefaultValue(formData *formdata.FormData, prefix string, child *itemChild,
	childPermissions *Permission, oldRelationData *itemsRelationData, store *database.DataStore,
) error {
	var oldPropagationValue *bool
	if oldRelationData != nil {
		oldPropagationValue = &oldRelationData.EditPropagation
	}
	return validateChildBooleanPropagationAndApplyDefaultValue(formData, "edit_propagation", prefix,
		&child.EditPropagation, oldPropagationValue,
		childPermissions.CanEditGeneratedValue, store.PermissionsGranted().PermissionIndexByKindAndName("edit", "all_with_grant"))
}

func validateChildRequestHelpPropagationAndApplyDefaultValue(formData *formdata.FormData, prefix string, child *itemChild,
	childPermissions *Permission, oldRelationData *itemsRelationData, store *database.DataStore,
) error {
	var oldPropagationValue *bool
	if oldRelationData != nil {
		oldPropagationValue = &oldRelationData.RequestHelpPropagation
	}

	return validateChildBooleanPropagationAndApplyDefaultValue(formData, "request_help_propagation", prefix,
		&child.RequestHelpPropagation, oldPropagationValue,
		childPermissions.CanGrantViewGeneratedValue, store.PermissionsGranted().PermissionIndexByKindAndName("can_grant_view", "content"))
}

func validateChildWatchPropagationAndApplyDefaultValue(formData *formdata.FormData, prefix string, child *itemChild,
	childPermissions *Permission, oldRelationData *itemsRelationData, store *database.DataStore,
) error {
	var oldPropagationValue *bool
	if oldRelationData != nil {
		oldPropagationValue = &oldRelationData.WatchPropagation
	}
	return validateChildBooleanPropagationAndApplyDefaultValue(formData, "watch_propagation", prefix,
		&child.WatchPropagation, oldPropagationValue,
		childPermissions.CanWatchGeneratedValue, store.PermissionsGranted().PermissionIndexByKindAndName("watch", "answer_with_grant"))
}

func validateChildGrantViewPropagationAndApplyDefaultValue(formData *formdata.FormData, prefix string, child *itemChild,
	childPermissions *Permission, oldRelationData *itemsRelationData, store *database.DataStore,
) error {
	var oldPropagationValue *bool
	if oldRelationData != nil {
		oldPropagationValue = &oldRelationData.GrantViewPropagation
	}
	return validateChildBooleanPropagationAndApplyDefaultValue(formData, "grant_view_propagation", prefix,
		&child.GrantViewPropagation, oldPropagationValue,
		childPermissions.CanGrantViewGeneratedValue, store.PermissionsGranted().PermissionIndexByKindAndName("grant_view", "solution_with_grant"))
}

//nolint:dupl // It looks like a duplicate of validateChildContentViewPropagationAndApplyDefaultValue, but it is not.
func validateChildUpperViewLevelsPropagationAndApplyDefaultValue(formData *formdata.FormData, prefix string,
	child *itemChild, childPermissions *Permission, oldRelationData *itemsRelationData, store *database.DataStore,
) error {
	switch {
	case formData.IsSet(prefix + "upper_view_levels_propagation"):
		if oldRelationData != nil &&
			store.ItemItems().UpperViewLevelsPropagationIndexByName(child.UpperViewLevelsPropagation) <=
				oldRelationData.UpperViewLevelsPropagationValue {
			return nil
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
	case oldRelationData != nil:
		child.UpperViewLevelsPropagation = store.
			ItemItems().
			UpperViewLevelsPropagationNameByIndex(oldRelationData.UpperViewLevelsPropagationValue)
	default:
		child.UpperViewLevelsPropagation = defaultUpperViewLevelsPropagationForNewItemItems(childPermissions.CanGrantViewGeneratedValue, store)
	}
	return nil
}

//nolint:dupl // It looks like a duplicate of validateChildUpperViewLevelsPropagationAndApplyDefaultValue, but it is not.
func validateChildContentViewPropagationAndApplyDefaultValue(formData *formdata.FormData, prefix string,
	child *itemChild, childPermissions *Permission, oldRelationData *itemsRelationData, store *database.DataStore,
) error {
	switch {
	case formData.IsSet(prefix + "content_view_propagation"):
		if oldRelationData != nil &&
			store.ItemItems().ContentViewPropagationIndexByName(child.ContentViewPropagation) <= oldRelationData.ContentViewPropagationValue {
			return nil
		}
		var failed bool
		switch child.ContentViewPropagation {
		case asInfo:
			failed = childPermissions.CanGrantViewGeneratedValue == store.PermissionsGranted().PermissionIndexByKindAndName("grant_view", "none")
		case asContent:
			failed = childPermissions.CanGrantViewGeneratedValue < store.PermissionsGranted().PermissionIndexByKindAndName("grant_view", "content")
		}
		if failed {
			return service.ErrForbidden(errors.New("not enough permissions for setting content_view_propagation"))
		}
	case oldRelationData != nil:
		child.ContentViewPropagation = store.ItemItems().ContentViewPropagationNameByIndex(oldRelationData.ContentViewPropagationValue)
	default:
		child.ContentViewPropagation = defaultContentViewPropagationForNewItemItems(childPermissions.CanGrantViewGeneratedValue, store)
	}
	return nil
}

// insertItemsItems is used by itemCreate/itemEdit services to insert data constructed by
// constructItemsItemsForChildren() into the DB.
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
			"request_help_propagation":      spec[index].RequestHelpPropagation,
		})
	}

	service.MustNotBeError(store.ItemItems().InsertOrUpdateMaps(values,
		[]string{
			"child_order", "category", "score_weight", "content_view_propagation", "upper_view_levels_propagation",
			"grant_view_propagation", "watch_propagation", "edit_propagation", "request_help_propagation",
		}))
}

// createParticipantsGroupForItemRequiringExplicitEntry creates a new contest participants group for the given item and
// gives "can_manage:content" permission on the item to this new group.
// The method doesn't update `items.participants_group_id` or run items ancestors recalculation
// (a caller should do both on their own).
func createParticipantsGroupForItemRequiringExplicitEntry(store *database.DataStore, itemID int64) int64 {
	var participantsGroupID int64
	service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError("groups", func(store *database.DataStore) error {
		participantsGroupID = store.NewID()
		return store.Groups().InsertMap(map[string]interface{}{
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

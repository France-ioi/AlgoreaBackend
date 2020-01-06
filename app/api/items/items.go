// Package items provides API services for items managing
package items

import (
	"fmt"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/auth"
	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/service"
	"github.com/France-ioi/AlgoreaBackend/app/token"
)

// Service is the mount point for services related to `items`
type Service struct {
	service.Base
}

// SetRoutes defines the routes for this package in a route group
func (srv *Service) SetRoutes(router chi.Router) {
	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Use(auth.UserMiddleware(srv.Store.Sessions()))
	router.Post("/items", service.AppHandler(srv.addItem).ServeHTTP)
	router.Get(`/items/{ids:(\d+/)+}breadcrumbs`, service.AppHandler(srv.getBreadcrumbs).ServeHTTP)
	router.Get("/items/{item_id}", service.AppHandler(srv.getItem).ServeHTTP)
	router.Put("/items/{item_id}", service.AppHandler(srv.updateItem).ServeHTTP)
	router.Get("/items/{item_id}/as-nav-tree", service.AppHandler(srv.getNavigationData).ServeHTTP)
	router.Get("/attempts/{attempt_id}/task-token", service.AppHandler(srv.getTaskToken).ServeHTTP)
	router.Put("/attempts/{attempt_id}/active", service.AppHandler(srv.updateActiveAttempt).ServeHTTP)
	router.Get("/items/{item_id}/attempts", service.AppHandler(srv.getAttempts).ServeHTTP)
	router.Post("/items/{item_id}/attempts", service.AppHandler(srv.createAttempt).ServeHTTP)
	router.Put("/items/{item_id}/strings/{language_id}", service.AppHandler(srv.updateItemString).ServeHTTP)
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

type permission struct {
	ItemID                     int64
	CanViewGeneratedValue      int
	CanGrantViewGeneratedValue int
	CanWatchGeneratedValue     int
	CanEditGeneratedValue      int
}

type itemChild struct {
	// required: true
	ItemID int64 `json:"item_id,string" sql:"column:child_item_id" validate:"set,child_item_id"`
	// default: 0
	Order int32 `json:"order" sql:"column:child_order"`
}

type insertItemItemsSpec struct {
	ParentItemID               int64
	ChildItemID                int64
	Order                      int32
	ContentViewPropagation     string
	UpperViewLevelsPropagation string
	GrantViewPropagation       bool
	WatchPropagation           bool
	EditPropagation            bool
}

// constructItemsItemsForChildren constructs items_items rows to be inserted by itemCreate/itemEdit services.
// `items_items.content_view_propagation` is set to 'as_info'
// while values of other `items_items.*_propagation` columns depend on the user's permissions on each child item.
func constructItemsItemsForChildren(childrenPermissions []permission, children []itemChild,
	store *database.DataStore, itemID int64) []*insertItemItemsSpec {
	childrenPermissionsMap := make(map[int64]*permission, len(childrenPermissions))
	for index := range childrenPermissions {
		childrenPermissionsMap[childrenPermissions[index].ItemID] = &childrenPermissions[index]
	}

	permissionGrantedStore := store.PermissionsGranted()
	parentChildSpec := make([]*insertItemItemsSpec, 0, len(children))
	for _, child := range children {
		permissions := childrenPermissionsMap[child.ItemID]

		upperViewLevelsPropagation := "use_content_view_propagation"
		if permissions.CanViewGeneratedValue >= permissionGrantedStore.ViewIndexByName("solution") {
			upperViewLevelsPropagation = "as_is"
		} else if permissions.CanViewGeneratedValue >= permissionGrantedStore.ViewIndexByName("content_with_descendants") {
			upperViewLevelsPropagation = "as_content_with_descendants"
		}
		parentChildSpec = append(parentChildSpec,
			&insertItemItemsSpec{
				ParentItemID: itemID, ChildItemID: child.ItemID, Order: child.Order,
				ContentViewPropagation:     "as_info",
				UpperViewLevelsPropagation: upperViewLevelsPropagation,
				GrantViewPropagation: permissions.CanGrantViewGeneratedValue >=
					permissionGrantedStore.PermissionIndexByKindAndName("grant_view", "transfer"),
				WatchPropagation: permissions.CanWatchGeneratedValue >=
					permissionGrantedStore.PermissionIndexByKindAndName("watch", "transfer"),
				EditPropagation: permissions.CanEditGeneratedValue >=
					permissionGrantedStore.PermissionIndexByKindAndName("edit", "transfer"),
			})
	}
	return parentChildSpec
}

// insertItemsItems is used by itemCreate/itemEdit services to insert data constructed by
// constructItemsItemsForChildren() into the DB
func insertItemItems(store *database.DataStore, spec []*insertItemItemsSpec) {
	if len(spec) == 0 {
		return
	}

	var values = make([]interface{}, 0, len(spec)*9)

	for index := range spec {
		values = append(values, store.NewID(),
			spec[index].ParentItemID, spec[index].ChildItemID, spec[index].Order, spec[index].ContentViewPropagation,
			spec[index].UpperViewLevelsPropagation, spec[index].GrantViewPropagation, spec[index].WatchPropagation,
			spec[index].EditPropagation)
	}

	valuesMarks := strings.Repeat("(?, ?, ?, ?, ?, ?, ?, ?, ?), ", len(spec)-1) + "(?, ?, ?, ?, ?, ?, ?, ?, ?)"
	// nolint:gosec
	query :=
		`INSERT INTO items_items (
			id, parent_item_id, child_item_id, child_order,
			content_view_propagation, upper_view_levels_propagation, grant_view_propagation,
			watch_propagation, edit_propagation) VALUES ` + valuesMarks
	service.MustNotBeError(store.Exec(query, values...).Error())
}

// createContestParticipantsGroup creates a new contest participants group for the given item and
// gives "can_manage:content" permission on the item to this new group.
// The method doesn't update `items.contest_participants_group_id` or run ItemItemStore.After()
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

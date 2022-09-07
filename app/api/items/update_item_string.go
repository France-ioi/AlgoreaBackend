package items

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

// itemStringUpdateRequest is the expected input for item's strings updating
// swagger:model itemStringUpdateRequest
type itemStringUpdateRequest struct {
	// maxLength: 200
	Title string `json:"title" validate:"max=200"`
	// Nullable
	// maxLength: 100
	ImageURL *string `json:"image_url" validate:"omitempty,max=100"`
	// Nullable
	// maxLength: 200
	Subtitle *string `json:"subtitle" validate:"omitempty,max=200"`
	// Nullable
	Description *string `json:"description"`
}

// swagger:operation PUT /items/{item_id}/strings/{language_tag} items itemStringUpdate
// ---
// summary: Update an item string entry
// description: >
//
//   Updates the corresponding `items_strings` row identified by `item_id` and `language_tag` if exists or
//   creates a new one otherwise.
//
//
//   If `language_tag` = 'default', uses the item’s default language.
//
//
//   The user should have `can_view` >= 'content' and `can_edit` >= 'all' on the item, otherwise the "forbidden" response is returned.
// parameters:
// - name: item_id
//   in: path
//   type: integer
//   format: int64
//   required: true
// - name: language_tag
//   in: path
//   type: string
//   required: true
// - in: body
//   name: data
//   required: true
//   description: New item property values
//   schema:
//     "$ref": "#/definitions/itemStringUpdateRequest"
// responses:
//   "200":
//     "$ref": "#/responses/updatedResponse"
//   "400":
//     "$ref": "#/responses/badRequestResponse"
//   "401":
//     "$ref": "#/responses/unauthorizedResponse"
//   "403":
//     "$ref": "#/responses/forbiddenResponse"
//   "500":
//     "$ref": "#/responses/internalErrorResponse"
func (srv *Service) updateItemString(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error
	user := srv.GetUser(r)

	itemID, err := service.ResolveURLQueryPathInt64Field(r, "item_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	var languageTag string
	useDefaultLanguage := true
	if chi.URLParam(r, "language_tag") != "default" {
		languageTag = chi.URLParam(r, "language_tag")
		useDefaultLanguage = false
	}

	input := itemStringUpdateRequest{}
	data := formdata.NewFormData(&input)
	apiError := service.NoError
	err = srv.GetStore(r).InTransaction(func(store *database.DataStore) error {
		err = data.ParseJSONRequestData(r)
		if err != nil {
			apiError = service.ErrInvalidRequest(err)
			return err // rollback
		}

		var found bool
		found, err = store.Permissions().MatchingUserAncestors(user).WithWriteLock().
			Where("item_id = ?", itemID).
			WherePermissionIsAtLeast("view", "content").
			WherePermissionIsAtLeast("edit", "all").
			HasRows()
		service.MustNotBeError(err)
		if !found {
			apiError = service.ErrForbidden(errors.New("no access rights to edit the item"))
			return apiError.Error // rollback
		}

		if useDefaultLanguage {
			service.MustNotBeError(store.Items().ByID(itemID).WithWriteLock().PluckFirst("default_language_tag", &languageTag).Error())
		} else {
			found, err = store.Languages().ByTag(languageTag).WithWriteLock().HasRows()
			service.MustNotBeError(err)
			if !found {
				apiError = service.ErrInvalidRequest(errors.New("no such language"))
				return apiError.Error // rollback
			}
		}
		updateItemStringData(store, itemID, languageTag, data.ConstructMapForDB())
		return nil // commit
	})

	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	// response
	service.MustNotBeError(render.Render(w, r, service.UpdateSuccess(nil)))
	return service.NoError
}

func updateItemStringData(store *database.DataStore, itemID int64, languageTag string, dbMap map[string]interface{}) {
	if len(dbMap) == 0 {
		return
	}

	columnsToUpdate := make([]string, 0, len(dbMap))
	for column := range dbMap {
		columnsToUpdate = append(columnsToUpdate, column)
	}

	dbMap["item_id"] = itemID
	dbMap["language_tag"] = languageTag

	service.MustNotBeError(store.ItemStrings().InsertOrUpdateMap(dbMap, columnsToUpdate))
}

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

// UpdateItemStringRequest is the expected input for item's strings updating
type UpdateItemStringRequest struct {
	// Nullable fields are of pointer types
	Title       string  `json:"title" validate:"max=200"`     // max length = 200
	ImageURL    *string `json:"image_url" validate:"max=100"` // max length = 100
	Subtitle    *string `json:"subtitle" validate:"max=200"`  // max length = 200
	Description *string `json:"description"`
}

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

	input := UpdateItemStringRequest{}
	data := formdata.NewFormData(&input)
	apiError := service.NoError
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		err = data.ParseJSONRequestData(r)
		if err != nil {
			apiError = service.ErrInvalidRequest(err)
			return err // rollback
		}

		var found bool
		found, err = store.Permissions().MatchingUserAncestors(user).WithWriteLock().
			Where("item_id = ?", itemID).
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

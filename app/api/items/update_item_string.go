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

	var languageID int64
	useDefaultLanguage := true
	if chi.URLParam(r, "language_id") != "default" {
		languageID, err = service.ResolveURLQueryPathInt64Field(r, "language_id")
		if err != nil {
			return service.ErrInvalidRequest(err)
		}
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
		found, err = store.PermissionsGenerated().MatchingUserAncestors(user).WithWriteLock().
			Where("item_id = ?", itemID).
			WherePermissionIsAtLeast("edit", "all").
			HasRows()
		service.MustNotBeError(err)
		if !found {
			apiError = service.ErrForbidden(errors.New("no access rights to edit the item"))
			return apiError.Error // rollback
		}

		if useDefaultLanguage {
			service.MustNotBeError(store.Items().ByID(itemID).WithWriteLock().PluckFirst("default_language_id", &languageID).Error())
		} else {
			found, err = store.Languages().ByID(languageID).WithWriteLock().HasRows()
			service.MustNotBeError(err)
			if !found {
				apiError = service.ErrInvalidRequest(errors.New("no such language"))
				return apiError.Error // rollback
			}
		}
		updateItemStringData(store, itemID, languageID, data.ConstructMapForDB())
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

func updateItemStringData(store *database.DataStore, itemID, languageID int64, dbMap map[string]interface{}) {
	scope := store.ItemStrings().
		Where("language_id = ?", languageID).
		Where("item_id = ?", itemID)
	found, err := scope.HasRows()
	service.MustNotBeError(err)
	if !found {
		service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(retryStore *database.DataStore) error {
			dbMap["id"] = retryStore.NewID()
			dbMap["item_id"] = itemID
			dbMap["language_id"] = languageID
			return retryStore.ItemStrings().InsertMap(dbMap)
		}))
	} else {
		service.MustNotBeError(scope.UpdateColumn(dbMap).Error())
	}
}

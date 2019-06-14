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
	Title       string  `json:"title" sql:"column:sTitle" validate:"max=200"`        // max length = 200
	ImageURL    *string `json:"image_url" sql:"column:sImageUrl" validate:"max=100"` // max length = 100
	Subtitle    *string `json:"subtitle" sql:"column:sSubtitle" validate:"max=200"`  // max length = 200
	Description *string `json:"description" sql:"column:sDescription"`
}

func (srv *Service) updateItemString(w http.ResponseWriter, r *http.Request) service.APIError {
	var err error
	user := srv.GetUser(r)
	err = user.Load()
	if err == database.ErrUserNotFound {
		return service.InsufficientAccessRightsError
	}
	service.MustNotBeError(err)

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
		found, err = store.Items().HasManagerAccess(user, itemID)
		service.MustNotBeError(err)
		if !found {
			apiError = service.ErrForbidden(errors.New("no access rights to manage the item"))
			return apiError.Error // rollback
		}

		if useDefaultLanguage {
			service.MustNotBeError(store.Items().ByID(itemID).PluckFirst("idDefaultLanguage", &languageID).Error())
		} else {
			found, err := store.Languages().ByID(languageID).WithWriteLock().HasRows()
			service.MustNotBeError(err)
			if !found {
				apiError = service.ErrInvalidRequest(errors.New("no such language"))
				return apiError.Error // rollback
			}
		}
		dbMap := data.ConstructMapForDB()
		scope := store.ItemStrings().
			Where("idLanguage = ?", languageID).
			Where("idItem = ?", itemID)
		found, err = scope.HasRows()
		service.MustNotBeError(err)

		if !found {
			service.MustNotBeError(store.RetryOnDuplicatePrimaryKeyError(func(retryStore *database.DataStore) error {
				dbMap["ID"] = retryStore.NewID()
				dbMap["idItem"] = itemID
				dbMap["idLanguage"] = languageID
				return retryStore.ItemStrings().InsertMap(dbMap)
			}))
		} else {
			service.MustNotBeError(scope.UpdateColumn(dbMap).Error())
		}

		return err
	})

	if apiError != service.NoError {
		return apiError
	}
	service.MustNotBeError(err)

	// response
	service.MustNotBeError(render.Render(w, r, service.UpdateSuccess(nil)))
	return service.NoError
}

package items

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"gopkg.in/go-playground/validator.v9"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/service"
)

type itemWithDefaultLanguageIDAndOptionalType struct {
	Item              item   `json:"item,squash"`
	DefaultLanguageID int64  `json:"default_language_id" sql:"column:idDefaultLanguage" validate:"default_language_id"`
	Type              string `json:"type" validate:"oneof=Root Category Chapter Task Course" sql:"column:sType"`
}

// UpdateItemRequest is the expected input for item updating
type UpdateItemRequest struct {
	// Nullable fields are of pointer types
	Item     itemWithDefaultLanguageIDAndOptionalType `json:"item,squash"`
	Children []itemChild                              `json:"children" validate:"children"`
}

func (in *UpdateItemRequest) checkItemsRelationsCycles(store *database.DataStore, itemID int64) bool {
	if len(in.Children) == 0 {
		return true
	}
	ids := make([]int64, len(in.Children)+1)
	for index := range in.Children {
		if in.Children[index].ItemID == itemID {
			return false
		}
		ids[index] = in.Children[index].ItemID
	}
	var count int64
	service.MustNotBeError(store.ItemAncestors().WithWriteLock().
		Where("idItemChild = ?", itemID).
		Where("idItemAncestor IN (?)", ids).Count(&count).Error())
	return count == 0
}

func (srv *Service) updateItem(w http.ResponseWriter, r *http.Request) service.APIError {
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

	input := UpdateItemRequest{}
	formData := formdata.NewFormData(&input)

	apiError := service.NoError
	err = srv.Store.InTransaction(func(store *database.DataStore) error {
		registerChildrenValidator(formData, store, user)
		registerItemValidators(formData, store, user)
		registerDefaultLanguageIDValidator(formData, store, itemID)

		err = formData.ParseJSONRequestData(r)
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

		service.MustNotBeError(store.Items().Where("ID = ?", itemID).UpdateColumn(formData.ConstructPartialMapForDB("Item")).Error())
		if formData.IsSet("children") {
			err = store.WithNamedLock("items_items", 3*time.Second, func(lockedStore *database.DataStore) error {
				service.MustNotBeError(lockedStore.ItemItems().Delete("idItemParent = ?", itemID).Error())

				if !input.checkItemsRelationsCycles(lockedStore, itemID) {
					apiError = service.ErrForbidden(errors.New("an item cannot become an ancestor of itself"))
					return apiError.Error // rollback
				}

				service.MustNotBeError(lockedStore.RetryOnDuplicatePrimaryKeyError(func(retryStore *database.DataStore) error {
					parentChildSpec := make([]*insertItemItemsSpec, 0, len(input.Children))
					for _, child := range input.Children {
						parentChildSpec = append(parentChildSpec,
							&insertItemItemsSpec{ParentItemID: itemID, ChildItemID: child.ItemID, Order: child.Order})
					}
					insertItemItems(retryStore, parentChildSpec)
					return nil
				}))
				return lockedStore.ItemItems().After()
			})
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

func registerDefaultLanguageIDValidator(formData *formdata.FormData, store *database.DataStore, itemID int64) {
	formData.RegisterValidation("default_language_id", constructDefaultLanguageIDValidator(formData, store, itemID))
	formData.RegisterTranslation("default_language_id",
		"default language should exist and there should be item's strings in this language")
}

// constructDefaultLanguageIDValidator constructs a validator for the DefaultLanguageID field.
// The validator checks that the language exists and there is an items_strings row in this language for the item.
func constructDefaultLanguageIDValidator(formData *formdata.FormData, store *database.DataStore, itemID int64) validator.Func {
	return validator.Func(func(fl validator.FieldLevel) bool {
		if !formData.IsValid("default_language_id") {
			return true
		}
		found, err := store.Languages().ByID(fl.Field().Interface().(int64)).WithWriteLock().
			Joins("JOIN items_strings ON items_strings.idLanguage = languages.ID AND items_strings.idItem = ?", itemID).
			HasRows()
		service.MustNotBeError(err)
		return found
	})
}

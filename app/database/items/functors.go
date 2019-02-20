package items

import (
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// OnlyVisibleBy returns a subview of the visible items for the given user basing on the given view
func OnlyVisibleBy(user database.AuthUser) database.Functor {
	return func(context *database.Context) {
		groupItemsPerms := database.NewDataStore(context.DB.New()).GroupItems().
			MatchingUserAncestors(user).
			Select("idItem, MAX(bCachedFullAccess) AS fullAccess, MAX(bCachedPartialAccess) AS partialAccess, MAX(bCachedGrayedAccess) AS grayedAccess").
			Group("idItem")

		context.DB = context.DB.Joins("JOIN ? as visible ON visible.idItem = items.ID", groupItemsPerms.SubQuery()).
			Where("fullAccess > 0 OR partialAccess > 0 OR grayedAccess > 0")
	}
}

// JoinStrings joins items_strings with the given view twice
// (as default_strings for item's default language and as user_strings for the user's default language)
func JoinStrings(user database.AuthUser) database.Functor {
	return func(context *database.Context) {
		context.DB = context.DB.
			Joins(
				`LEFT JOIN items_strings default_strings FORCE INDEX (idItem)
         ON default_strings.idItem = items.ID AND default_strings.idLanguage = items.idDefaultLanguage`).
			Joins(`LEFT JOIN items_strings user_strings
         ON user_strings.idItem=items.ID AND user_strings.idLanguage = ?`, user.DefaultLanguageID())
	}
}

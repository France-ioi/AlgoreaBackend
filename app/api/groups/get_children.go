package groups

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/app/service"
)

func (srv *Service) getChildren(w http.ResponseWriter, r *http.Request) service.APIError {
	user := srv.GetUser(r)

	groupID, err := service.ResolveURLQueryPathInt64Field(r, "group_id")
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	typesList, err := service.ResolveURLQueryGetStringSliceFieldFromIncludeExcludeParameters(r, "types",
		map[string]bool{
			"Root": true, "Class": true, "Team": true, "Club": true, "Friends": true,
			"Other": true, "UserSelf": true, "UserAdmin": true, "RootSelf": true, "RootAdmin": true,
		})
	if err != nil {
		return service.ErrInvalidRequest(err)
	}

	if apiError := checkThatUserOwnsTheGroup(srv.Store, user, groupID); apiError != service.NoError {
		return apiError
	}

	query := srv.Store.Groups().
		Select(`
			groups.ID as ID, groups.sName, groups.sType, groups.iGrade,
			groups.bOpened, groups.bFreeAccess, groups.sPassword,
			(
				SELECT COUNT(*) FROM groups AS user_groups
				JOIN groups_ancestors
				ON groups_ancestors.idGroupChild = user_groups.ID AND
					groups_ancestors.idGroupAncestor != groups_ancestors.idGroupChild
				WHERE user_groups.sType = 'UserSelf' AND groups_ancestors.idGroupAncestor = groups.ID
			) AS iUserCount`).
		Joins(`
			JOIN groups_groups ON groups.ID = groups_groups.idGroupChild AND
				groups_groups.sType IN ('direct', 'requestAccepted', 'invitationAccepted') AND
				groups_groups.idGroupParent = ?`, groupID).
		Where("groups.sType IN (?)", typesList)
	query = service.NewQueryLimiter().Apply(r, query)
	query, apiError := service.ApplySortingAndPaging(r, query,
		map[string]*service.FieldSortingParams{
			"name":  {ColumnName: "groups.sName", FieldType: "string"},
			"type":  {ColumnName: "groups.sType", FieldType: "string"},
			"grade": {ColumnName: "groups.iGrade", FieldType: "int64"},
			"id":    {ColumnName: "groups.ID", FieldType: "int64"}},
		"name")
	if apiError != service.NoError {
		return apiError
	}

	var result []map[string]interface{}
	service.MustNotBeError(query.ScanIntoSliceOfMaps(&result).Error())
	convertedResult := service.ConvertSliceOfMapsFromDBToJSON(result)

	render.Respond(w, r, convertedResult)
	return service.NoError
}

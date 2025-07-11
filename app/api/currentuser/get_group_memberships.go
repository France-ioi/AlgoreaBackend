package currentuser

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

// swagger:model membershipsViewResponseRow
type membershipsViewResponseRow struct {
	// MAX(`group_membership_changes.at`)
	// required: true
	MemberSince *database.Time `json:"member_since"`
	// `group_membership_changes.action` of the latest change
	// required: true
	// enum: invitation_accepted,join_request_accepted,joined_by_badge,joined_by_code,added_directly
	Action string `json:"action"`

	// required: true
	Group struct {
		// `groups.id`
		// required: true
		ID int64 `json:"id,string"`
		// required: true
		Name string `json:"name"`
		// required: true
		Description *string `json:"description"`
		// required: true
		// enum: Class,Team,Club,Friends,Other,Session,Base
		Type string `json:"type"`
	} `json:"group" gorm:"embedded;embedded_prefix:group__"`

	// required: true
	IsMembershipLocked bool `json:"is_membership_locked"`

	// Only for teams
	// enum: frozen_membership,would_break_entry_conditions,free_to_leave
	CanLeaveTeam string `json:"can_leave_team,omitempty"`
}

// swagger:operation GET /current-user/group-memberships group-memberships membershipsView
//
//	---
//	summary: List current user's groups
//	description:
//		Returns the list of groups memberships of the current user. Groups with `type`='ContestParticipants' are not displayed.
//	parameters:
//		- name: sort
//			in: query
//			default: [-member_since,id]
//			type: array
//			items:
//				type: string
//				enum: [member_since,-member_since,id,-id]
//		- name: from.id
//			description: Start the page from the membership next to one with `groups.id`=`{from.id}`
//			in: query
//			type: integer
//			format: int64
//		- name: limit
//			description: Display the first N memberships
//			in: query
//			type: integer
//			maximum: 1000
//			default: 500
//	responses:
//		"200":
//				description: OK. Success response with an array of groups memberships
//				schema:
//					type: array
//					items:
//						"$ref": "#/definitions/membershipsViewResponseRow"
//		"400":
//			"$ref": "#/responses/badRequestResponse"
//		"401":
//			"$ref": "#/responses/unauthorizedResponse"
//		"408":
//			"$ref": "#/responses/requestTimeoutResponse"
//		"500":
//			"$ref": "#/responses/internalErrorResponse"
func (srv *Service) getGroupMemberships(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	user := srv.GetUser(httpRequest)
	store := srv.GetStore(httpRequest)

	query := store.ActiveGroupGroups().
		Select(`
			latest_change.at AS member_since,
			IFNULL(latest_change.action, 'added_directly') AS action,
			groups.id AS group__id,
			groups.name AS group__name,
			groups.description AS group__description,
			groups.type AS group__type,
			groups_groups_active.lock_membership_approved AND NOW() < groups.require_lock_membership_approval_until AS is_membership_locked,
			IF(groups.type = 'Team',
				IF(groups.frozen_membership,
					'frozen_membership',
					IF(?,
						'would_break_entry_conditions',
						'free_to_leave'
					)
				),
				NULL
			) AS can_leave_team`,
			store.Groups().GenerateQueryCheckingIfActionBreaksEntryConditionsForActiveParticipations(
				gorm.Expr("groups.id"), user.GroupID, false, false).SubQuery()).
		Joins("JOIN `groups` ON `groups`.id = groups_groups_active.parent_group_id").
		Joins(`
			LEFT JOIN LATERAL (
				SELECT at, action FROM group_membership_changes
				WHERE group_id = groups_groups_active.parent_group_id AND member_id = groups_groups_active.child_group_id
				ORDER BY at DESC
				LIMIT 1
			) AS latest_change ON 1`).
		Where("groups_groups_active.child_group_id = ?", user.GroupID).
		Where("groups.type != 'ContestParticipants'")

	query = service.NewQueryLimiter().Apply(httpRequest, query)
	query, err := service.ApplySortingAndPaging(
		httpRequest, query,
		&service.SortingAndPagingParameters{
			Fields: service.SortingAndPagingFields{
				"member_since": {ColumnName: "latest_change.at"},
				"id":           {ColumnName: "groups.id"},
			},
			DefaultRules: "-member_since,id",
			TieBreakers:  service.SortingAndPagingTieBreakers{"id": service.FieldTypeInt64},
		})
	service.MustNotBeError(err)

	var result []membershipsViewResponseRow
	service.MustNotBeError(query.Scan(&result).Error())

	render.Respond(responseWriter, httpRequest, result)
	return nil
}

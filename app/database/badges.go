package database

import (
	"github.com/jinzhu/gorm"

	"github.com/France-ioi/AlgoreaBackend/app/logging"
)

// BadgeGroupPathElement represents an element of a badge's group path
type BadgeGroupPathElement struct {
	Manager bool   `json:"manager" validate:"set"`
	Name    string `json:"name" validate:"set"`
	URL     string `json:"url" validate:"min=1"` // length >= 1
}

// BadgeInfo contains a name and group path of a badge
type BadgeInfo struct {
	Name      string                  `json:"name" validate:"set"`
	GroupPath []BadgeGroupPathElement `json:"group_path"`
}

// Badge represents a badge from the login module
type Badge struct {
	Manager   bool      `json:"manager" validate:"set"`
	URL       string    `json:"url" validate:"min=1"` // length >= 1
	BadgeInfo BadgeInfo `json:"badge_info"`
}

// StoreBadges stores badges into the DB. It also creates groups for badge group paths and makes the given user
// a manager or member of badge groups if needed.
//
// For each badge:
// 1) if the badge's group exists and the user is already a member (or a manager if badge.Manager is true) of it: does nothing;
// 2) if the badge's group exists and the user is not already member (or a manager if badge.Manager is true) of it:
//    makes him a member of the group;
// 3) if the badge's group does not exist, creates a group with badge.BadgeInfo.Name as its name and type "Other" and
//    adds it into the group identified by an url of the last element from badge.BadgeInfo.GroupPath. If this latter group does not exist,
//    creates it (with the given name, and current user managership/membership) and puts it into the previous group from
//    badge.BadgeInfo.GroupPath, etc.
func (s *GroupStore) StoreBadges(badges []Badge, userID int64, newUser bool) (err error) {
	s.mustBeInTransaction()
	recoverPanics(&err)

	var ancestorsCalculationNeeded bool
	for index := range badges {
		s.storeBadge(&ancestorsCalculationNeeded, &badges[index], userID, newUser)
	}

	if ancestorsCalculationNeeded {
		mustNotBeError(s.GroupGroups().After())
		mustNotBeError(s.Results().Propagate())
	}
	return nil
}

func (s *GroupStore) storeBadge(ancestorsCalculationNeeded *bool, badge *Badge, userID int64, newUser bool) {
	badgeGroupID, groupCreated := s.findOrCreateGroupForBadge(badge.URL, badge.BadgeInfo.Name)
	*ancestorsCalculationNeeded = *ancestorsCalculationNeeded || groupCreated

	if badge.Manager {
		s.makeUserManagerOfBadgeGroup(badgeGroupID, userID)
	} else {
		if !groupCreated && !newUser {
			alreadyMember, err := s.ActiveGroupGroups().
				Where("parent_group_id = ? AND child_group_id = ?", badgeGroupID, userID).HasRows()
			mustNotBeError(err)
			if alreadyMember {
				return
			}
		}
		*ancestorsCalculationNeeded = !s.makeUserMemberOfBadgeGroup(badgeGroupID, userID, badge.URL) && *ancestorsCalculationNeeded
	}

	if !groupCreated {
		return
	}

	s.storeBadgeGroupPath(ancestorsCalculationNeeded, badge, userID, badgeGroupID)
}

func (s *GroupStore) storeBadgeGroupPath(ancestorsCalculationNeeded *bool, badge *Badge, userID, badgeGroupID int64) {
	for ancestorBadgeIndex := len(badge.BadgeInfo.GroupPath) - 1; ancestorBadgeIndex >= 0; ancestorBadgeIndex-- {
		childBadgeGroupID := badgeGroupID
		ancestorBadge := badge.BadgeInfo.GroupPath[ancestorBadgeIndex]
		var groupCreated bool
		badgeGroupID, groupCreated = s.findOrCreateGroupForBadge(ancestorBadge.URL, ancestorBadge.Name)
		*ancestorsCalculationNeeded = *ancestorsCalculationNeeded || groupCreated
		err := s.GroupGroups().CreateRelation(badgeGroupID, childBadgeGroupID)
		if err == ErrRelationCycle {
			logging.Warnf("Cannot add badge group %d into badge group %d (%s) because it would create a cycle",
				childBadgeGroupID, badgeGroupID, ancestorBadge.URL)
		} else {
			mustNotBeError(err)
			*ancestorsCalculationNeeded = false
		}
		if !groupCreated {
			break
		}
		if ancestorBadge.Manager {
			s.makeUserManagerOfBadgeGroup(badgeGroupID, userID)
		} else {
			*ancestorsCalculationNeeded = !s.makeUserMemberOfBadgeGroup(badgeGroupID, userID, badge.URL) && *ancestorsCalculationNeeded
		}
	}
}

func (s *GroupStore) makeUserManagerOfBadgeGroup(badgeGroupID, userID int64) bool {
	mustNotBeError(s.InsertIgnoreMaps("group_managers", []map[string]interface{}{{
		"group_id":               badgeGroupID,
		"manager_id":             userID,
		"can_manage":             "memberships",
		"can_grant_group_access": true,
		"can_watch_members":      true,
	}}))
	return s.RowsAffected() > 0
}

func (s *GroupStore) makeUserMemberOfBadgeGroup(badgeGroupID, userID int64, badgeURL string) bool {
	// This approach prevents cycles in group relations, logs the membership change, checks approvals, and respects group limits
	results, _, err := s.GroupGroups().Transition(
		UserJoinsGroupByCode, badgeGroupID, []int64{userID}, map[int64]GroupApprovals{}, userID)
	mustNotBeError(err)

	if results[userID] != Success {
		logging.Warnf("Cannot add the user %d into a badge group %d (%s), reason: %s",
			userID, badgeGroupID, badgeURL, results[userID])
	}
	return results[userID] == Success
}

func (s *GroupStore) findOrCreateGroupForBadge(badgeURL, badgeName string) (int64, bool) {
	var badgeGroupID int64
	var groupCreated bool

	err := s.WithWriteLock().Where("text_id = ?", badgeURL).PluckFirst("id", &badgeGroupID).Error()
	if gorm.IsRecordNotFoundError(err) {
		mustNotBeError(s.RetryOnDuplicatePrimaryKeyError(func(retryIDStore *DataStore) error {
			badgeGroupID = retryIDStore.NewID()
			return retryIDStore.Groups().InsertMap(map[string]interface{}{
				"id":          badgeGroupID,
				"name":        badgeName,
				"text_id":     badgeURL,
				"type":        "Other",
				"created_at":  Now(),
				"is_open":     false,
				"send_emails": false,
			})
		}))
		groupCreated = true
		err = nil
	}
	mustNotBeError(err)
	return badgeGroupID, groupCreated
}

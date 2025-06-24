package database

import (
	"errors"
	"strconv"
	"strings"

	"github.com/France-ioi/AlgoreaBackend/v2/app/logging"
)

// BadgeGroupPathElement represents an element of a badge's group path.
type BadgeGroupPathElement struct {
	Manager bool   `json:"manager" validate:"set"`
	Name    string `json:"name"    validate:"set"`
	URL     string `json:"url"     validate:"min=1"` // length >= 1
}

// BadgeInfo contains a name and group path of a badge.
type BadgeInfo struct {
	Name      string                  `json:"name"       validate:"set"`
	GroupPath []BadgeGroupPathElement `json:"group_path"`
}

// Badge represents a badge from the login module.
type Badge struct {
	Manager   bool      `json:"manager"    validate:"set"`
	URL       string    `json:"url"        validate:"min=1"` // length >= 1
	BadgeInfo BadgeInfo `json:"badge_info"`
}

// StoreBadges stores badges into the DB. It also creates groups for badge group paths and makes the given user
// a manager or member of badge groups if needed.
//
// For each badge:
//  1. if the badge's group exists and the user is already a member (or a manager if badge.Manager is true) of it: does nothing;
//  2. if the badge's group exists and the user is not already member (or a manager if badge.Manager is true) of it:
//     makes him a member (or a manager) of the group;
//  3. if the badge's group does not exist, creates a group with badge.BadgeInfo.Name as its name and type "Other" and
//     adds it into the group identified by an url of the last element from badge.BadgeInfo.GroupPath. If this latter group does not exist,
//     creates it (with the given name, and current user managership if `manager`=true) and puts it into the previous group from
//     badge.BadgeInfo.GroupPath, etc.
//  4. for every existing badge group or badge.BadgeInfo.GroupPath group makes the user a manager of the group (if he is not a manager yet).
func (s *GroupStore) StoreBadges(badges []Badge, userID int64, newUser bool) (err error) {
	s.mustBeInTransaction()
	recoverPanics(&err)

	knownBadgeGroups := s.loadKnownBadgeGroups(badges)

	managedBadgeGroups := make(map[int64]struct{})
	for index := range badges {
		s.storeBadge(&badges[index], userID, newUser, managedBadgeGroups, knownBadgeGroups)
	}

	if len(managedBadgeGroups) > 0 {
		s.makeUserManagerOfBadgeGroups(managedBadgeGroups, userID)
	}
	return nil
}

func (s *GroupStore) loadKnownBadgeGroups(badges []Badge) map[string]int64 {
	badgeURLsMap := make(map[string]struct{})
	for _, badge := range badges {
		badgeURLsMap[badge.URL] = struct{}{}
		for _, pathElement := range badge.BadgeInfo.GroupPath {
			badgeURLsMap[pathElement.URL] = struct{}{}
		}
	}
	badgeURLs := make([]string, 0, len(badgeURLsMap))
	for badgeURL := range badgeURLsMap {
		badgeURLs = append(badgeURLs, badgeURL)
	}
	var dbBadgeGroups []struct {
		ID     int64
		TextID string
	}
	mustNotBeError(s.Where("text_id IN(?)", badgeURLs).WithExclusiveWriteLock().Select("id, text_id").Scan(&dbBadgeGroups).Error())
	knownBadgeGroups := make(map[string]int64, len(badgeURLs))
	for _, dbBadgeGroup := range dbBadgeGroups {
		knownBadgeGroups[dbBadgeGroup.TextID] = dbBadgeGroup.ID
	}
	return knownBadgeGroups
}

func (s *GroupStore) storeBadge(
	badge *Badge, userID int64, newUser bool,
	managedBadgeGroups map[int64]struct{}, knownBadgeGroups map[string]int64,
) {
	badgeGroupID, groupCreated := s.findOrCreateBadgeGroup(badge.URL, badge.BadgeInfo.Name, knownBadgeGroups)

	if badge.Manager {
		managedBadgeGroups[badgeGroupID] = struct{}{}
	} else {
		var alreadyMember bool
		if !groupCreated && !newUser {
			var err error
			alreadyMember, err = s.ActiveGroupGroups().
				Where("parent_group_id = ? AND child_group_id = ?", badgeGroupID, userID).
				WithExclusiveWriteLock().
				HasRows()
			mustNotBeError(err)
		}
		if !alreadyMember {
			s.makeUserMemberOfBadgeGroup(badgeGroupID, userID, badge.URL)
		}
	}

	if groupCreated {
		s.storeBadgeGroupPath(badge, badgeGroupID, managedBadgeGroups, knownBadgeGroups)
	}
}

func (s *GroupStore) storeBadgeGroupPath(
	badge *Badge, badgeGroupID int64,
	managedBadgeGroups map[int64]struct{}, knownBadgeGroups map[string]int64,
) {
	createGroupsAndRelations := true
	for ancestorBadgeIndex := len(badge.BadgeInfo.GroupPath) - 1; ancestorBadgeIndex >= 0; ancestorBadgeIndex-- {
		childBadgeGroupID := badgeGroupID
		ancestorBadge := badge.BadgeInfo.GroupPath[ancestorBadgeIndex]
		var found, groupCreated bool
		badgeGroupID, found = knownBadgeGroups[ancestorBadge.URL]
		if !found && createGroupsAndRelations {
			badgeGroupID = s.createAndCacheBadgeGroup(ancestorBadge.URL, ancestorBadge.Name, knownBadgeGroups)
			groupCreated = true
		}
		badgeGroupIDValid := found || groupCreated
		if badgeGroupIDValid {
			if createGroupsAndRelations {
				s.createBadgeGroupRelation(badgeGroupID, childBadgeGroupID, ancestorBadge.URL)
			}
			if ancestorBadge.Manager {
				managedBadgeGroups[badgeGroupID] = struct{}{}
			}
		}
		createGroupsAndRelations = createGroupsAndRelations && groupCreated
	}
}

func (s *GroupStore) createBadgeGroupRelation(badgeGroupID, childBadgeGroupID int64, badgeURL string) bool {
	err := s.GroupGroups().CreateRelation(badgeGroupID, childBadgeGroupID)
	if errors.Is(err, ErrRelationCycle) {
		logging.SharedLogger.WithContext(s.ctx()).
			Warnf("Cannot add badge group %d into badge group %d (%s) because it would create a cycle",
				childBadgeGroupID, badgeGroupID, badgeURL)
		return false
	}
	mustNotBeError(err)
	return true
}

func (s *GroupStore) makeUserManagerOfBadgeGroups(badgeGroupIDsMap map[int64]struct{}, userID int64) {
	badgeGroupIDs := make([]string, 0, len(badgeGroupIDsMap))
	for badgeGroupID := range badgeGroupIDsMap {
		badgeGroupIDs = append(badgeGroupIDs, strconv.FormatInt(badgeGroupID, 10))
	}
	badgeGroupIDsList := strings.Join(badgeGroupIDs, ", ")
	mustNotBeError(s.Exec(`
		INSERT IGNORE INTO group_managers (group_id, manager_id, can_manage, can_grant_group_access, can_watch_members)
		SELECT badge_groups.group_id, ?, "memberships", 1, 1
			FROM JSON_TABLE('[`+badgeGroupIDsList+`]', "$[*]" COLUMNS(group_id BIGINT PATH "$")) AS badge_groups`, userID).Error())
}

func (s *GroupStore) makeUserMemberOfBadgeGroup(badgeGroupID, userID int64, badgeURL string) bool {
	// This approach prevents cycles in group relations, logs the membership change, checks approvals, and respects group limits
	results, _, err := s.GroupGroups().Transition(
		UserJoinsGroupByBadge, badgeGroupID, []int64{userID}, map[int64]GroupApprovals{}, userID)
	mustNotBeError(err)

	if results[userID] != Success {
		logging.SharedLogger.WithContext(s.ctx()).
			Warnf("Cannot add the user %d into a badge group %d (%s), reason: %s",
				userID, badgeGroupID, badgeURL, results[userID])
	}

	return results[userID] == Success
}

func (s *GroupStore) findOrCreateBadgeGroup(
	badgeURL, badgeName string, knownBadgeGroups map[string]int64,
) (int64, bool) {
	var groupCreated bool

	badgeGroupID, found := knownBadgeGroups[badgeURL]
	if !found {
		badgeGroupID = s.createAndCacheBadgeGroup(badgeURL, badgeName, knownBadgeGroups)
		groupCreated = true
	}
	return badgeGroupID, groupCreated
}

func (s *GroupStore) createBadgeGroup(badgeURL, badgeName string) int64 {
	var badgeGroupID int64
	mustNotBeError(s.RetryOnDuplicatePrimaryKeyError("groups", func(retryIDStore *DataStore) error {
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
	return badgeGroupID
}

func (s *GroupStore) createAndCacheBadgeGroup(
	badgeURL, badgeName string, knownBadgeGroups map[string]int64,
) int64 {
	badgeGroupID := s.createBadgeGroup(badgeURL, badgeName)
	knownBadgeGroups[badgeURL] = badgeGroupID
	return badgeGroupID
}

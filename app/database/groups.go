package database

import "github.com/jinzhu/gorm"

// GroupJoiningByCodeInfo represents info related to ability to join a team by code.
type GroupJoiningByCodeInfo struct {
	GroupID             int64
	Type                string
	CodeExpiresAtIsNull bool
	CodeLifetimeIsNull  bool
	FrozenMembership    bool
}

// GetGroupJoiningByCodeInfoByCode returns GroupJoiningByCodeInfo for a given code
// if there is a public team with this code and the code has not expired.
func (s *DataStore) GetGroupJoiningByCodeInfoByCode(code string, withLock bool) (groupInfo GroupJoiningByCodeInfo, found bool, err error) {
	var info GroupJoiningByCodeInfo
	query := s.Groups().
		Where("type <> 'User'").
		Where("code = ?", code).
		Where("code_expires_at IS NULL OR NOW() < code_expires_at").
		Select(`
			id AS group_id, type, code_expires_at IS NULL AS code_expires_at_is_null,
			code_lifetime IS NULL AS code_lifetime_is_null, frozen_membership`)
	if withLock {
		query = query.WithExclusiveWriteLock()
	}
	err = query.Take(&info).Error()
	if gorm.IsRecordNotFoundError(err) {
		return GroupJoiningByCodeInfo{}, false, nil
	}
	if err != nil {
		return GroupJoiningByCodeInfo{}, false, err
	}
	return info, true, nil
}

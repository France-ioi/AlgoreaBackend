package database

import "github.com/jinzhu/gorm"

// TeamJoiningByCodeInfo represents info related to ability to join a team by code
type TeamJoiningByCodeInfo struct {
	TeamID              int64
	CodeExpiresAtIsNull bool
	CodeLifetimeIsNull  bool
	FrozenMembership    bool
}

// GetTeamJoiningByCodeInfoByCode returns TeamJoiningByCodeInfo for a given code
// (or null if there is no public team with this code or the code has expired)
func (s *DataStore) GetTeamJoiningByCodeInfoByCode(code string, withLock bool) (*TeamJoiningByCodeInfo, error) {
	var info TeamJoiningByCodeInfo
	query := s.Groups().
		Where("type = 'Team'").Where("is_public").
		Where("code = ?", code).
		Where("code_expires_at IS NULL OR NOW() < code_expires_at").
		Select(`
			id AS team_id, code_expires_at IS NULL AS code_expires_at_is_null,
			code_lifetime IS NULL AS code_lifetime_is_null, frozen_membership`)
	if withLock {
		query = query.WithWriteLock()
	}
	err := query.Take(&info).Error()
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &info, err
}

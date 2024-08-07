package database

import "github.com/jinzhu/gorm"

// AccessTokenMaxLength is the maximum length of an access token.
const AccessTokenMaxLength = 2000

// AccessTokenStore implements database operations on `access_tokens`.
type AccessTokenStore struct {
	*DataStore
}

// InsertNewToken inserts a new OAuth token for the given sessionID into the DB.
func (s *AccessTokenStore) InsertNewToken(sessionID int64, token string, secondsUntilExpiry int32) error {
	return s.InsertMap(map[string]interface{}{
		"session_id": sessionID,
		"token":      token,
		"expires_at": gorm.Expr("?  + INTERVAL ? SECOND", Now(), secondsUntilExpiry),
		"issued_at":  Now(),
	})
}

// MostRecentToken represents the most recent token for a session.
type MostRecentToken struct {
	Token              string
	SecondsUntilExpiry int32
	TooNewToRefresh    bool
}

// GetMostRecentValidTokenForSession returns the most recent valid token for the given sessionID.
func (s *AccessTokenStore) GetMostRecentValidTokenForSession(sessionID int64) MostRecentToken {
	var mostRecentToken MostRecentToken

	// A token is considered too new to refresh if it was issued less than 5 minutes ago.
	err := s.Select(`
			token,
			TIMESTAMPDIFF(SECOND, NOW(), expires_at) AS seconds_until_expiry,
			issued_at > (NOW() - INTERVAL 5 MINUTE) AS too_new_to_refresh
		`).
		Where("session_id = ?", sessionID).
		Order("expires_at DESC").
		Limit(1).
		Scan(&mostRecentToken).
		Error()
	mustNotBeError(err)

	return mostRecentToken
}

// DeleteExpiredTokensOfUser deletes all expired tokens of the given user.
func (s *AccessTokenStore) DeleteExpiredTokensOfUser(userID int64) {
	sessionIDofUserQuery := s.Sessions().
		Select("session_id").
		Where("user_id = ?", userID).
		SubQuery()

	err := s.
		Where("session_id IN (?)", sessionIDofUserQuery).
		Where("expires_at < NOW()").
		Delete().
		Error()
	mustNotBeError(err)
}

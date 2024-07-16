package database

import "github.com/France-ioi/AlgoreaBackend/v2/app/database/mysqldb"

// SessionStore implements database operations on `sessions`.
type SessionStore struct {
	*DataStore
}

func (s *SessionStore) GetUserAndSessionIDByValidAccessToken(token string) (user User, sessionID int64, err error) {
	result := struct {
		User      User `gorm:"embedded"`
		SessionID int64
	}{}

	err = s.
		Select(`
			users.login,
			users.login_id,
			users.is_admin,
			users.group_id,
			users.access_group_id,
			users.temp_user,
			users.notifications_read_at,
			users.default_language,
			sessions.session_id
		`).
		Joins("JOIN users ON users.group_id = sessions.user_id").
		Joins(`
			JOIN access_tokens
			  ON access_tokens.session_id = sessions.session_id
			 AND access_tokens.token = ?
			 AND access_tokens.expires_at > NOW()
		`, token).
		Take(&result).
		Error()

	return result.User, result.SessionID, err
}

// DeleteOldSessionsToKeepMaximum deletes old sessions to keep the maximum number of sessions.
func (s *SessionStore) DeleteOldSessionsToKeepMaximum(userID int64, max int) {
	// Sessions without access tokens are treated as the oldest ones.
	// So they will be deleted first when the maximum number of sessions is reached.
	sessionsWithoutAccessTokensQuery := s.
		Select(`
			sessions.session_id AS session_id,
			 FROM_UNIXTIME(0) AS issued_at
		`).
		Joins("LEFT JOIN access_tokens ON access_tokens.session_id = sessions.session_id").
		Where("sessions.user_id = ?", userID).
		Where("access_tokens.issued_at IS NULL")

	sessionsWithAccessTokensQuery := s.
		Select(`
			access_tokens.session_id AS session_id,
			MAX(access_tokens.issued_at) AS issued_at
		`).
		Joins("JOIN access_tokens ON access_tokens.session_id = sessions.session_id").
		Where("sessions.user_id = ?", userID).
		Group("access_tokens.session_id").
		SubQuery()

	sessionToDeleteQuery := sessionsWithoutAccessTokensQuery.
		UnionAll(sessionsWithAccessTokensQuery).
		Select("session_id").
		Order("issued_at DESC, session_id").
		Limit(mysqldb.MaxRowsReturned). // Offset requires a limit in MySQL.
		Offset(max).
		SubQuery()

	// The use of tmp_table is a workaround for the following MySQL error:
	// Error 1235: This version of MySQL doesn't yet support 'LIMIT & IN/ALL/ANY/SOME subquery.
	// @see https://stackoverflow.com/questions/17892762/mysql-this-version-of-mysql-doesnt-yet-support-limit-in-all-any-some-subqu
	// Otherwise, we would just have used: "session_id IN ?".
	err := s.Delete("session_id IN (SELECT session_id FROM ? tmp_table)", sessionToDeleteQuery).
		Error()
	mustNotBeError(err)
}

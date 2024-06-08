package database

import "github.com/France-ioi/AlgoreaBackend/app/database/mysqldb"

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

type sessionWithMostRecentIssuedAt struct {
	SessionID int64
	IssuedAt  Time
}

// DeleteOldSessionsToKeepMaximum deletes old sessions to keep the maximum number of sessions.
func (s *SessionStore) DeleteOldSessionsToKeepMaximum(userID int64, max int) {
	var sessions []sessionWithMostRecentIssuedAt

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

	err := sessionsWithoutAccessTokensQuery.
		UnionAll(sessionsWithAccessTokensQuery).
		Order("issued_at DESC, session_id").
		Limit(mysqldb.MaxRowsReturned). // Offset requires a limit in MySQL.
		Offset(max).
		Scan(&sessions).
		Error()
	mustNotBeError(err)

	if len(sessions) > 0 {
		sessionIDs := make([]int64, len(sessions))
		for i, session := range sessions {
			sessionIDs[i] = session.SessionID
		}

		err := s.Delete("session_id IN (?)", sessionIDs).
			Error()
		mustNotBeError(err)
	}
}

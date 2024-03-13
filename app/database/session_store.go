package database

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

// GetUserSessionsSortedByMostRecentIssuedAt returns the user's sessions sorted by most recent issued_at.
func (s *SessionStore) GetUserSessionsSortedByMostRecentIssuedAt(userID int64) []sessionWithMostRecentIssuedAt {
	var sessions []sessionWithMostRecentIssuedAt

	err := s.
		Select(`
			access_tokens.session_id AS session_id,
			MAX(access_tokens.issued_at) AS issued_at
		`).
		Joins("JOIN access_tokens ON access_tokens.session_id = sessions.session_id").
		Where("sessions.user_id = ?", userID).
		Group("access_tokens.session_id").
		Order("issued_at DESC, sessions.session_id").
		Scan(&sessions).
		Error()
	mustNotBeError(err)

	return sessions
}

// DeleteOldSessionsToKeepMaximum deletes old sessions to keep the maximum number of sessions.
func (s *SessionStore) DeleteOldSessionsToKeepMaximum(userID int64, max int) {
	sessions := s.GetUserSessionsSortedByMostRecentIssuedAt(userID)

	if len(sessions) > max {
		// Delete the oldest sessions.
		oldestSessions := sessions[max:]

		oldestSessionIDs := make([]int64, len(oldestSessions))
		for i, session := range oldestSessions {
			oldestSessionIDs[i] = session.SessionID
		}

		err := s.Delete("session_id IN (?)", oldestSessionIDs).
			Error()
		mustNotBeError(err)
	}
}

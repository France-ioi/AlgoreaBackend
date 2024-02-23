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
		Joins("JOIN access_tokens ON access_tokens.session_id = sessions.session_id").
		Where("access_tokens.token = ?", token).
		Where("access_tokens.expires_at > NOW()").
		Take(&result).
		Error()

	return result.User, result.SessionID, err
}

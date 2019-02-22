package auth

// NewMockUser creates a mock user, to be used for testing
func NewMockUser(id int64, selfGroupID int64) *User {
	return &User{UserID: id, data: &userData{ID: id, SelfGroupID: selfGroupID}}
}

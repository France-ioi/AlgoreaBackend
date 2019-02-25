package auth

// NewMockUser creates a mock user, to be used for testing
func NewMockUser(id, selfGroupID, ownedGroupID, defaultLanguageID int64) *User {
	return &User{
		UserID: id,
		data: &userData{
			ID:                id,
			SelfGroupID:       selfGroupID,
			OwnedGroupID:      ownedGroupID,
			DefaultLanguageID: defaultLanguageID,
		}}
}

package database

// PlatformStore implements database operations on `platforms`.
type PlatformStore struct {
	*DataStore
}

// GetPublicKeyByItemID returns the public key for a specific item ID.
// Returns an empty string if there is no public key but the platform exists.
// Returns an error if the platform doesn't exist.
func (s PlatformStore) GetPublicKeyByItemID(itemID int64) (*string, error) {
	var publicKey *string
	err := s.Platforms().
		Select("public_key").
		Joins("JOIN items ON items.platform_id = platforms.id").
		Where("items.id = ?", itemID).
		PluckFirst("public_key", &publicKey).
		Error()
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

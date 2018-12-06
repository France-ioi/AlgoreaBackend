package database

// UserStore implements database operations on `users`
type UserStore struct {
  *DataStore
}

// GetByID populates `dest` with the user identified by userID
func (s *UserStore) GetByID(userID int64, dest interface{}) error {
  query := s.db.
    Table("users").
    Where("ID = ?", userID)
  query.Scan(dest)

  errors := query.GetErrors()
  if len(errors) > 0 {
    return errors[0]
  }

  return nil
}

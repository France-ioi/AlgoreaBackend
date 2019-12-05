package database

func (s *ItemStore) CheckSubmissionRightsForTimeLimitedContest(itemID int64, user *User) (bool, error) {
	return s.checkSubmissionRightsForTimeLimitedContest(itemID, user)
}

func (s *ItemStore) GetActiveContestItemIDForUser(user *User) *int64 {
	return s.getActiveContestItemIDForUser(user)
}

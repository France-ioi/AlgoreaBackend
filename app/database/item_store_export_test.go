package database

func (s *ItemStore) CheckSubmissionRightsForTimeLimitedContest(itemID int64, user *User) (bool, error) {
	return s.checkSubmissionRightsForTimeLimitedContest(itemID, user)
}

type ActiveContestInfo activeContestInfo

func (s *ItemStore) GetActiveContestInfoForUser(user *User) *ActiveContestInfo {
	return (*ActiveContestInfo)(s.getActiveContestInfoForUser(user))
}

func (s *ItemStore) CloseContest(itemID int64, user *User) {
	s.closeContest(itemID, user)
}

func (s *ItemStore) CloseTeamContest(itemID int64, user *User) {
	s.closeTeamContest(itemID, user)
}

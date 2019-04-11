package database

func (s *GroupGroupStore) Transition(action GroupGroupTransitionAction, parentGroupID int64, childGroupIDs []int64) *GroupGroupTransitionResults {
	return s.transition(action, parentGroupID, childGroupIDs)
}

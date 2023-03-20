package database

// IsThreadOpenStatus checks whether a status is considered open
func IsThreadOpenStatus(status string) bool {
	return status == "waiting_for_trainer" || status == "waiting_for_participant"
}

// IsThreadClosedStatus checks whether a status is considered closed
func IsThreadClosedStatus(status string) bool {
	return !IsThreadOpenStatus(status)
}

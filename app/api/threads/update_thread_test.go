package threads

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
)

func Test_userCanChangeThreadStatus_EdgeCases(t *testing.T) {
	user := database.User{}

	assert.False(t, userCanChangeThreadStatus(&user, "not_started", "", 1, &threadInfo{}))
	assert.True(t, userCanChangeThreadStatus(
		&user, "waiting_for_trainer", "waiting_for_trainer", 1, &threadInfo{}))
	assert.True(t, userCanChangeThreadStatus(
		&user, "closed", "closed", 1, &threadInfo{}))
}

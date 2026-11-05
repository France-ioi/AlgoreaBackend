package database

import "fmt"

// StartedResultAttemptOrderSQL is the 1-based rank of one started result among the participant's
// other started results on the same item, inside the same parent attempt. Ordering is started_at,
// then attempt_id. The expression is NULL when the item does not allow multiple attempts.
//
// Column arguments are fixed SQL references (for example "items.id"), never request data.
// The subquery names its own results and attempts tables, so the outer attempt reference must
// use a different name.
//
// When the current attempt is rooted at this item, the sibling set matches list_attempts:
// the parent attempt's own result (attempts.id = parent_attempt_id — including when that
// parent is itself rooted at this item) plus child attempts rooted at this item under
// that parent. Production only sets root_item_id on created attempts (CreateNew / enter);
// the default attempt has root_item_id NULL.
// Otherwise siblings are attempts with the same parent_attempt_id, so child items under a
// multi-attempt chapter still rank among those chapter attempts, plus attempts rooted
// at this item under the current attempt (the default attempt viewing its created children).
func StartedResultAttemptOrderSQL(
	itemAllowsMultipleAttempts, itemID, attemptParticipantID, attemptID, attemptRootItemID, attemptParentAttemptID string,
) string {
	return fmt.Sprintf(`IF(%s, (
		SELECT number FROM (
			SELECT results.attempt_id, ROW_NUMBER() OVER (ORDER BY started_at, results.attempt_id) AS number
			FROM results
			JOIN attempts ON attempts.participant_id = results.participant_id AND attempts.id = results.attempt_id
			WHERE results.participant_id = %s AND
				results.item_id = %s AND
				results.started_at IS NOT NULL AND
				IF(%s = %s,
					attempts.id = %s OR
						(attempts.parent_attempt_id <=> %s AND
						 attempts.root_item_id = %s),
					attempts.parent_attempt_id <=> %s OR
						(attempts.parent_attempt_id = %s AND attempts.root_item_id = %s)
				)
		) AS numbers WHERE numbers.attempt_id = %s
	), NULL)`,
		itemAllowsMultipleAttempts,
		attemptParticipantID,
		itemID,
		attemptRootItemID, itemID,
		attemptParentAttemptID,
		attemptParentAttemptID,
		itemID,
		attemptParentAttemptID,
		attemptID, itemID,
		attemptID,
	)
}

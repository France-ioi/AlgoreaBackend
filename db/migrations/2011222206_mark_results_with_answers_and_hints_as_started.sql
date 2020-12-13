-- +migrate Up

/* Set started_at for results having at least one of latest_submission_at or latest_hint_at set */
UPDATE `results`
SET`started_at` = COALESCE(
    (SELECT MIN(`created_at`) FROM `answers`
    WHERE `answers`.`participant_id` = `results`.`participant_id` AND
          `answers`.`attempt_id` = `results`.`attempt_id` AND
          `answers`.`item_id` = `results`.`item_id`),
    `latest_submission_at`, `latest_hint_at`),
    started_at = LEAST(`started_at`,
        IFNULL(`latest_submission_at`, `latest_hint_at`),
        IFNULL(`latest_hint_at`, `latest_submission_at`))
WHERE (`latest_submission_at` IS NOT NULL OR `latest_hint_at` IS NOT NULL) AND `started_at` IS NULL;

/*
  Set started_at for results having descendant results with started_at set for the same attempt.
  We only start a result for an ancestor item if both ancestor and descendant items are visible for the participant.
*/
INSERT INTO `results` (`participant_id`, `attempt_id`, `item_id`, `started_at`)
SELECT STRAIGHT_JOIN results.participant_id, results.attempt_id, results.item_id,
                     (SELECT MIN(started_at) FROM results AS descendant_results
                      WHERE descendant_results.participant_id = results.participant_id AND
                           descendant_results.attempt_id = results.attempt_id AND
                           descendant_results.item_id IN (SELECT child_item_id FROM items_ancestors WHERE ancestor_item_id = results.item_id)
                      ) AS new_started_at
                      FROM results
                      WHERE results.started_at IS NULL AND
                            EXISTS(
                                SELECT 1
                                FROM items_ancestors
                                    JOIN results AS child_results
                                        ON child_results.participant_id = results.participant_id AND
                                           child_results.attempt_id = results.attempt_id AND
                                           child_results.item_id = items_ancestors.child_item_id AND
                                           child_results.started_at IS NOT NULL
                                WHERE items_ancestors.ancestor_item_id = results.item_id AND
                                      EXISTS(
                                          SELECT 1
                                          FROM groups_ancestors_active
                                              JOIN permissions_generated
                                                  ON permissions_generated.group_id = groups_ancestors_active.ancestor_group_id AND
                                                     permissions_generated.item_id = child_results.item_id
                                          WHERE groups_ancestors_active.child_group_id = results.participant_id AND
                                                permissions_generated.can_view_generated_value >= 3
                                      )
                            ) AND
                            EXISTS(
                                SELECT 1 FROM groups_ancestors_active
                                    JOIN permissions_generated
                                        ON permissions_generated.group_id = groups_ancestors_active.ancestor_group_id AND
                                           permissions_generated.item_id = results.item_id
                                WHERE groups_ancestors_active.child_group_id = results.participant_id AND
                                      permissions_generated.can_view_generated_value >= 3
                            )
ON DUPLICATE KEY UPDATE started_at = VALUES(started_at);

-- +migrate Down

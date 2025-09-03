-- +migrate Up
DROP TRIGGER `after_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_groups` AFTER INSERT ON `groups_groups` FOR EACH ROW BEGIN
    IF NEW.`expires_at` > NOW() AND NOT NEW.`is_team_membership` THEN
        INSERT IGNORE INTO `results_propagate`
        SELECT `participant_id`, `attempt_id`, `results`.`item_id`, 'to_be_propagated' AS `state`
        FROM (
                 SELECT `item_id`
                 FROM (
                          SELECT DISTINCT `item_id`
                          FROM `results`
                                   JOIN `groups_ancestors_active`
                                        ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                           `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
                          FOR SHARE
                      ) AS `result_items`
                 WHERE EXISTS(
                         SELECT 1
                         FROM `permissions_generated`
                                  JOIN `groups_ancestors_active` AS `grand_ancestors`
                                       ON `grand_ancestors`.`child_group_id` = NEW.`parent_group_id` AND
                                          `grand_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                                  JOIN `items_ancestors`
                                       ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                         WHERE `items_ancestors`.`child_item_id` = `result_items`.`item_id`
                           AND `permissions_generated`.`can_view_generated` != 'none'
                         FOR SHARE
                     )
                   AND NOT EXISTS(
                         SELECT 1
                         FROM `permissions_generated`
                                  JOIN `groups_ancestors_active` AS `child_ancestors`
                                       ON `child_ancestors`.`child_group_id` = NEW.`child_group_id` AND
                                          `child_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                                  JOIN `items_ancestors`
                                       ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                         WHERE `items_ancestors`.`child_item_id` = `result_items`.`item_id`
                           AND `permissions_generated`.`can_view_generated` != 'none'
                         FOR SHARE
                     )
                 FOR SHARE
             ) AS `result_items_filtered`
        JOIN `results` ON `results`.`item_id` = `result_items_filtered`.`item_id`
        JOIN `groups_ancestors_active`
          ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
             `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
        FOR SHARE;
    END IF;
END
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER `after_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups_groups` AFTER INSERT ON `groups_groups` FOR EACH ROW BEGIN
    IF NEW.`expires_at` > NOW() THEN
        INSERT IGNORE INTO `results_propagate`
        SELECT `participant_id`, `attempt_id`, `results`.`item_id`, 'to_be_propagated' AS `state`
        FROM (
                 SELECT `item_id`
                 FROM (
                          SELECT DISTINCT `item_id`
                          FROM `results`
                                   JOIN `groups_ancestors_active`
                                        ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                                           `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`
                      ) AS `result_items`
                 WHERE EXISTS(
                         SELECT 1
                         FROM `permissions_generated`
                                  JOIN `groups_ancestors_active` AS `grand_ancestors`
                                       ON `grand_ancestors`.`child_group_id` = NEW.`parent_group_id` AND
                                          `grand_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                                  JOIN `items_ancestors`
                                       ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                         WHERE `items_ancestors`.`child_item_id` = `result_items`.`item_id`
                           AND `permissions_generated`.`can_view_generated` != 'none'
                     )
                   AND NOT EXISTS(
                         SELECT 1
                         FROM `permissions_generated`
                                  JOIN `groups_ancestors_active` AS `child_ancestors`
                                       ON `child_ancestors`.`child_group_id` = NEW.`child_group_id` AND
                                          `child_ancestors`.`ancestor_group_id` = `permissions_generated`.`group_id`
                                  JOIN `items_ancestors`
                                       ON `items_ancestors`.`ancestor_item_id` = `permissions_generated`.`item_id`
                         WHERE `items_ancestors`.`child_item_id` = `result_items`.`item_id`
                           AND `permissions_generated`.`can_view_generated` != 'none'
                     )
             ) AS `result_items_filtered`
                 JOIN `results` ON `results`.`item_id` = `result_items_filtered`.`item_id`
                 JOIN `groups_ancestors_active`
                      ON `groups_ancestors_active`.`child_group_id` = `results`.`participant_id` AND
                         `groups_ancestors_active`.`ancestor_group_id` = NEW.`child_group_id`;
    END IF;
END
-- +migrate StatementEnd

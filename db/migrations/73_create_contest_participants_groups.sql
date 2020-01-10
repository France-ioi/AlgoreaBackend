-- +migrate Up
# add new type 'ContestParticipants'
ALTER TABLE `groups` MODIFY COLUMN `type` enum('Class','Team','Club','Friends','Other','UserSelf','Base','ContestParticipants') NOT NULL;

# create a "contest participants" group for each contest
INSERT INTO `groups` (`id`, `name`, `type`, `team_item_id`)
    SELECT FLOOR(RAND(3) * 1000000000) + FLOOR(RAND(4) * 1000000000) * 1000000000,
           CONCAT(`items`.`id`, '-participants'), 'ContestParticipants', `items`.`id`
    FROM `items`
    WHERE `duration` IS NOT NULL
    ORDER BY `items`.`id`;

# link contests to their "contest participants" groups
UPDATE `items`
JOIN `groups` ON `groups`.`team_item_id` = `items`.`id` AND `groups`.`type` = 'ContestParticipants'
SET `items`.`contest_participants_group_id` = `groups`.`id`,
    `groups`.`team_item_id` = NULL
WHERE `items`.`duration` IS NOT NULL;

# Add all the participants into "contest participants" groups.
# Here we require groups_attempts to exist for each group_id-item_id pair in order to create groups_groups rows,
# but there is a pair in groups_items of the example database (group_id = 261836104618448530, item_id = 261836104618448530)
# for which that is wrong. `can_view:content` permission for this pair will be lost and the group 261836104618448530
# will not be added as a member of the item's contest participants group.
INSERT INTO `groups_groups` (`id`, `parent_group_id`, `child_group_id`, `expires_at`, `child_order`)
    SELECT FLOOR(RAND(5) * 1000000000) + FLOOR(RAND(6) * 1000000000) * 1000000000,
           `items`.`contest_participants_group_id`,
           `groups_attempts`.`group_id`,
           IFNULL(MIN(`finished_at`),
               DATE_ADD(
                   ADDTIME(MIN(`entered_at`), `duration`),
                   INTERVAL IFNULL(SUM(TIME_TO_SEC(`groups_contest_items`.`additional_time`)), 0) SECOND)),
           (SELECT IFNULL(MAX(`child_order`), 0)+1
            FROM `groups_groups` AS gg
            WHERE gg.`parent_group_id` = `items`.`contest_participants_group_id`)
    FROM `items`
    JOIN `groups_attempts` ON `groups_attempts`.`item_id` = `items`.`id` AND `groups_attempts`.`entered_at` IS NOT NULL
    JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `groups_attempts`.`group_id`
    LEFT JOIN `groups_contest_items` ON `groups_contest_items`.`group_id` = `groups_ancestors_active`.`ancestor_group_id` AND
                                   `groups_contest_items`.`item_id` = `items`.`id`
    WHERE `items`.`contest_participants_group_id` IS NOT NULL AND
          `items`.`duration` IS NOT NULL
    GROUP BY `groups_attempts`.`group_id`, `groups_attempts`.`item_id`
    ORDER BY `groups_attempts`.`group_id`, `groups_attempts`.`item_id`;

# remove 'content' permissions given to participants directly on entering
DELETE `permissions_granted`
FROM `permissions_granted`
WHERE `permissions_granted`.`item_id` IN (SELECT `items`.`id` FROM `items` WHERE `duration` IS NOT NULL) AND
      `permissions_granted`.`group_id` IN (
        SELECT `groups_groups`.`child_group_id`
        FROM `items`
        JOIN `groups_groups` ON `groups_groups`.`parent_group_id` = `items`.`contest_participants_group_id`
    ) AND
    `source_group_id` = -1 AND `can_view` = 'content' AND
    `can_grant_view` = 'none' AND `can_watch` = 'none' AND
    `can_edit` = 'none' AND `is_owner` = 0;

# give 'content' permissions to "contest participants" groups
INSERT INTO `permissions_granted` (`group_id`, `item_id`, `can_view`, `source_group_id`)
    SELECT `items`.`contest_participants_group_id`, `items`.`id`, 'content', -4
    FROM `items`
    WHERE `items`.`contest_participants_group_id` IS NOT NULL;

ALTER TABLE `groups_attempts` DROP COLUMN `finished_at`;

-- +migrate Down
DELETE `permissions_granted`
FROM `permissions_granted`
JOIN `items` ON `items`.id = `permissions_granted`.`item_id` AND
                `items`.`contest_participants_group_id` = `permissions_granted`.`group_id`;

INSERT INTO `permissions_granted` (`item_id`, `group_id`, `can_view`, `source_group_id`)
    SELECT `items`.`id`, `groups_groups`.`child_group_id`, 'content', -1
    FROM `items`
    JOIN `groups_groups` ON `groups_groups`.`parent_group_id` = `items`.`contest_participants_group_id`
    WHERE `items`.`duration` IS NOT NULL
ON DUPLICATE KEY UPDATE `can_view` = IF(`can_view_value` < 3 /* content */, 'content', `can_view`);

ALTER TABLE `groups_attempts`
    ADD COLUMN `finished_at` datetime DEFAULT NULL COMMENT 'When the item was finished, within this attempt.'
    AFTER `validated_at`;

UPDATE `groups_attempts`
JOIN `items` ON `items`.`id` = `groups_attempts`.`item_id` AND `items`.`duration` IS NOT NULL
JOIN `groups_groups` ON `groups_groups`.`child_group_id` = `groups_attempts`.`group_id` AND
                        `groups_groups`.`parent_group_id` = `items`.`contest_participants_group_id`
SET `finished_at` = IF(`groups_groups`.`expires_at` <= NOW(), `groups_groups`.`expires_at`, NULL);

DELETE `groups_groups`
FROM `groups_groups`
JOIN `items` ON `items`.`contest_participants_group_id` = `groups_groups`.`parent_group_id`;

DELETE `groups`
FROM `groups`
JOIN `items` ON `items`.`contest_participants_group_id` = `groups`.`id`;

UPDATE `items` SET `contest_participants_group_id` = NULL WHERE `contest_participants_group_id` IS NOT NULL;

ALTER TABLE `groups` MODIFY COLUMN `type` enum('Class','Team','Club','Friends','Other','UserSelf','Base') NOT NULL;

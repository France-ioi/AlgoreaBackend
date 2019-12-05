-- +migrate Up
CREATE TABLE `items_unlocking_rules` (
    `unlocking_item_id` BIGINT(20) NOT NULL,
    `unlocked_item_id` BIGINT(20) NOT NULL,
    `score` int(11) NOT NULL DEFAULT '100'
        COMMENT 'Score of the unlocking item from which the unlocked item is unlocked, i.e. can_view:content is given.',
    PRIMARY KEY (`unlocking_item_id`, `unlocked_item_id`),
    CONSTRAINT `fk_items_unlocking_rules_unlocking_item_id_items_id`
        FOREIGN KEY (`unlocking_item_id`) REFERENCES `items`(`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_items_unlocking_rules_unlocked_item_id_items_id`
        FOREIGN KEY (`unlocked_item_id`) REFERENCES `items`(`id`) ON DELETE CASCADE
);

INSERT INTO `items_unlocking_rules` (`unlocking_item_id`, `unlocked_item_id`, `score`)
    SELECT `items`.`id` AS `unlocking_item_id`,
           `ids`.`id` AS `unlocked_item_id`,
           `items`.`score_min_unlock` AS `score`
    FROM `items`
        JOIN JSON_TABLE(CONCAT('[', `items`.`unlocked_item_ids`, ']'), "$[*]" COLUMNS(`id` BIGINT PATH '$')) AS `ids`
        JOIN `items` AS `items_to_unlock` ON `items_to_unlock`.`id` = `ids`.`id`;

ALTER TABLE `items`
    DROP COLUMN `unlocked_item_ids`,
    DROP COLUMN `score_min_unlock`;

ALTER TABLE `groups_attempts`
    DROP COLUMN `has_unlocked_items`;

-- +migrate Down
ALTER TABLE `groups_attempts`
    ADD COLUMN `has_unlocked_items` tinyint(1) NOT NULL DEFAULT '0'
        COMMENT 'Whether the score of this attempt allows unlocking other items (score >= items.score_min_unlock)'
        AFTER `finished`;

UPDATE `groups_attempts`
JOIN (SELECT `unlocking_item_id`, MIN(`score`) AS `score` FROM `items_unlocking_rules` GROUP BY `unlocking_item_id`) AS `rules`
    ON `rules`.`unlocking_item_id` = `groups_attempts`.`item_id` AND `rules`.`score` <= `groups_attempts`.`score`
SET `groups_attempts`.`has_unlocked_items` = 1;

ALTER TABLE `items`
    ADD COLUMN `unlocked_item_ids` text
        COMMENT 'Comma-separated list of item_ids which will be unlocked if this item is validated'
        AFTER `validation_type`,
    ADD COLUMN `score_min_unlock` INT(11) NOT NULL DEFAULT '100'
        COMMENT 'Minimum score to obtain so that the item, indicated by "unlocked_item_ids", is actually unlocked'
        AFTER `unlocked_item_ids`;

UPDATE `items`
JOIN (
        SELECT `unlocking_item_id` AS `id`,
               GROUP_CONCAT(`unlocked_item_id`) AS `unlocked_item_ids`,
               MAX(`score`) AS `score_min_unlock`
        FROM `items_unlocking_rules`
        GROUP BY `unlocking_item_id`
    ) AS `rules` USING (`id`)
SET `items`.`unlocked_item_ids` = `rules`.`unlocked_item_ids`,
    `items`.`score_min_unlock` = `rules`.`score_min_unlock`;

DROP TABLE `items_unlocking_rules`;

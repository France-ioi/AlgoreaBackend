-- +migrate Up
ALTER TABLE `groups_attempts` ADD COLUMN `entered_at` DATETIME DEFAULT NULL
    COMMENT 'Time at which the group entered the contest' AFTER `autonomy`;

# The old data contains no more than one groups_attempts row per contest_participations row
UPDATE `groups_attempts` JOIN contest_participations USING(group_id, item_id)
SET `groups_attempts`.`entered_at` = `contest_participations`.`entered_at`,
    `groups_attempts`.`finished_at` = `contest_participations`.`finished_at`;

INSERT INTO `groups_attempts` (`id`, `group_id`, `item_id`, `entered_at`, `finished_at`, `order`)
SELECT (`contest_participations`.`group_id` + `contest_participations`.`item_id`) % 9223372036854775806 + 1,
       `contest_participations`.`group_id`, `contest_participations`.`item_id`,
       `contest_participations`.`entered_at`, `contest_participations`.`finished_at`, 1
FROM `contest_participations`
LEFT JOIN `groups_attempts` AS `existing_attempts` USING (`group_id`, `item_id`)
WHERE `existing_attempts`.`id` IS NULL;

DROP TABLE `contest_participations`;

-- +migrate Down
CREATE TABLE `contest_participations` (
  `group_id` bigint(20) NOT NULL,
  `item_id` bigint(20) NOT NULL,
  `entered_at` datetime NOT NULL COMMENT 'Time at which the group entered the contest',
  `finished_at` datetime DEFAULT NULL COMMENT 'Time at which the contest has been finished for the group',
  PRIMARY KEY (`group_id`,`item_id`)
) ENGINE=InnoDB COMMENT='Information on when teams or users entered contests';

INSERT INTO `contest_participations` (`group_id`, `item_id`, `entered_at`, `finished_at`)
SELECT `group_id`, `item_id`, `entered_at`, `finished_at`
FROM `groups_attempts`
WHERE `entered_at` IS NOT NULL;

ALTER TABLE `groups_attempts` DROP COLUMN `entered_at`;

-- +migrate Up
ALTER TABLE `groups_items` ADD COLUMN `sAdditionalTime` TIME DEFAULT NULL AFTER `sPropagateAccess`;

UPDATE `groups_items`
    JOIN `users_items` ON `users_items`.`idItem` = `groups_items`.`idItem`
    JOIN `users` ON `users`.`idGroupSelf` = `groups_items`.`idGroup` AND `users`.ID = `users_items`.`idUser`
    SET `groups_items`.`sAdditionalTime` = `users_items`.`sAdditionalTime`
    WHERE `groups_items`.`sAdditionalTime` IS NULL;

UPDATE `groups_items` JOIN (
        SELECT `groups_attempts`.`idGroup`,
           `groups_attempts`.`idItem`,
           MAX(`groups_attempts`.`sAdditionalTime`) AS maxAdditionalTime
        FROM `groups_attempts`
        GROUP BY `groups_attempts`.`idGroup`, `groups_attempts`.`idItem`
    ) AS `max_times`
    ON `max_times`.`idGroup` = `groups_items`.`idGroup` AND `max_times`.`idItem` = `groups_items`.`idItem`
    SET `groups_items`.`sAdditionalTime` = GREATEST(
            IFNULL(`groups_items`.`sAdditionalTime`, `max_times`.`maxAdditionalTime`),
            `max_times`.`maxAdditionalTime`
        )
    WHERE `max_times`.`maxAdditionalTime` IS NOT NULL;

ALTER TABLE `groups_attempts` DROP COLUMN `sAdditionalTime`;
ALTER TABLE `history_groups_attempts` DROP COLUMN `sAdditionalTime`;
ALTER TABLE `users_items` DROP COLUMN `sAdditionalTime`;
ALTER TABLE `history_users_items` DROP COLUMN `sAdditionalTime`;

-- +migrate Down
ALTER TABLE `groups_items` DROP COLUMN `sAdditionalTime`;
ALTER TABLE `users_items` ADD COLUMN `sAdditionalTime` TIME DEFAULT NULL AFTER `sLastHintDate`;
ALTER TABLE `history_users_items` ADD COLUMN `sAdditionalTime` TIME DEFAULT NULL AFTER `sLastHintDate`;
ALTER TABLE `groups_attempts` ADD COLUMN `sAdditionalTime` DATETIME DEFAULT NULL AFTER `sLastHintDate`;
ALTER TABLE `history_groups_attempts` ADD COLUMN `sAdditionalTime` DATETIME DEFAULT NULL AFTER `sLastHintDate`;

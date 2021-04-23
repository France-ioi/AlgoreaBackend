-- +migrate Up
ALTER TABLE `users`
    DROP COLUMN `allow_subgroups`;

-- +migrate Down
ALTER TABLE `users`
    ADD COLUMN `allow_subgroups` TINYINT DEFAULT NULL COMMENT 'Allow to create subgroups' AFTER `creator_id`;

UPDATE `users` SET users.`allow_subgroups` = 1 WHERE `group_id` = 670968966872011405;

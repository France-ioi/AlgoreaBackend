-- +migrate Up
ALTER TABLE `items`
    DROP COLUMN `qualified_group_id`,
    DROP COLUMN `contest_opens_at`,
    DROP COLUMN `contest_closes_at`;

-- +migrate Down
ALTER TABLE `items`
    ADD COLUMN `qualified_group_id` bigint(20) DEFAULT NULL
        COMMENT 'group id in which "qualified" users will belong. contest_entering_condition dictates how many of a team''s members must be "qualified" in order to start the item.'
        AFTER `teams_editable`,
    ADD COLUMN  `contest_opens_at` datetime DEFAULT NULL
        COMMENT 'When access to this item can be opened'
        AFTER `has_attempts`,
    ADD COLUMN `contest_closes_at` datetime DEFAULT NULL
        COMMENT 'Until when people can start participating to this contest'
        AFTER `duration`;

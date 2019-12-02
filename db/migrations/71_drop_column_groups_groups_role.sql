-- +migrate Up
DELETE FROM `groups_groups` WHERE `role` != 'member';
ALTER TABLE `groups_groups` DROP COLUMN `role`;

DROP VIEW IF EXISTS groups_groups_active;
CREATE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;

-- +migrate Down
ALTER TABLE `groups_groups`
    ADD COLUMN `role` enum('manager','owner','member','observer') NOT NULL DEFAULT 'member'
        COMMENT 'Role that the child has relative to the parent.' AFTER `child_order`;

DROP VIEW IF EXISTS groups_groups_active;
CREATE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;

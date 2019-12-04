-- +migrate Up
DELETE FROM `groups_groups` WHERE `role` != 'member';
ALTER TABLE `groups_groups`
  COMMENT 'Parent-child (N-N) relationships between groups (acyclic graph).',
  DROP COLUMN `role`;

DROP VIEW IF EXISTS groups_groups_active;
CREATE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;

-- +migrate Down
ALTER TABLE `groups_groups`
  COMMENT 'Parent-child (N-N) relationships between groups (acyclic graph). It includes potential relationships such as invitations or requests to join groups, as well as past relationships.',
  ADD COLUMN `role` enum('manager','owner','member','observer') NOT NULL DEFAULT 'member'
    COMMENT 'Role that the child has relative to the parent.' AFTER `child_order`;

DROP VIEW IF EXISTS groups_groups_active;
CREATE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;

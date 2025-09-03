-- +migrate Up
ALTER TABLE `groups_groups` ADD COLUMN `is_team_membership` TINYINT(1) NOT NULL DEFAULT 0
  COMMENT 'true if the parent group is a team'
  AFTER `expires_at`;

-- +migrate Down
ALTER TABLE `groups_groups` DROP COLUMN `is_team_membership`;

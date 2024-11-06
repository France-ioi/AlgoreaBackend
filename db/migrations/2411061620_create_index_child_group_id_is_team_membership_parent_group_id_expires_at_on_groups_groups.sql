-- +migrate Up
ALTER TABLE `groups_groups` ADD INDEX `child_group_id_is_team_membership_parent_group_id_expires_at`
  (`child_group_id`,`is_team_membership`,`parent_group_id`,`expires_at`);

-- +migrate Down
ALTER TABLE `groups_groups` DROP INDEX `child_group_id_is_team_membership_parent_group_id_expires_at`;

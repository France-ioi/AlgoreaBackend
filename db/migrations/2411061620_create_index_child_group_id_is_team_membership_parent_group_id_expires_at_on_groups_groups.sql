-- +migrate Up
ALTER TABLE `groups_groups` ADD INDEX `child_group_id_is_team_membership_parent_group_id_expires_at`
  (`child_group_id`,`is_team_membership`,`parent_group_id`,`expires_at`);

-- +migrate Down
ALTER TABLE `groups_groups`
  DROP FOREIGN KEY `fk_groups_groups_child_group_id_groups_id`,
  DROP INDEX `child_group_id_is_team_membership_parent_group_id_expires_at`;
ALTER TABLE `groups_groups`
  ADD CONSTRAINT `fk_groups_groups_child_group_id_groups_id` FOREIGN KEY (`child_group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE;

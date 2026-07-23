-- +goose Up
ALTER TABLE `group_membership_changes`
  ADD INDEX `group_id_member_id_action_at` (`group_id`,`member_id`,`action`,`at`);

-- +goose Down
ALTER TABLE `group_membership_changes`
  DROP INDEX `group_id_member_id_action_at`;

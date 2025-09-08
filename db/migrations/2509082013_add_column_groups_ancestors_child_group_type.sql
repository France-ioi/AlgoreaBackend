-- +goose Up
ALTER TABLE `groups_ancestors`
  ADD COLUMN `child_group_type` ENUM('Class','Team','Club','Friends','Other','User','Session','Base','ContestParticipants') DEFAULT NULL
  COMMENT 'The type of the child group in the relationship' AFTER `child_group_id`;

-- +goose Down
ALTER TABLE `groups_ancestors`
  DROP COLUMN `child_group_type`;

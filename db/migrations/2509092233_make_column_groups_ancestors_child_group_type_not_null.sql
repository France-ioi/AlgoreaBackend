-- +goose Up
ALTER TABLE `groups_ancestors`
  MODIFY COLUMN `child_group_type` ENUM('Class','Team','Club','Friends','Other','User','Session','Base','ContestParticipants') NOT NULL
  COMMENT 'The type of the child group in the relationship (we duplicate groups.type to improve performance)' AFTER `child_group_id`;

-- +goose Down
ALTER TABLE `groups_ancestors`
  MODIFY COLUMN `child_group_type` ENUM('Class','Team','Club','Friends','Other','User','Session','Base','ContestParticipants') DEFAULT NULL
  COMMENT 'The type of the child group in the relationship (we duplicate groups.type to improve performance)' AFTER `child_group_id`;

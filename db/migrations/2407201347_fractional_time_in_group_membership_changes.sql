-- +migrate Up
ALTER TABLE `group_membership_changes`
  MODIFY COLUMN `at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'Time of the action';

-- +migrate Down
ALTER TABLE `group_membership_changes`
  MODIFY COLUMN `at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Time of the action';

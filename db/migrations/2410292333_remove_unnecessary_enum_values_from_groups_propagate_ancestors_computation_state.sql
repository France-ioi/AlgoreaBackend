-- +migrate Up
ALTER TABLE `groups_propagate` MODIFY COLUMN `ancestors_computation_state` ENUM('todo', 'done') NOT NULL;

-- +migrate Down
ALTER TABLE `groups_propagate` MODIFY COLUMN `ancestors_computation_state` ENUM('todo', 'done', 'processing', '') NOT NULL;

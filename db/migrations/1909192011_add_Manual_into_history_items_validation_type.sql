-- +migrate Up
ALTER TABLE `history_items` MODIFY COLUMN `validation_type` enum('None','All','AllButOne','Categories','One','Manual') NOT NULL DEFAULT 'All';

-- +migrate Down
UPDATE `history_items` SET `validation_type` = 'None' WHERE `validation_type` = 'Manual';
ALTER TABLE `history_items` MODIFY COLUMN `validation_type` enum('None','All','AllButOne','Categories','One') NOT NULL DEFAULT 'All';

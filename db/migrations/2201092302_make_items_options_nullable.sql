-- +migrate Up
ALTER TABLE `items`
  MODIFY COLUMN `options` TEXT DEFAULT NULL
    COMMENT 'Options passed to the task, formatted as a JSON object';

-- +migrate Down
UPDATE `items` SET `options` = '{}' WHERE `options` IS NULL;
ALTER TABLE `items`
  MODIFY COLUMN `options` TEXT NOT NULL
    COMMENT 'Options passed to the task, formatted as a JSON object';

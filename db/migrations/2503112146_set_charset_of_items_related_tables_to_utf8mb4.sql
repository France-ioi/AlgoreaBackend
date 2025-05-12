-- +migrate Up
SET foreign_key_checks = 0;
ALTER TABLE `items` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `items`
  MODIFY `options` text COMMENT 'Options passed to the task, formatted as a JSON object',
  MODIFY `repository_path` text;
ALTER TABLE `items_ancestors` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `items_items` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `items_propagate` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `items_strings` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `items_strings`
  MODIFY `description` text COMMENT 'Description of the item in the specified language',
  MODIFY `edu_comment` text COMMENT 'Information about what this item teaches, in the specified language.';
ALTER TABLE `languages` CONVERT TO CHARACTER SET utf8mb4;
SET foreign_key_checks = 1;

-- +migrate Down
SET foreign_key_checks = 0;
ALTER TABLE `languages` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `items_strings` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `items_propagate` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `items_items` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `items_ancestors` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `items` CONVERT TO CHARACTER SET utf8mb3;
SET foreign_key_checks = 1;

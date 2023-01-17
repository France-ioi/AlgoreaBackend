-- +migrate Up
UPDATE `items_strings` SET `image_url` = SUBSTRING(`image_url`, 1, 2048) WHERE LENGTH(`image_url`) > 2048;
ALTER TABLE `items_strings`
  MODIFY COLUMN `image_url` varchar(2048) DEFAULT NULL COMMENT 'Url of a small image associated with this item.';

-- +migrate Down
ALTER TABLE `items_strings`
  MODIFY COLUMN `image_url` TEXT DEFAULT NULL COMMENT 'Url of a small image associated with this item.';

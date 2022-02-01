-- +migrate Up
ALTER TABLE `items`
  MODIFY COLUMN `url` varchar(2048) DEFAULT NULL COMMENT 'Url of the item, as will be loaded in the iframe';

-- +migrate Down
UPDATE `items` SET `url` = SUBSTRING(`url`, 1, 200) WHERE LENGTH(`url`) > 200;
ALTER TABLE `items`
  MODIFY COLUMN `url` varchar(200) DEFAULT NULL COMMENT 'Url of the item, as will be loaded in the iframe';

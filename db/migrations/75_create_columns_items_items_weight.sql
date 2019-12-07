-- +migrate Up
ALTER TABLE `items_items`
    ADD COLUMN `weight` FLOAT NOT NULL DEFAULT '1' COMMENT 'Weight of this child in his parent\'s score computation'
        AFTER `difficulty`;

-- +migrate Down
ALTER TABLE `items_items` DROP COLUMN `weight`;

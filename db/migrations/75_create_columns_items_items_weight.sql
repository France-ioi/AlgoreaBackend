-- +migrate Up
ALTER TABLE `items_items`
    ADD COLUMN `score_weight` TINYINT NOT NULL DEFAULT '1' COMMENT 'Weight of this child in his parent\'s score computation'
        AFTER `difficulty`;

-- +migrate Down
ALTER TABLE `items_items` DROP COLUMN `score_weight`;

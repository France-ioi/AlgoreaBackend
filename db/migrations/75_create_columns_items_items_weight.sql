-- +migrate Up
ALTER TABLE `items_items`
    ADD COLUMN `score_weight` TINYINT UNSIGNED NOT NULL DEFAULT '1' COMMENT 'Weight of this child in his parent\'s score computation'
        AFTER `category`,
    DROP COLUMN `difficulty`;

-- +migrate Down
ALTER TABLE `items_items`
    DROP COLUMN `score_weight`,
    ADD COLUMN `difficulty` int(11) NOT NULL DEFAULT '0'
        COMMENT 'Indication of the difficulty of this item relative to its siblings.';

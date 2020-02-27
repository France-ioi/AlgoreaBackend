-- +migrate Up
ALTER TABLE `items_items` MODIFY COLUMN `iDifficulty` int(11) NOT NULL DEFAULT '0' COMMENT 'Indication of the difficulty of this item relative to its siblings.';
ALTER TABLE `history_items_items` MODIFY COLUMN `iDifficulty` int(11) NOT NULL DEFAULT '0';

-- +migrate Down
ALTER TABLE `items_items` MODIFY COLUMN `iDifficulty` int(11) NOT NULL COMMENT 'Indication of the difficulty of this item relative to its siblings.';
ALTER TABLE `history_items_items` MODIFY COLUMN `iDifficulty` int(11) NOT NULL;

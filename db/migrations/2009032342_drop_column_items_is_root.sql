-- +migrate Up
ALTER TABLE `items` DROP COLUMN `is_root`;

-- +migrate Down
ALTER TABLE `items`
    ADD COLUMN `is_root` TINYINT(1) NOT NULL DEFAULT '0'
        COMMENT 'Whether this item is intended to be a root chapter (in order to detect real orphans more easily)'
        AFTER `type`;

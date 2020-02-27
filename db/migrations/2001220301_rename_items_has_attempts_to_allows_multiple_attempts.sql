-- +migrate Up
ALTER TABLE `items`
    CHANGE COLUMN `has_attempts` `allows_multiple_attempts` TINYINT(1) DEFAULT 0 NOT NULL
        COMMENT 'Whether participants can create multiple attempts when working on this item';

-- +migrate Down
ALTER TABLE `items`
    CHANGE COLUMN `allows_multiple_attempts` `has_attempts` TINYINT(1) DEFAULT 0 NOT NULL
        COMMENT 'Whether team participation is mandatory. Whether users can create multiple attempts when working on this item.';

-- +migrate Up
ALTER TABLE `attempts` ADD COLUMN `ended_at` DATETIME DEFAULT NULL
    COMMENT 'Time at which the attempt was (typically manually) ended'
        AFTER `created_at`;

-- +migrate Down
ALTER TABLE `attempts` DROP COLUMN `ended_at`;

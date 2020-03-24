-- +migrate Up
ALTER TABLE `attempts` ADD COLUMN `allows_submissions_until` DATETIME NOT NULL DEFAULT '9999-12-31 23:59:59'
    COMMENT 'Time until which the participant can submit an answer on this attempt'
    AFTER `created_at`;

-- +migrate Down
ALTER TABLE `attempts` DROP COLUMN `allows_submissions_until`;

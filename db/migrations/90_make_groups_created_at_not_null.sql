-- +migrate Up
ALTER TABLE `groups`
    MODIFY COLUMN `created_at` DATETIME NOT NULL DEFAULT NOW();

-- +migrate Down
ALTER TABLE `groups`
    MODIFY COLUMN `created_at` DATETIME DEFAULT NULL;

-- +migrate Up
UPDATE `items` SET `full_screen` = 'default' WHERE `full_screen` = '';
ALTER TABLE `items` MODIFY `full_screen`
    ENUM('forceYes','forceNo','default') NOT NULL DEFAULT 'default'
    COMMENT 'Whether the item should be loaded in full screen mode (without the navigation panel and most of the top header). By default, tasks are displayed in full screen, but not chapters.';

-- +migrate Down
ALTER TABLE `items` MODIFY `full_screen`
    ENUM('forceYes','forceNo','default','') NOT NULL DEFAULT 'default'
    COMMENT 'Whether the item should be loaded in full screen mode (without the navigation panel and most of the top header). By default, tasks are displayed in full screen, but not chapters.';

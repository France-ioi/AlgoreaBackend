-- +migrate Up

ALTER TABLE `items`
    DROP COLUMN `show_difficulty`,
    DROP COLUMN `show_source`,
    DROP COLUMN `transparent_folder`;

-- +migrate Down

ALTER TABLE `items`
  ADD COLUMN `show_difficulty` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Display an indication about the difficulty of each child relative to the other children' AFTER `full_screen`,
  ADD COLUMN `show_source` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'If false, we hide any information about the origin of tasks for this item and its descendants. Intended to be used during contests, so that users can''t find the solution online.' AFTER `show_difficulty`,
  ADD COLUMN `transparent_folder` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT 'Whether the breadcrumbs should hide this folder' AFTER `title_bar_visible`
  ;

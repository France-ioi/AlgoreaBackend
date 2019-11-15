-- +migrate Up

ALTER TABLE `items`
    DROP COLUMN `transparent_folder`;

-- +migrate Down

ALTER TABLE `items`
  ADD COLUMN `transparent_folder` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT 'Whether the breadcrumbs should hide this folder' AFTER `title_bar_visible`
  ;

-- +migrate Up

ALTER TABLE `items`
    DROP COLUMN `show_difficulty`,
    DROP COLUMN `show_source`,
    DROP COLUMN `validation_min`,
    DROP COLUMN `contest_phase`,
    DROP COLUMN `level`,
    DROP COLUMN `transparent_folder`;

-- +migrate Down

ALTER TABLE `items`
  ADD COLUMN `show_difficulty` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Display an indication about the difficulty of each child relative to the other children' AFTER `full_screen`,
  ADD COLUMN `show_source` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'If false, we hide any information about the origin of tasks for this item and its descendants. Intended to be used during contests, so that users can''t find the solution online.' AFTER `show_difficulty`,
  ADD COLUMN `validation_min` int(11) DEFAULT NULL COMMENT 'Minimum score to obtain so that the item is considered as validated.' AFTER `validation_type`,
  ADD COLUMN `contest_phase` enum('Running','Analysis','Closed') NOT NULL COMMENT 'In running mode, users can only access to the tasks during their participation. In Analysis, they may look at their submissions, or try new attemps. Not sure what Closed would be.' AFTER show_user_infos,
  ADD COLUMN `level` int(11) DEFAULT NULL COMMENT 'On our old website, chapters were attached to a level, and we could displayed stats specific to a level.' AFTER `contest_phase`,
  ADD COLUMN `transparent_folder` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT 'Whether the breadcrumbs should hide this folder' AFTER `title_bar_visible`
  ;

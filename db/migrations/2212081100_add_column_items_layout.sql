-- +migrate Up
ALTER TABLE `items`
  ADD COLUMN `children_layout` ENUM('List', 'Grid') DEFAULT 'List'
    COMMENT 'How the children list are displayed (for chapters and skills)'
    AFTER `show_user_infos`;

-- +migrate Down
ALTER TABLE `items` DROP COLUMN `children_layout`;

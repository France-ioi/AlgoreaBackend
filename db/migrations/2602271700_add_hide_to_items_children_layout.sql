-- +goose Up
ALTER TABLE `items`
  MODIFY COLUMN `children_layout` enum('List','Grid','Hide') DEFAULT 'List'
    COMMENT 'How the children list are displayed (for chapters and skills)';

-- +goose Down
UPDATE `items` SET `children_layout` = 'List' WHERE `children_layout` = 'Hide';
ALTER TABLE `items`
  MODIFY COLUMN `children_layout` enum('List','Grid') DEFAULT 'List'
    COMMENT 'How the children list are displayed (for chapters and skills)';

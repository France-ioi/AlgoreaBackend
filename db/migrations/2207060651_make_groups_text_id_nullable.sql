-- +migrate Up
ALTER TABLE `groups`
  MODIFY COLUMN `text_id` VARCHAR(255) NULL DEFAULT NULL
    COMMENT 'Internal text id for special groups. Used to refer o them and avoid breaking features if an admin renames the group';

UPDATE `groups` SET `text_id` = NULL WHERE `text_id` = '';

-- +migrate Down
UPDATE `groups` SET `text_id` = '' WHERE `text_id` IS NULL;

ALTER TABLE `groups`
  MODIFY COLUMN `text_id` VARCHAR(255) NOT NULL DEFAULT ''
    COMMENT 'Internal text id for special groups. Used to refer o them and avoid breaking features if an admin renames the group';

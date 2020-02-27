-- +migrate Up
ALTER TABLE `items`
    CHANGE COLUMN `group_code_enter` `prompt_to_join_group_by_code` TINYINT(1) NOT NULL DEFAULT '0'
        COMMENT 'Whether the UI should display a form for joining a group by code on the item page';

-- +migrate Down
ALTER TABLE `items`
    CHANGE COLUMN `prompt_to_join_group_by_code` `group_code_enter` TINYINT(1) DEFAULT '0'
        COMMENT 'Whether users can enter through a group code';

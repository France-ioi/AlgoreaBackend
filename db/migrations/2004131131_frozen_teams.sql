-- +migrate Up
ALTER TABLE `groups` ADD COLUMN `frozen_membership` TINYINT(1) DEFAULT 0
    COMMENT 'Whether members can be added/removed to the group (intended for teams)'
    AFTER `require_members_to_join_parent`;

ALTER TABLE `items`
    CHANGE COLUMN `teams_editable` `entry_frozen_teams` TINYINT(1) DEFAULT 0
    COMMENT 'Whether teams require to be have `frozen_membership` for entering';

UPDATE `items` SET items.`entry_frozen_teams` = NOT items.`entry_frozen_teams`;

-- +migrate Down
ALTER TABLE `items`
    CHANGE COLUMN `entry_frozen_teams` `teams_editable` TINYINT(1) DEFAULT 0
        COMMENT 'Whether users can create and edit their teams';

UPDATE `items` SET items.`teams_editable` = NOT items.`teams_editable`;

ALTER TABLE `groups` DROP COLUMN `frozen_membership`;

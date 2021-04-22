-- +migrate Up
ALTER TABLE `items`
    MODIFY COLUMN `entry_frozen_teams` TINYINT(1) NOT NULL DEFAULT '0'
        COMMENT 'Whether teams require to have `frozen_membership` for entering';

-- +migrate Down
ALTER TABLE `items`
    MODIFY COLUMN `entry_frozen_teams` TINYINT(1) DEFAULT '0'
        COMMENT 'Whether teams require to be have `frozen_membership` for entering';

-- +migrate Up
ALTER TABLE `group_managers`
    ADD COLUMN `can_manage_value` TINYINT(3) UNSIGNED GENERATED ALWAYS AS ((`can_manage` + 0))
        VIRTUAL NOT NULL COMMENT 'can_manage as an integer (to use comparison operators)';

-- +migrate Down
ALTER TABLE `group_managers` DROP COLUMN `can_manage_value`;

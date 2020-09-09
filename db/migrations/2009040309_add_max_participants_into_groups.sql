-- +migrate Up
ALTER TABLE `groups`
    ADD COLUMN `max_participants` INT(10) UNSIGNED DEFAULT NULL
        COMMENT 'The maximum number of participants (users and teams) in this group (strict limit if `enforce_max_participants`)'
        AFTER `frozen_membership`,
    ADD COLUMN `enforce_max_participants` TINYINT(1) DEFAULT 0
        COMMENT 'Whether the number of participants is a strict constraint'
        AFTER `max_participants`,
    ADD CONSTRAINT `cs_can_enforce_max_participants` CHECK (NOT `enforce_max_participants` OR `max_participants` IS NOT NULL);

-- +migrate Down
ALTER TABLE `groups`
    DROP CONSTRAINT `cs_can_enforce_max_participants`,
    DROP COLUMN `enforce_max_participants`,
    DROP COLUMN `max_participants`;

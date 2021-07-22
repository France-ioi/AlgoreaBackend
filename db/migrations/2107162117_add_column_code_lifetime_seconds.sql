-- +migrate Up
ALTER TABLE `groups`
    ADD COLUMN `code_lifetime_seconds` INT DEFAULT NULL
        COMMENT 'How long after the first use of the code it will expire (in seconds), NULL means infinity'
        AFTER `code_lifetime`;
UPDATE `groups` SET `groups`.`code_lifetime_seconds` = TIME_TO_SEC(`groups`.`code_lifetime`);

-- +migrate Down
ALTER TABLE `groups` DROP COLUMN `code_lifetime_seconds`;

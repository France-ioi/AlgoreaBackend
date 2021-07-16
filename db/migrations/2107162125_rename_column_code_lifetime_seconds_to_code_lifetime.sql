-- +migrate Up
ALTER TABLE `groups`
    DROP COLUMN `code_lifetime`,
    RENAME COLUMN `code_lifetime_seconds` TO `code_lifetime`;

-- +migrate Down
ALTER TABLE `groups`
    RENAME COLUMN `code_lifetime` TO `code_lifetime_seconds`,
    ADD COLUMN `code_lifetime` TIME DEFAULT NULL
        COMMENT 'How long after the first use of the code it will expire'
        AFTER `code`;
UPDATE `groups` SET `groups`.`code_lifetime` = SEC_TO_TIME(`groups`.`code_lifetime_seconds`);

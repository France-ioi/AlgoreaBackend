-- +migrate Up
ALTER TABLE `sessions`
    ADD COLUMN `use_cookie` TINYINT(1) NOT NULL DEFAULT 0 AFTER `issuer`,
    ADD COLUMN `cookie_secure` TINYINT(1) NOT NULL DEFAULT 0 AFTER `use_cookie`,
    ADD COLUMN `cookie_same_site` TINYINT(1) NOT NULL DEFAULT 0 AFTER `cookie_secure`,
    ADD COLUMN `cookie_domain` VARCHAR(255) DEFAULT NULL AFTER `cookie_same_site`,
    ADD COLUMN `cookie_path` VARCHAR(255) DEFAULT NULL AFTER `cookie_domain`;

-- +migrate Down
ALTER TABLE `sessions`
    DROP COLUMN `use_cookie`,
    DROP COLUMN `cookie_secure`,
    DROP COLUMN `cookie_same_site`,
    DROP COLUMN `cookie_domain`,
    DROP COLUMN `cookie_path`;

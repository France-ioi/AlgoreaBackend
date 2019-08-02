-- +migrate Up
ALTER TABLE `history_groups_login_prefixes` ADD INDEX  `idGroup` (`idGroup`);

-- +migrate Down
ALTER TABLE `history_groups_login_prefixes` DROP INDEX `idGroup`;

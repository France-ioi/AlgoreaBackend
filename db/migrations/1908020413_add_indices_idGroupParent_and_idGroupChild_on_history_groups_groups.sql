-- +migrate Up
ALTER TABLE `history_groups_groups` ADD INDEX  `idGroupParent` (`idGroupParent`);
ALTER TABLE `history_groups_groups` ADD INDEX  `idGroupChild` (`idGroupChild`);

-- +migrate Down
ALTER TABLE `history_groups_groups` DROP INDEX `idGroupChild`;
ALTER TABLE `history_groups_groups` DROP INDEX `idGroupParent`;

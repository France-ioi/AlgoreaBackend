-- +migrate Up
ALTER TABLE `groups_groups` ADD INDEX  `ParentOrder` (`idGroupParent`, `iChildOrder`);

-- +migrate Down
ALTER TABLE `groups_groups` DROP INDEX `ParentOrder`;

-- +migrate Up
ALTER TABLE `history_groups_attempts` ADD INDEX `ID` (`ID`);

-- +migrate Down
ALTER TABLE `history_groups_attempts` DROP INDEX `ID`;

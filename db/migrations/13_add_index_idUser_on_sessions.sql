-- +migrate Up
ALTER TABLE `sessions` ADD INDEX `idUser` (`idUser`);

-- +migrate Down
ALTER TABLE `sessions` DROP INDEX `idUser`;

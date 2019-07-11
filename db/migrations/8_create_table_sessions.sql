-- +migrate Up
CREATE TABLE `sessions` (
    `sAccessToken` VARCHAR(64) NOT NULL,
    `idUser` BIGINT(20) NOT NULL,
    `sExpirationDate` DATETIME NOT NULL,
    PRIMARY KEY (`sAccessToken`),
    KEY `sExpirationDate` (`sExpirationDate`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- +migrate Down
DROP TABLE `sessions`;

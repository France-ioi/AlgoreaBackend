-- +migrate Up
CREATE TABLE `sessions` (
    `sAccessToken` VARCHAR(64) NOT NULL,
    `idUser` BIGINT(20) NOT NULL,
    `sExpirationDate` DATETIME NOT NULL,
    `sIssuedAtDate` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `sIssuer` ENUM('backend', 'login-module'),
    PRIMARY KEY (`sAccessToken`),
    KEY `sExpirationDate` (`sExpirationDate`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- +migrate Down
DROP TABLE `sessions`;

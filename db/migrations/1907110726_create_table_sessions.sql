-- +migrate Up
CREATE TABLE `sessions` (
  `sAccessToken` VARBINARY(64) NOT NULL,
  `idUser` BIGINT(20) NOT NULL,
  `sExpirationDate` DATETIME NOT NULL,
  `sIssuedAtDate` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `sIssuer` ENUM('backend', 'login-module'),
  PRIMARY KEY (`sAccessToken`),
  KEY `sExpirationDate` (`sExpirationDate`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT 'Access tokens (short lifetime) distributed to users';

-- +migrate Down
DROP TABLE `sessions`;

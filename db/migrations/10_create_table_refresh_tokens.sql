-- +migrate Up
CREATE TABLE `refresh_tokens` (
  `idUser` BIGINT(20) NOT NULL,
  `sRefreshToken` VARBINARY(2000) NOT NULL,
  PRIMARY KEY (`idUser`),
  KEY `sRefreshTokenPrefix` (`sRefreshToken`(767))
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- +migrate Down
DROP TABLE `refresh_tokens`;

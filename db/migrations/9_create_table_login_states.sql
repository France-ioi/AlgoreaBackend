-- +migrate Up
CREATE TABLE `login_states` (
  `sCookie` BINARY(32) NOT NULL,
  `sState` BINARY(32) NOT NULL,
  `sExpirationDate` DATETIME NOT NULL,
  PRIMARY KEY (`sCookie`),
  KEY `sExpirationDate` (`sExpirationDate`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT 'States used in OAuth authorization requests to prevent CSRF attacks';

-- +migrate Down
DROP TABLE `login_states`;

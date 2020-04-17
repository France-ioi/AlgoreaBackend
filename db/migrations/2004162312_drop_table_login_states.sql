-- +migrate Up
DROP TABLE `login_states`;

-- +migrate Down
CREATE TABLE `login_states` (
  `cookie` BINARY(32) NOT NULL,
  `state` BINARY(32) NOT NULL,
  `expires_at` DATETIME NOT NULL,
  PRIMARY KEY (`cookie`),
  KEY `expires_at` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='States used in OAuth authorization requests to prevent CSRF attacks';

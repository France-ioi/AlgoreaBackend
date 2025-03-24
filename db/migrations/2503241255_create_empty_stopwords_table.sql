-- +migrate Up
CREATE TABLE `stopwords`(value VARCHAR(30)) ENGINE = INNODB;

-- +migrate Down
DROP TABLE `stopwords`;

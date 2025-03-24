-- +migrate Up
SET GLOBAL innodb_ft_server_stopword_table = CONCAT(DATABASE(), '/stopwords');

ALTER TABLE `items_strings` DROP INDEX `fullTextTitle`;
CREATE FULLTEXT INDEX `fullTextTitle` ON `items_strings`(`title`);
ALTER TABLE `groups` DROP INDEX `fullTextName`;
CREATE FULLTEXT INDEX `fullTextName` ON `groups`(`name`);

-- +migrate Down
SET GLOBAL innodb_ft_server_stopword_table = NULL;

ALTER TABLE `items_strings` DROP INDEX `fullTextTitle`;
CREATE FULLTEXT INDEX `fullTextTitle` ON `items_strings`(`title`);
ALTER TABLE `groups` DROP INDEX `fullTextName`;
CREATE FULLTEXT INDEX `fullTextName` ON `groups`(`name`);

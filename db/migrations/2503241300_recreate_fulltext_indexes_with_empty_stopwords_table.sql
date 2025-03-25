-- +migrate Up
SET @old_stopword_table := @@innodb_ft_server_stopword_table;
SET GLOBAL innodb_ft_server_stopword_table = CONCAT(DATABASE(), '/stopwords');

ALTER TABLE `items_strings` DROP INDEX `fullTextTitle`;
CREATE FULLTEXT INDEX `fullTextTitle` ON `items_strings`(`title`);
ALTER TABLE `groups` DROP INDEX `fullTextName`;
CREATE FULLTEXT INDEX `fullTextName` ON `groups`(`name`);

SET GLOBAL innodb_ft_server_stopword_table = @old_stopword_table;

-- +migrate Down
SET @old_stopword_table := @@innodb_ft_server_stopword_table;
SET GLOBAL innodb_ft_server_stopword_table = NULL;

ALTER TABLE `items_strings` DROP INDEX `fullTextTitle`;
CREATE FULLTEXT INDEX `fullTextTitle` ON `items_strings`(`title`);
ALTER TABLE `groups` DROP INDEX `fullTextName`;
CREATE FULLTEXT INDEX `fullTextName` ON `groups`(`name`);

SET GLOBAL innodb_ft_server_stopword_table = @old_stopword_table;

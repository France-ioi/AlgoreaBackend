-- +goose Up
ALTER TABLE `items_strings`
  MODIFY COLUMN `description` MEDIUMTEXT
    COMMENT 'Description of the item in the specified language';

-- MODIFY COLUMN rebuilds FULLTEXT indexes with the session stopword list; restore the
-- empty stopwords table used by schema.sql so tokens like "with" stay searchable.
SET @old_stopword_table := @@innodb_ft_user_stopword_table;
SET SESSION innodb_ft_user_stopword_table = CONCAT(DATABASE(), '/stopwords');
ALTER TABLE `items_strings` DROP INDEX `fullTextTitle`;
CREATE FULLTEXT INDEX `fullTextTitle` ON `items_strings`(`title`);
SET SESSION innodb_ft_user_stopword_table = @old_stopword_table;

-- +goose Down
-- Rollback requires every description to be ≤ 65,535 bytes, or MySQL may fail with "Data too long" / truncate.
ALTER TABLE `items_strings`
  MODIFY COLUMN `description` TEXT
    COMMENT 'Description of the item in the specified language';

SET @old_stopword_table := @@innodb_ft_user_stopword_table;
SET SESSION innodb_ft_user_stopword_table = CONCAT(DATABASE(), '/stopwords');
ALTER TABLE `items_strings` DROP INDEX `fullTextTitle`;
CREATE FULLTEXT INDEX `fullTextTitle` ON `items_strings`(`title`);
SET SESSION innodb_ft_user_stopword_table = @old_stopword_table;

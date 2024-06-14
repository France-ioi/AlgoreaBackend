-- +migrate Up

CREATE FULLTEXT INDEX `fullTextName` ON `groups`(`name`);
CREATE FULLTEXT INDEX `fullTextTitle` ON `items_strings`(`title`);

-- +migrate Down

ALTER TABLE `groups` DROP INDEX `fullTextName`;
ALTER TABLE `items_strings` DROP INDEX `fullTextTitle`;

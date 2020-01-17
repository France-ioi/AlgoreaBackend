-- +migrate Up
ALTER TABLE `languages`
    DROP INDEX `code`,
    CHANGE COLUMN `code` `tag` VARCHAR(6) NOT NULL COMMENT 'Language tag as defined in RFC5646' FIRST,
    COMMENT 'Languages supported for content';

ALTER TABLE `items`
    ADD COLUMN `default_language_tag` VARCHAR(6) DEFAULT NULL
        COMMENT 'Default language tag of this task (the reference, used when comparing translations)'
        AFTER `default_language_id`,
    ADD COLUMN `is_root` TINYINT(1) NOT NULL DEFAULT 0
        COMMENT 'Whether this item is intended to be a root chapter (in order to detect real orphans more easily)'
        AFTER `type`;

ALTER TABLE `items_strings`
    ADD COLUMN `language_tag` VARCHAR(6) DEFAULT NULL
        AFTER `language_id`;

UPDATE `items` JOIN `languages` ON `languages`.`id` = `items`.`default_language_id`
SET `items`.`default_language_tag` = `languages`.`tag`;

UPDATE `items_strings` JOIN `languages` ON `languages`.`id` = `items_strings`.`language_id`
SET `items_strings`.`language_tag` = `languages`.`tag`;

UPDATE `items_strings` SET `items_strings`.`language_tag` = 'fr' WHERE `language_id` = 0;

# 4 rows
DELETE `items_strings` FROM `items_strings` LEFT JOIN `items` ON `items`.`id` = `items_strings`.`item_id`
WHERE `items`.`id` IS NULL;

# 27 rows
DELETE `items` FROM `items`
    LEFT JOIN `items_strings` ON `items_strings`.`item_id` = `items`.`id`
WHERE `items_strings`.`item_id` IS NULL;

DELETE `items_items` FROM `items_items` LEFT JOIN `items` ON `items`.`id` = `items_items`.`child_item_id`
WHERE `items`.`id` IS NULL;

DELETE `items_items` FROM `items_items` LEFT JOIN `items` ON `items`.`id` = `items_items`.`parent_item_id`
WHERE `items`.`id` IS NULL;

UPDATE `items` SET `items`.`default_language_tag` = (
    SELECT `language_tag` FROM `items_strings`
    WHERE `items_strings`.`item_id` = `items`.`id` AND `items_strings`.`language_tag` IS NOT NULL
    ORDER BY `items_strings`.`language_tag` = 'fr' DESC, `items_strings`.`language_tag` = 'en' DESC
    LIMIT 1
) WHERE `items`.`default_language_tag` IS NULL;

UPDATE `items` SET `is_root` = 1, `type` = 'Chapter' WHERE `type` IN ('Root', 'Category');

DROP TRIGGER `before_insert_languages`;
ALTER TABLE `languages`
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (`tag`),
    DROP COLUMN `id`;

ALTER TABLE `items`
    MODIFY COLUMN `default_language_tag` VARCHAR(6) NOT NULL
        COMMENT 'Default language tag of this task (the reference, used when comparing translations)',
    DROP COLUMN `default_language_id`,
    ADD CONSTRAINT `fk_items_default_language_tag_languages_tag`
        FOREIGN KEY (`default_language_tag`) REFERENCES `languages`(`tag`),
    DROP COLUMN `display_children_as_tabs`,
    MODIFY COLUMN `type` ENUM('Chapter','Task','Course') NOT NULL;

DROP TRIGGER `before_insert_items_strings`;

# 5 rows
DELETE `items_strings` FROM `items_strings`
    JOIN (
        SELECT `id`, ROW_NUMBER() OVER (PARTITION BY `items_strings`.`item_id`, `items_strings`.`language_tag` ORDER BY `title` DESC) as number
        FROM `items_strings`
            JOIN (
                SELECT `item_id`, `language_tag`, COUNT(*) AS cnt
                FROM `items_strings`
                GROUP BY `item_id`, `language_tag`
                HAVING cnt > 1
            ) AS `duplicates` -- not unique (item_id, language_tag) pairs
                ON `duplicates`.`item_id` = `items_strings`.`item_id` AND
                   `duplicates`.`language_tag` = `items_strings`.`language_tag`
    ) AS `duplicated_rows` -- ids of duplicated row with row numbers within (item_id, language_tag) group, ordered by title DESC
        ON `duplicated_rows`.`id` = `items_strings`.`id` AND `duplicated_rows`.`number` > 1; -- remove duplicates keeping first rows

ALTER TABLE `items_strings`
    MODIFY COLUMN `language_tag` VARCHAR(6) NOT NULL COMMENT 'Language tag of this content',
    DROP KEY `item_id_language_id`,
    DROP COLUMN `language_id`,
    DROP PRIMARY KEY,
    DROP COLUMN `id`,
    ADD PRIMARY KEY (`item_id`, `language_tag`),
    ADD CONSTRAINT `fk_items_strings_item_id_items_id`
        FOREIGN KEY (`item_id`) REFERENCES `items`(`id`),
    ADD CONSTRAINT `fk_items_strings_language_tag_languages_tag`
        FOREIGN KEY (`language_tag`) REFERENCES `languages`(`tag`),
    DROP COLUMN `ranking_comment`;

-- +migrate Down
ALTER TABLE `languages`
    ADD COLUMN `id` BIGINT(20) FIRST,
    COMMENT '';

-- +migrate StatementBegin
CREATE TRIGGER `before_insert_languages` BEFORE INSERT ON `languages` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd

UPDATE `languages` SET `id` = CASE `tag`
    WHEN 'fr' THEN 1
    WHEN 'en' THEN 2
    WHEN 'sl' THEN 3
    ELSE FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000
END;

ALTER TABLE `items`
    ADD COLUMN `default_language_id` bigint(20) DEFAULT '1'
        COMMENT 'Default language of this task (the reference, used when comparing translations)'
        AFTER `default_language_tag`,
    MODIFY COLUMN `type` ENUM('Root','Category','Chapter','Task','Course') NOT NULL;

UPDATE `items` SET `type` = 'Root' WHERE `is_root`;

UPDATE `items` JOIN `languages` ON `languages`.`tag`
SET `default_language_id` = `languages`.`id`;

ALTER TABLE `items`
    DROP FOREIGN KEY `fk_items_default_language_tag_languages_tag`,
    DROP COLUMN `default_language_tag`,
    DROP COLUMN `is_root`,
    ADD COLUMN `display_children_as_tabs` TINYINT(3) UNSIGNED NOT NULL DEFAULT '0'
        COMMENT 'If true, display the children of the item (a chapter) as tabs, instead of as a list of items.'
        AFTER `display_details_in_parent`;

ALTER TABLE `items_strings`
    ADD COLUMN `id` BIGINT(20) DEFAULT NULL FIRST,
    ADD COLUMN `language_id` BIGINT(20) DEFAULT NULL AFTER `item_id`;

UPDATE `items_strings`
    LEFT JOIN `languages` ON `languages`.`tag` = `items_strings`.`language_tag`
SET `items_strings`.`id` = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000,
    `items_strings`.`language_id` = `languages`.`id`;

-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items_strings` BEFORE INSERT ON `items_strings` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd

ALTER TABLE `items_strings`
    MODIFY COLUMN `id` BIGINT(20) NOT NULL,
    MODIFY COLUMN `language_id` BIGINT(20) NOT NULL,
    DROP FOREIGN KEY `fk_items_strings_item_id_items_id`,
    DROP FOREIGN KEY `fk_items_strings_language_tag_languages_tag`,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (`id`),
    ADD UNIQUE KEY `item_id_language_id` (`item_id`,`language_id`),
    DROP `language_tag`,
    ADD COLUMN `ranking_comment` TEXT AFTER `edu_comment`;

ALTER TABLE `languages`
    CHANGE COLUMN `tag` `code` VARCHAR(2) NOT NULL DEFAULT '' COMMENT '' AFTER `name`,
    ADD INDEX `code` (`code`),
    DROP PRIMARY KEY,
    MODIFY COLUMN `id` BIGINT(20) NOT NULL,
    ADD PRIMARY KEY (`id`);

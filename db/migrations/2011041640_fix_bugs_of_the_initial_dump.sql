-- +migrate Up
UPDATE `groups` SET `type` = 'Other' WHERE `type` = '';
UPDATE `items` SET `validation_type` = 'All' WHERE `validation_type` = '';
UPDATE `items` SET `type` = 'Chapter' WHERE `id` = 4028;

SET @old_fk_checks = @@SESSION.FOREIGN_KEY_CHECKS;
SET FOREIGN_KEY_CHECKS = 0;
DELETE items_strings FROM items JOIN items_strings ON items_strings.item_id = items.id WHERE items.type = '';
SET FOREIGN_KEY_CHECKS = @old_fk_checks;

DELETE FROM `items` WHERE `type` = '';
UPDATE `items_items` SET `category` = 'Undefined' WHERE `category` = '';

-- +migrate Down

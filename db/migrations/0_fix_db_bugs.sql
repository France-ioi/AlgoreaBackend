-- +migrate Up
ALTER TABLE `history_groups_attempts` MODIFY COLUMN `bDeleted` tinyint(1) NOT NULL DEFAULT '0';

SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';

UPDATE `groups` SET `sDateCreated` = NULL WHERE `sDateCreated` = '0000-00-00 00:00:00';
UPDATE `history_groups` SET `sDateCreated` = NULL WHERE `sDateCreated` = '0000-00-00 00:00:00';
UPDATE `history_groups` SET `lockUserDeletionDate` = NULL WHERE `lockUserDeletionDate` ='0000-00-00';
UPDATE `groups_attempts` SET `sStartDate` = NULL WHERE `sStartDate` = '0000-00-00 00:00:00';
UPDATE `items` SET `sAccessOpenDate` = NULL WHERE `sAccessOpenDate` = '0000-00-00 00:00:00';
UPDATE `items` SET `sEndContestDate` = NULL WHERE `sEndContestDate` = '0000-00-00 00:00:00';
UPDATE `users` SET `sLastLoginDate` = NULL WHERE `sLastLoginDate` = '0000-00-00 00:00:00';

SET sql_mode              = @saved_sql_mode;

# 1 row (4029)
UPDATE `items` JOIN `items_items` ON `items_items`.`idItemParent` = `items`.`id`
SET `items`.`sType` = 'Chapter'
WHERE items.sType = 'Course';

# 2 rows (with items.ID = 1869634473184123671)
UPDATE `users_items` JOIN `items` ON `items`.`ID` = `users_items`.`idItem` AND `items`.`sType` = 'Course'
SET `users_items`.`iScore` = 0
WHERE `users_items`.`iScore` > 0;

# 2 rows
DELETE `items_items` FROM `items_items`
    JOIN `items` on `items`.`ID` = `items_items`.`idItemParent` AND `items`.`sType` = 'Task';

# 2 rows
DELETE `items_ancestors` FROM `items_ancestors`
    JOIN `items` on `items`.`ID` = `items_ancestors`.`idItemAncestor` AND `items`.`sType` = 'Task';

-- +migrate Down
ALTER TABLE `history_groups_attempts` MODIFY COLUMN `bDeleted` tinyint(1) NOT NULL;

UPDATE `items` SET `sType` = 'Course' WHERE `ID` = 4029;

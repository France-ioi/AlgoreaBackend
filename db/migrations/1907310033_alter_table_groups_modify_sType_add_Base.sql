-- +migrate Up
ALTER TABLE `history_groups` MODIFY COLUMN `sType` enum('Root','Class','Team','Club','Friends','Other','UserSelf','UserAdmin','RootSelf','RootAdmin', 'Base') NOT NULL;
UPDATE `history_groups` SET `sType` = 'Base' WHERE `sType` IN ('Root', 'RootSelf', 'RootAdmin');
ALTER TABLE `history_groups` MODIFY COLUMN `sType` enum('Class','Team','Club','Friends','Other','UserSelf','UserAdmin','Base') NOT NULL;

ALTER TABLE `groups` MODIFY COLUMN `sType` enum('Root','Class','Team','Club','Friends','Other','UserSelf','UserAdmin','RootSelf','RootAdmin', 'Base') NOT NULL;
UPDATE `groups` SET `sType` = 'Base' WHERE `sType` IN ('Root', 'RootSelf', 'RootAdmin');
UPDATE `groups` SET `sType` = 'Base' WHERE `sType` = 'UserSelf' AND `sName` = 'RootTemp' AND `sTextId` = 'RootTemp';
ALTER TABLE `groups` MODIFY COLUMN `sType` enum('Class','Team','Club','Friends','Other','UserSelf','UserAdmin','Base') NOT NULL;

-- +migrate Down
ALTER TABLE `groups` MODIFY COLUMN `sType` enum('Root','Class','Team','Club','Friends','Other','UserSelf','UserAdmin','RootSelf','RootAdmin', 'Base') NOT NULL;
ALTER TABLE `history_groups` MODIFY COLUMN `sType` enum('Root','Class','Team','Club','Friends','Other','UserSelf','UserAdmin','RootSelf','RootAdmin', 'Base') NOT NULL;

UPDATE `groups` SET `sType` = 'UserSelf' WHERE `sType` = 'Base' AND `sName` = 'RootTemp' AND `sTextId` = 'RootTemp';
UPDATE `groups` SET `sType` = 'Root' WHERE `sType` = 'Base' AND `sName` = 'Root' AND `sTextId` = 'Root';
UPDATE `groups` SET `sType` = 'RootSelf' WHERE `sType` = 'Base' AND `sName` = 'RootSelf' AND `sTextId` = 'RootSelf';
UPDATE `groups` SET `sType` = 'RootAdmin' WHERE `sType` = 'Base' AND `sName` = 'RootAdmin' AND `sTextId` = 'RootAdmin';
UPDATE `groups` SET `sType` = 'Root' WHERE `sType` = 'Base'; -- some RootSelf groups will become Root groups (still better than nothing)

DELETE FROM `history_groups` WHERE `sType` = 'Base';

ALTER TABLE `groups` MODIFY COLUMN `sType` enum('Root','Class','Team','Club','Friends','Other','UserSelf','UserAdmin','RootSelf','RootAdmin') NOT NULL;
ALTER TABLE `history_groups` MODIFY COLUMN `sType` enum('Root','Class','Team','Club','Friends','Other','UserSelf','UserAdmin','RootSelf','RootAdmin') NOT NULL;

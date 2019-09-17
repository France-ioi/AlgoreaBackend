-- +migrate Up
ALTER TABLE `history_groups` MODIFY COLUMN `sType` enum('Root','Class','Team','Club','Friends','Other','UserSelf','UserAdmin','RootSelf','RootAdmin', 'Base') NOT NULL;
UPDATE `history_groups` SET `sType` = 'Base' WHERE `history_groups`.`sType` IN ('Root', 'RootSelf', 'RootAdmin');
UPDATE `history_groups` JOIN `groups` USING (`ID`) SET `history_groups`.`sType` = 'Base' WHERE `history_groups`.`sType` = 'UserSelf' AND `history_groups`.`sName` = 'RootTemp' AND `groups`.`sTextId` = 'RootTemp';
ALTER TABLE `history_groups` MODIFY COLUMN `sType` enum('Class','Team','Club','Friends','Other','UserSelf','UserAdmin','Base') NOT NULL;

-- +migrate Down
ALTER TABLE `history_groups` MODIFY COLUMN `sType` enum('Root','Class','Team','Club','Friends','Other','UserSelf','UserAdmin','RootSelf','RootAdmin', 'Base') NOT NULL;
UPDATE `history_groups` JOIN `groups` USING (`ID`) SET `history_groups`.`sType` = 'UserSelf' WHERE `history_groups`.`sType` = 'Base' AND `history_groups`.`sName` = 'RootTemp' AND `groups`.`sTextId` = 'RootTemp';
UPDATE `history_groups` JOIN `groups` USING (`ID`) SET `history_groups`.`sType` = 'Root' WHERE `history_groups`.`sType` = 'Base' AND `history_groups`.`sName` = 'Root' AND `groups`.`sTextId` = 'Root';
UPDATE `history_groups` JOIN `groups` USING (`ID`) SET `history_groups`.`sType` = 'RootSelf' WHERE `history_groups`.`sType` = 'Base' AND `history_groups`.`sName` = 'RootSelf' AND `groups`.`sTextId` = 'RootSelf';
UPDATE `history_groups` JOIN `groups` USING (`ID`) SET `history_groups`.`sType` = 'RootAdmin' WHERE `history_groups`.`sType` = 'Base' AND `history_groups`.`sName` = 'RootAdmin' AND `groups`.`sTextId` = 'RootAdmin';
ALTER TABLE `history_groups` MODIFY COLUMN `sType` enum('Root','Class','Team','Club','Friends','Other','UserSelf','UserAdmin','RootSelf','RootAdmin') NOT NULL;

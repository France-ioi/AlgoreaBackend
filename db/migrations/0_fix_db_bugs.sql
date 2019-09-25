-- +migrate Up
ALTER TABLE `history_groups_attempts` MODIFY COLUMN `bDeleted` tinyint(1) NOT NULL DEFAULT '0';

/*!50003 SET @saved_sql_mode       = @@sql_mode */ ;
/*!50003 SET sql_mode              = 'NO_ENGINE_SUBSTITUTION' */ ;

UPDATE `groups` SET `sDateCreated` = NULL WHERE `sDateCreated` = '0000-00-00 00:00:00';
UPDATE `history_groups` SET `sDateCreated` = NULL WHERE `sDateCreated` = '0000-00-00 00:00:00';
UPDATE `history_groups` SET `lockUserDeletionDate` = NULL WHERE `lockUserDeletionDate` ='0000-00-00';
UPDATE `groups_attempts` SET `sStartDate` = NULL WHERE `sStartDate` = '0000-00-00 00:00:00';
UPDATE `items` SET `sAccessOpenDate` = NULL WHERE `sAccessOpenDate` = '0000-00-00 00:00:00';
UPDATE `items` SET `sEndContestDate` = NULL WHERE `sEndContestDate` = '0000-00-00 00:00:00';
UPDATE `users` SET `sLastLoginDate` = NULL WHERE `sLastLoginDate` = '0000-00-00 00:00:00';

/*!50003 SET sql_mode              = @saved_sql_mode */ ;

-- +migrate Down
ALTER TABLE `history_groups_attempts` MODIFY COLUMN `bDeleted` tinyint(1) NOT NULL;

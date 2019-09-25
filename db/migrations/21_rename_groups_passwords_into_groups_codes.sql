-- +migrate Up
ALTER TABLE `groups` CHANGE COLUMN `sPassword` `sCode` varchar(50) DEFAULT NULL;
ALTER TABLE `groups` CHANGE COLUMN `sPasswordTimer` `sCodeTimer` time DEFAULT NULL;
ALTER TABLE `groups` CHANGE COLUMN `sPasswordEnd` `sCodeEnd` datetime DEFAULT NULL;
ALTER TABLE `history_groups` CHANGE COLUMN `sPassword` `sCode` varchar(50) DEFAULT NULL;
ALTER TABLE `history_groups` CHANGE COLUMN `sPasswordTimer` `sCodeTimer` time DEFAULT NULL;
ALTER TABLE `history_groups` CHANGE COLUMN `sPasswordEnd` `sCodeEnd` datetime DEFAULT NULL;

DROP TRIGGER `after_insert_groups`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50003 TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN INSERT INTO `history_groups` (`ID`,`iVersion`,`sName`,`iGrade`,`sGradeDetails`,`sDescription`,`sDateCreated`,`bOpened`,`bFreeAccess`,`idTeamItem`,`iTeamParticipating`,`sCode`,`sCodeTimer`,`sCodeEnd`,`sRedirectPath`,`bOpenContest`,`sType`,`bSendEmails`) VALUES (NEW.`ID`,@curVersion,NEW.`sName`,NEW.`iGrade`,NEW.`sGradeDetails`,NEW.`sDescription`,NEW.`sDateCreated`,NEW.`bOpened`,NEW.`bFreeAccess`,NEW.`idTeamItem`,NEW.`iTeamParticipating`,NEW.`sCode`,NEW.`sCodeTimer`,NEW.`sCodeEnd`,NEW.`sRedirectPath`,NEW.`bOpenContest`,NEW.`sType`,NEW.`bSendEmails`); INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (NEW.`ID`, 'todo') ; END */;;
-- +migrate StatementEnd

DROP TRIGGER `before_update_groups`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50003 TRIGGER `before_update_groups` BEFORE UPDATE ON `groups` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`sName` <=> NEW.`sName` AND OLD.`iGrade` <=> NEW.`iGrade` AND OLD.`sGradeDetails` <=> NEW.`sGradeDetails` AND OLD.`sDescription` <=> NEW.`sDescription` AND OLD.`sDateCreated` <=> NEW.`sDateCreated` AND OLD.`bOpened` <=> NEW.`bOpened` AND OLD.`bFreeAccess` <=> NEW.`bFreeAccess` AND OLD.`idTeamItem` <=> NEW.`idTeamItem` AND OLD.`iTeamParticipating` <=> NEW.`iTeamParticipating` AND OLD.`sCode` <=> NEW.`sCode` AND OLD.`sCodeTimer` <=> NEW.`sCodeTimer` AND OLD.`sCodeEnd` <=> NEW.`sCodeEnd` AND OLD.`sRedirectPath` <=> NEW.`sRedirectPath` AND OLD.`bOpenContest` <=> NEW.`bOpenContest` AND OLD.`sType` <=> NEW.`sType` AND OLD.`bSendEmails` <=> NEW.`bSendEmails`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_groups` (`ID`,`iVersion`,`sName`,`iGrade`,`sGradeDetails`,`sDescription`,`sDateCreated`,`bOpened`,`bFreeAccess`,`idTeamItem`,`iTeamParticipating`,`sCode`,`sCodeTimer`,`sCodeEnd`,`sRedirectPath`,`bOpenContest`,`sType`,`bSendEmails`)       VALUES (NEW.`ID`,@curVersion,NEW.`sName`,NEW.`iGrade`,NEW.`sGradeDetails`,NEW.`sDescription`,NEW.`sDateCreated`,NEW.`bOpened`,NEW.`bFreeAccess`,NEW.`idTeamItem`,NEW.`iTeamParticipating`,NEW.`sCode`,NEW.`sCodeTimer`,NEW.`sCodeEnd`,NEW.`sRedirectPath`,NEW.`bOpenContest`,NEW.`sType`,NEW.`bSendEmails`) ; END IF; END */;;
-- +migrate StatementEnd

DROP TRIGGER `before_delete_groups`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50003 TRIGGER `before_delete_groups` BEFORE DELETE ON `groups` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_groups` (`ID`,`iVersion`,`sName`,`iGrade`,`sGradeDetails`,`sDescription`,`sDateCreated`,`bOpened`,`bFreeAccess`,`idTeamItem`,`iTeamParticipating`,`sCode`,`sCodeTimer`,`sCodeEnd`,`sRedirectPath`,`bOpenContest`,`sType`,`bSendEmails`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`sName`,OLD.`iGrade`,OLD.`sGradeDetails`,OLD.`sDescription`,OLD.`sDateCreated`,OLD.`bOpened`,OLD.`bFreeAccess`,OLD.`idTeamItem`,OLD.`iTeamParticipating`,OLD.`sCode`,OLD.`sCodeTimer`,OLD.`sCodeEnd`,OLD.`sRedirectPath`,OLD.`bOpenContest`,OLD.`sType`,OLD.`bSendEmails`, 1); END */;;
-- +migrate StatementEnd



-- +migrate Down
ALTER TABLE `groups` CHANGE COLUMN `sCode` `sPassword` varchar(50) DEFAULT NULL;
ALTER TABLE `groups` CHANGE COLUMN `sCodeTimer` `sPasswordTimer` time DEFAULT NULL;
ALTER TABLE `groups` CHANGE COLUMN `sCodeEnd` `sPasswordEnd` datetime DEFAULT NULL;
ALTER TABLE `history_groups` CHANGE COLUMN `sCode` `sPassword` varchar(50) DEFAULT NULL;
ALTER TABLE `history_groups` CHANGE COLUMN `sCodeTimer` `sPasswordTimer` time DEFAULT NULL;
ALTER TABLE `history_groups` CHANGE COLUMN `sCodeEnd` `sPasswordEnd` datetime DEFAULT NULL;

DROP TRIGGER IF EXISTS `after_insert_groups`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50003 TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN INSERT INTO `history_groups` (`ID`,`iVersion`,`sName`,`iGrade`,`sGradeDetails`,`sDescription`,`sDateCreated`,`bOpened`,`bFreeAccess`,`idTeamItem`,`iTeamParticipating`,`sPassword`,`sPasswordTimer`,`sPasswordEnd`,`sRedirectPath`,`bOpenContest`,`sType`,`bSendEmails`) VALUES (NEW.`ID`,@curVersion,NEW.`sName`,NEW.`iGrade`,NEW.`sGradeDetails`,NEW.`sDescription`,NEW.`sDateCreated`,NEW.`bOpened`,NEW.`bFreeAccess`,NEW.`idTeamItem`,NEW.`iTeamParticipating`,NEW.`sPassword`,NEW.`sPasswordTimer`,NEW.`sPasswordEnd`,NEW.`sRedirectPath`,NEW.`bOpenContest`,NEW.`sType`,NEW.`bSendEmails`); INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (NEW.`ID`, 'todo') ; END */;;
-- +migrate StatementEnd

DROP TRIGGER IF EXISTS `before_update_groups`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50003 TRIGGER `before_update_groups` BEFORE UPDATE ON `groups` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`sName` <=> NEW.`sName` AND OLD.`iGrade` <=> NEW.`iGrade` AND OLD.`sGradeDetails` <=> NEW.`sGradeDetails` AND OLD.`sDescription` <=> NEW.`sDescription` AND OLD.`sDateCreated` <=> NEW.`sDateCreated` AND OLD.`bOpened` <=> NEW.`bOpened` AND OLD.`bFreeAccess` <=> NEW.`bFreeAccess` AND OLD.`idTeamItem` <=> NEW.`idTeamItem` AND OLD.`iTeamParticipating` <=> NEW.`iTeamParticipating` AND OLD.`sPassword` <=> NEW.`sPassword` AND OLD.`sPasswordTimer` <=> NEW.`sPasswordTimer` AND OLD.`sPasswordEnd` <=> NEW.`sPasswordEnd` AND OLD.`sRedirectPath` <=> NEW.`sRedirectPath` AND OLD.`bOpenContest` <=> NEW.`bOpenContest` AND OLD.`sType` <=> NEW.`sType` AND OLD.`bSendEmails` <=> NEW.`bSendEmails`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_groups` (`ID`,`iVersion`,`sName`,`iGrade`,`sGradeDetails`,`sDescription`,`sDateCreated`,`bOpened`,`bFreeAccess`,`idTeamItem`,`iTeamParticipating`,`sPassword`,`sPasswordTimer`,`sPasswordEnd`,`sRedirectPath`,`bOpenContest`,`sType`,`bSendEmails`)       VALUES (NEW.`ID`,@curVersion,NEW.`sName`,NEW.`iGrade`,NEW.`sGradeDetails`,NEW.`sDescription`,NEW.`sDateCreated`,NEW.`bOpened`,NEW.`bFreeAccess`,NEW.`idTeamItem`,NEW.`iTeamParticipating`,NEW.`sPassword`,NEW.`sPasswordTimer`,NEW.`sPasswordEnd`,NEW.`sRedirectPath`,NEW.`bOpenContest`,NEW.`sType`,NEW.`bSendEmails`) ; END IF; END */;;
-- +migrate StatementEnd

DROP TRIGGER IF EXISTS `before_delete_groups`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50003 TRIGGER `before_delete_groups` BEFORE DELETE ON `groups` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_groups` (`ID`,`iVersion`,`sName`,`iGrade`,`sGradeDetails`,`sDescription`,`sDateCreated`,`bOpened`,`bFreeAccess`,`idTeamItem`,`iTeamParticipating`,`sPassword`,`sPasswordTimer`,`sPasswordEnd`,`sRedirectPath`,`bOpenContest`,`sType`,`bSendEmails`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`sName`,OLD.`iGrade`,OLD.`sGradeDetails`,OLD.`sDescription`,OLD.`sDateCreated`,OLD.`bOpened`,OLD.`bFreeAccess`,OLD.`idTeamItem`,OLD.`iTeamParticipating`,OLD.`sPassword`,OLD.`sPasswordTimer`,OLD.`sPasswordEnd`,OLD.`sRedirectPath`,OLD.`bOpenContest`,OLD.`sType`,OLD.`bSendEmails`, 1); END */;;
-- +migrate StatementEnd

-- MySQL dump 10.13  Distrib 8.0.13, for osx10.12 (x86_64)
--
-- Host: 127.0.0.1    Database: algorea_db
-- ------------------------------------------------------
-- Server version	8.0.17

SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT;
SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS;
SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION;
 SET NAMES utf8mb4;
SET @OLD_TIME_ZONE=@@TIME_ZONE;
SET TIME_ZONE='+00:00';
SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0;
SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0;
SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO';
SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0;

--
-- Table structure for table `badges`
--

DROP TABLE IF EXISTS `badges`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `badges` (
  `ID` bigint(20) NOT NULL AUTO_INCREMENT,
  `idUser` bigint(20) NOT NULL,
  `name` text,
  `code` text NOT NULL,
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `badges`
--

LOCK TABLES `badges` WRITE;
ALTER TABLE `badges` DISABLE KEYS;
ALTER TABLE `badges` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `error_log`
--

DROP TABLE IF EXISTS `error_log`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `error_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `url` text CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL,
  `browser` text CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL,
  `details` text CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `error_log`
--

LOCK TABLES `error_log` WRITE;
ALTER TABLE `error_log` DISABLE KEYS;
ALTER TABLE `error_log` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `filters`
--

DROP TABLE IF EXISTS `filters`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `filters` (
  `ID` bigint(20) NOT NULL,
  `idUser` bigint(20) NOT NULL,
  `sName` varchar(45) NOT NULL DEFAULT '',
  `bSelected` tinyint(1) NOT NULL DEFAULT '0',
  `bStarred` tinyint(1) DEFAULT NULL,
  `sStartDate` datetime DEFAULT NULL,
  `sEndDate` datetime DEFAULT NULL,
  `bArchived` tinyint(1) DEFAULT NULL,
  `bParticipated` tinyint(1) DEFAULT NULL,
  `bUnread` tinyint(1) DEFAULT NULL,
  `idItem` bigint(20) DEFAULT NULL,
  `idGroup` int(11) DEFAULT NULL,
  `olderThan` int(11) DEFAULT NULL,
  `newerThan` int(11) DEFAULT NULL,
  `sUsersSearch` varchar(200) DEFAULT NULL,
  `sBodySearch` varchar(100) DEFAULT NULL,
  `bImportant` tinyint(1) DEFAULT NULL,
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `user_idx` (`idUser`),
  KEY `iVersion` (`iVersion`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `filters`
--

LOCK TABLES `filters` WRITE;
ALTER TABLE `filters` DISABLE KEYS;
ALTER TABLE `filters` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_filters` BEFORE INSERT ON `filters` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_filters` AFTER INSERT ON `filters` FOR EACH ROW BEGIN INSERT INTO `history_filters` (`ID`,`iVersion`,`idUser`,`sName`,`bSelected`,`bStarred`,`sStartDate`,`sEndDate`,`bArchived`,`bParticipated`,`bUnread`,`idItem`,`idGroup`,`olderThan`,`newerThan`,`sUsersSearch`,`sBodySearch`,`bImportant`) VALUES (NEW.`ID`,@curVersion,NEW.`idUser`,NEW.`sName`,NEW.`bSelected`,NEW.`bStarred`,NEW.`sStartDate`,NEW.`sEndDate`,NEW.`bArchived`,NEW.`bParticipated`,NEW.`bUnread`,NEW.`idItem`,NEW.`idGroup`,NEW.`olderThan`,NEW.`newerThan`,NEW.`sUsersSearch`,NEW.`sBodySearch`,NEW.`bImportant`); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_filters` BEFORE UPDATE ON `filters` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idUser` <=> NEW.`idUser` AND OLD.`sName` <=> NEW.`sName` AND OLD.`bStarred` <=> NEW.`bStarred` AND OLD.`sStartDate` <=> NEW.`sStartDate` AND OLD.`sEndDate` <=> NEW.`sEndDate` AND OLD.`bArchived` <=> NEW.`bArchived` AND OLD.`bParticipated` <=> NEW.`bParticipated` AND OLD.`bUnread` <=> NEW.`bUnread` AND OLD.`idItem` <=> NEW.`idItem` AND OLD.`idGroup` <=> NEW.`idGroup` AND OLD.`olderThan` <=> NEW.`olderThan` AND OLD.`newerThan` <=> NEW.`newerThan` AND OLD.`sUsersSearch` <=> NEW.`sUsersSearch` AND OLD.`sBodySearch` <=> NEW.`sBodySearch` AND OLD.`bImportant` <=> NEW.`bImportant`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_filters` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_filters` (`ID`,`iVersion`,`idUser`,`sName`,`bSelected`,`bStarred`,`sStartDate`,`sEndDate`,`bArchived`,`bParticipated`,`bUnread`,`idItem`,`idGroup`,`olderThan`,`newerThan`,`sUsersSearch`,`sBodySearch`,`bImportant`)       VALUES (NEW.`ID`,@curVersion,NEW.`idUser`,NEW.`sName`,NEW.`bSelected`,NEW.`bStarred`,NEW.`sStartDate`,NEW.`sEndDate`,NEW.`bArchived`,NEW.`bParticipated`,NEW.`bUnread`,NEW.`idItem`,NEW.`idGroup`,NEW.`olderThan`,NEW.`newerThan`,NEW.`sUsersSearch`,NEW.`sBodySearch`,NEW.`bImportant`) ; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_filters` BEFORE DELETE ON `filters` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_filters` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_filters` (`ID`,`iVersion`,`idUser`,`sName`,`bSelected`,`bStarred`,`sStartDate`,`sEndDate`,`bArchived`,`bParticipated`,`bUnread`,`idItem`,`idGroup`,`olderThan`,`newerThan`,`sUsersSearch`,`sBodySearch`,`bImportant`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idUser`,OLD.`sName`,OLD.`bSelected`,OLD.`bStarred`,OLD.`sStartDate`,OLD.`sEndDate`,OLD.`bArchived`,OLD.`bParticipated`,OLD.`bUnread`,OLD.`idItem`,OLD.`idGroup`,OLD.`olderThan`,OLD.`newerThan`,OLD.`sUsersSearch`,OLD.`sBodySearch`,OLD.`bImportant`, 1); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `gorp_migrations`
--

DROP TABLE IF EXISTS `gorp_migrations`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `gorp_migrations` (
  `id` varchar(255) NOT NULL,
  `applied_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `gorp_migrations`
--

LOCK TABLES `gorp_migrations` WRITE;
ALTER TABLE `gorp_migrations` DISABLE KEYS;
ALTER TABLE `gorp_migrations` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `groups`
--

DROP TABLE IF EXISTS `groups`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `groups` (
  `ID` bigint(20) NOT NULL,
  `sName` varchar(200) NOT NULL DEFAULT '',
  `sTextId` varchar(255) NOT NULL DEFAULT '',
  `iGrade` int(4) NOT NULL DEFAULT '-2',
  `sGradeDetails` varchar(50) DEFAULT NULL,
  `sDescription` text,
  `sDateCreated` datetime DEFAULT NULL,
  `bOpened` tinyint(1) NOT NULL DEFAULT '0',
  `bFreeAccess` tinyint(1) NOT NULL DEFAULT '0',
  `idTeamItem` bigint(20) DEFAULT NULL,
  `iTeamParticipating` tinyint(1) NOT NULL DEFAULT '0',
  `sPassword` varchar(50) DEFAULT NULL,
  `sPasswordTimer` time DEFAULT NULL,
  `sPasswordEnd` datetime DEFAULT NULL,
  `sRedirectPath` text,
  `bOpenContest` tinyint(1) NOT NULL DEFAULT '0',
  `sType` enum('Root','Class','Team','Club','Friends','Other','UserSelf','UserAdmin','RootSelf','RootAdmin') NOT NULL,
  `bSendEmails` tinyint(1) NOT NULL DEFAULT '0',
  `sAncestorsComputationState` enum('done','processing','todo') NOT NULL DEFAULT 'todo',
  `iVersion` bigint(20) NOT NULL,
  `lockUserDeletionDate` date DEFAULT NULL,
  PRIMARY KEY (`ID`),
  KEY `iVersion` (`iVersion`),
  KEY `bAncestorsComputed` (`sAncestorsComputationState`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `groups`
--

LOCK TABLES `groups` WRITE;
ALTER TABLE `groups` DISABLE KEYS;
ALTER TABLE `groups` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_groups` BEFORE INSERT ON `groups` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN INSERT INTO `history_groups` (`ID`,`iVersion`,`sName`,`iGrade`,`sGradeDetails`,`sDescription`,`sDateCreated`,`bOpened`,`bFreeAccess`,`idTeamItem`,`iTeamParticipating`,`sPassword`,`sPasswordTimer`,`sPasswordEnd`,`sRedirectPath`,`bOpenContest`,`sType`,`bSendEmails`) VALUES (NEW.`ID`,@curVersion,NEW.`sName`,NEW.`iGrade`,NEW.`sGradeDetails`,NEW.`sDescription`,NEW.`sDateCreated`,NEW.`bOpened`,NEW.`bFreeAccess`,NEW.`idTeamItem`,NEW.`iTeamParticipating`,NEW.`sPassword`,NEW.`sPasswordTimer`,NEW.`sPasswordEnd`,NEW.`sRedirectPath`,NEW.`bOpenContest`,NEW.`sType`,NEW.`bSendEmails`); INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (NEW.`ID`, 'todo') ; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_groups` BEFORE UPDATE ON `groups` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`sName` <=> NEW.`sName` AND OLD.`iGrade` <=> NEW.`iGrade` AND OLD.`sGradeDetails` <=> NEW.`sGradeDetails` AND OLD.`sDescription` <=> NEW.`sDescription` AND OLD.`sDateCreated` <=> NEW.`sDateCreated` AND OLD.`bOpened` <=> NEW.`bOpened` AND OLD.`bFreeAccess` <=> NEW.`bFreeAccess` AND OLD.`idTeamItem` <=> NEW.`idTeamItem` AND OLD.`iTeamParticipating` <=> NEW.`iTeamParticipating` AND OLD.`sPassword` <=> NEW.`sPassword` AND OLD.`sPasswordTimer` <=> NEW.`sPasswordTimer` AND OLD.`sPasswordEnd` <=> NEW.`sPasswordEnd` AND OLD.`sRedirectPath` <=> NEW.`sRedirectPath` AND OLD.`bOpenContest` <=> NEW.`bOpenContest` AND OLD.`sType` <=> NEW.`sType` AND OLD.`bSendEmails` <=> NEW.`bSendEmails`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_groups` (`ID`,`iVersion`,`sName`,`iGrade`,`sGradeDetails`,`sDescription`,`sDateCreated`,`bOpened`,`bFreeAccess`,`idTeamItem`,`iTeamParticipating`,`sPassword`,`sPasswordTimer`,`sPasswordEnd`,`sRedirectPath`,`bOpenContest`,`sType`,`bSendEmails`)       VALUES (NEW.`ID`,@curVersion,NEW.`sName`,NEW.`iGrade`,NEW.`sGradeDetails`,NEW.`sDescription`,NEW.`sDateCreated`,NEW.`bOpened`,NEW.`bFreeAccess`,NEW.`idTeamItem`,NEW.`iTeamParticipating`,NEW.`sPassword`,NEW.`sPasswordTimer`,NEW.`sPasswordEnd`,NEW.`sRedirectPath`,NEW.`bOpenContest`,NEW.`sType`,NEW.`bSendEmails`) ; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_groups` BEFORE DELETE ON `groups` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_groups` (`ID`,`iVersion`,`sName`,`iGrade`,`sGradeDetails`,`sDescription`,`sDateCreated`,`bOpened`,`bFreeAccess`,`idTeamItem`,`iTeamParticipating`,`sPassword`,`sPasswordTimer`,`sPasswordEnd`,`sRedirectPath`,`bOpenContest`,`sType`,`bSendEmails`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`sName`,OLD.`iGrade`,OLD.`sGradeDetails`,OLD.`sDescription`,OLD.`sDateCreated`,OLD.`bOpened`,OLD.`bFreeAccess`,OLD.`idTeamItem`,OLD.`iTeamParticipating`,OLD.`sPassword`,OLD.`sPasswordTimer`,OLD.`sPasswordEnd`,OLD.`sRedirectPath`,OLD.`bOpenContest`,OLD.`sType`,OLD.`bSendEmails`, 1); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_delete_groups` AFTER DELETE ON `groups` FOR EACH ROW BEGIN DELETE FROM groups_propagate where ID = OLD.ID ; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `groups_ancestors`
--

DROP TABLE IF EXISTS `groups_ancestors`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `groups_ancestors` (
  `ID` bigint(20) NOT NULL AUTO_INCREMENT,
  `idGroupAncestor` bigint(20) NOT NULL,
  `idGroupChild` bigint(20) NOT NULL,
  `bIsSelf` tinyint(1) NOT NULL DEFAULT '0',
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `idGroupAncestor` (`idGroupAncestor`,`idGroupChild`),
  KEY `ancestor` (`idGroupAncestor`),
  KEY `descendant` (`idGroupChild`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `groups_ancestors`
--

LOCK TABLES `groups_ancestors` WRITE;
ALTER TABLE `groups_ancestors` DISABLE KEYS;
ALTER TABLE `groups_ancestors` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_groups_ancestors` BEFORE INSERT ON `groups_ancestors` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_groups_ancestors` AFTER INSERT ON `groups_ancestors` FOR EACH ROW BEGIN INSERT INTO `history_groups_ancestors` (`ID`,`iVersion`,`idGroupAncestor`,`idGroupChild`,`bIsSelf`) VALUES (NEW.`ID`,@curVersion,NEW.`idGroupAncestor`,NEW.`idGroupChild`,NEW.`bIsSelf`); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_groups_ancestors` BEFORE UPDATE ON `groups_ancestors` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idGroupAncestor` <=> NEW.`idGroupAncestor` AND OLD.`idGroupChild` <=> NEW.`idGroupChild` AND OLD.`bIsSelf` <=> NEW.`bIsSelf`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_groups_ancestors` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_groups_ancestors` (`ID`,`iVersion`,`idGroupAncestor`,`idGroupChild`,`bIsSelf`)       VALUES (NEW.`ID`,@curVersion,NEW.`idGroupAncestor`,NEW.`idGroupChild`,NEW.`bIsSelf`) ; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_groups_ancestors` BEFORE DELETE ON `groups_ancestors` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_ancestors` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_groups_ancestors` (`ID`,`iVersion`,`idGroupAncestor`,`idGroupChild`,`bIsSelf`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idGroupAncestor`,OLD.`idGroupChild`,OLD.`bIsSelf`, 1); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `groups_attempts`
--

DROP TABLE IF EXISTS `groups_attempts`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `groups_attempts` (
  `ID` bigint(20) NOT NULL AUTO_INCREMENT,
  `idGroup` bigint(20) NOT NULL,
  `idItem` bigint(20) NOT NULL,
  `idUserCreator` bigint(20) DEFAULT NULL,
  `iOrder` int(11) NOT NULL,
  `iScore` float NOT NULL DEFAULT '0',
  `iScoreComputed` float NOT NULL DEFAULT '0',
  `iScoreReeval` float DEFAULT '0',
  `iScoreDiffManual` float NOT NULL DEFAULT '0',
  `sScoreDiffComment` varchar(200) NOT NULL DEFAULT '',
  `nbSubmissionsAttempts` int(11) NOT NULL DEFAULT '0',
  `nbTasksTried` int(11) NOT NULL DEFAULT '0',
  `nbTasksSolved` int(11) NOT NULL DEFAULT '0',
  `nbChildrenValidated` int(11) NOT NULL DEFAULT '0',
  `bValidated` tinyint(1) NOT NULL DEFAULT '0',
  `bFinished` tinyint(1) NOT NULL DEFAULT '0',
  `bKeyObtained` tinyint(1) NOT NULL DEFAULT '0',
  `nbTasksWithHelp` int(11) NOT NULL DEFAULT '0',
  `sHintsRequested` mediumtext,
  `nbHintsCached` int(11) NOT NULL DEFAULT '0',
  `nbCorrectionsRead` int(11) NOT NULL DEFAULT '0',
  `iPrecision` int(11) NOT NULL DEFAULT '0',
  `iAutonomy` int(11) NOT NULL DEFAULT '0',
  `sStartDate` datetime DEFAULT NULL,
  `sValidationDate` datetime DEFAULT NULL,
  `sFinishDate` datetime DEFAULT NULL,
  `sLastActivityDate` datetime DEFAULT NULL,
  `sThreadStartDate` datetime DEFAULT NULL,
  `sBestAnswerDate` datetime DEFAULT NULL,
  `sLastAnswerDate` datetime DEFAULT NULL,
  `sLastHintDate` datetime DEFAULT NULL,
  `sAdditionalTime` datetime DEFAULT NULL,
  `sContestStartDate` datetime DEFAULT NULL,
  `bRanked` tinyint(1) NOT NULL DEFAULT '0',
  `sAllLangProg` varchar(200) DEFAULT NULL,
  `iVersion` bigint(20) NOT NULL,
  `sAncestorsComputationState` enum('done','processing','todo','temp') NOT NULL DEFAULT 'done',
  PRIMARY KEY (`ID`),
  KEY `iVersion` (`iVersion`),
  KEY `sAncestorsComputationState` (`sAncestorsComputationState`),
  KEY `idItem` (`idItem`),
  KEY `idGroup` (`idGroup`),
  KEY `GroupItem` (`idGroup`,`idItem`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `groups_attempts`
--

LOCK TABLES `groups_attempts` WRITE;
ALTER TABLE `groups_attempts` DISABLE KEYS;
ALTER TABLE `groups_attempts` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_groups_attempts` BEFORE INSERT ON `groups_attempts` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_groups_attempts` AFTER INSERT ON `groups_attempts` FOR EACH ROW BEGIN INSERT INTO `history_groups_attempts` (`ID`,`iVersion`,`idGroup`,`idItem`,`idUserCreator`,`iOrder`,`iScore`,`iScoreComputed`,`iScoreReeval`,`iScoreDiffManual`,`sScoreDiffComment`,`nbSubmissionsAttempts`,`nbTasksTried`,`nbChildrenValidated`,`bValidated`,`bFinished`,`bKeyObtained`,`nbTasksWithHelp`,`sHintsRequested`,`nbHintsCached`,`nbCorrectionsRead`,`iPrecision`,`iAutonomy`,`sStartDate`,`sValidationDate`,`sBestAnswerDate`,`sLastAnswerDate`,`sThreadStartDate`,`sLastHintDate`,`sFinishDate`,`sLastActivityDate`,`sContestStartDate`,`bRanked`,`sAllLangProg`) VALUES (NEW.`ID`,@curVersion,NEW.`idGroup`,NEW.`idItem`,NEW.`idUserCreator`,NEW.`iOrder`,NEW.`iScore`,NEW.`iScoreComputed`,NEW.`iScoreReeval`,NEW.`iScoreDiffManual`,NEW.`sScoreDiffComment`,NEW.`nbSubmissionsAttempts`,NEW.`nbTasksTried`,NEW.`nbChildrenValidated`,NEW.`bValidated`,NEW.`bFinished`,NEW.`bKeyObtained`,NEW.`nbTasksWithHelp`,NEW.`sHintsRequested`,NEW.`nbHintsCached`,NEW.`nbCorrectionsRead`,NEW.`iPrecision`,NEW.`iAutonomy`,NEW.`sStartDate`,NEW.`sValidationDate`,NEW.`sBestAnswerDate`,NEW.`sLastAnswerDate`,NEW.`sThreadStartDate`,NEW.`sLastHintDate`,NEW.`sFinishDate`,NEW.`sLastActivityDate`,NEW.`sContestStartDate`,NEW.`bRanked`,NEW.`sAllLangProg`); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_groups_attempts` BEFORE UPDATE ON `groups_attempts` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idGroup` <=> NEW.`idGroup` AND OLD.`idItem` <=> NEW.`idItem` AND OLD.`idUserCreator` <=> NEW.`idUserCreator` AND OLD.`iOrder` <=> NEW.`iOrder` AND OLD.`iScore` <=> NEW.`iScore` AND OLD.`iScoreComputed` <=> NEW.`iScoreComputed` AND OLD.`iScoreReeval` <=> NEW.`iScoreReeval` AND OLD.`iScoreDiffManual` <=> NEW.`iScoreDiffManual` AND OLD.`sScoreDiffComment` <=> NEW.`sScoreDiffComment` AND OLD.`nbSubmissionsAttempts` <=> NEW.`nbSubmissionsAttempts` AND OLD.`nbTasksTried` <=> NEW.`nbTasksTried` AND OLD.`nbChildrenValidated` <=> NEW.`nbChildrenValidated` AND OLD.`bValidated` <=> NEW.`bValidated` AND OLD.`bFinished` <=> NEW.`bFinished` AND OLD.`bKeyObtained` <=> NEW.`bKeyObtained` AND OLD.`nbTasksWithHelp` <=> NEW.`nbTasksWithHelp` AND OLD.`sHintsRequested` <=> NEW.`sHintsRequested` AND OLD.`nbHintsCached` <=> NEW.`nbHintsCached` AND OLD.`nbCorrectionsRead` <=> NEW.`nbCorrectionsRead` AND OLD.`iPrecision` <=> NEW.`iPrecision` AND OLD.`iAutonomy` <=> NEW.`iAutonomy` AND OLD.`sStartDate` <=> NEW.`sStartDate` AND OLD.`sValidationDate` <=> NEW.`sValidationDate` AND OLD.`sBestAnswerDate` <=> NEW.`sBestAnswerDate` AND OLD.`sLastAnswerDate` <=> NEW.`sLastAnswerDate` AND OLD.`sThreadStartDate` <=> NEW.`sThreadStartDate` AND OLD.`sLastHintDate` <=> NEW.`sLastHintDate` AND OLD.`sFinishDate` <=> NEW.`sFinishDate` AND OLD.`sContestStartDate` <=> NEW.`sContestStartDate` AND OLD.`bRanked` <=> NEW.`bRanked` AND OLD.`sAllLangProg` <=> NEW.`sAllLangProg`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_groups_attempts` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_groups_attempts` (`ID`,`iVersion`,`idGroup`,`idItem`,`idUserCreator`,`iOrder`,`iScore`,`iScoreComputed`,`iScoreReeval`,`iScoreDiffManual`,`sScoreDiffComment`,`nbSubmissionsAttempts`,`nbTasksTried`,`nbChildrenValidated`,`bValidated`,`bFinished`,`bKeyObtained`,`nbTasksWithHelp`,`sHintsRequested`,`nbHintsCached`,`nbCorrectionsRead`,`iPrecision`,`iAutonomy`,`sStartDate`,`sValidationDate`,`sBestAnswerDate`,`sLastAnswerDate`,`sThreadStartDate`,`sLastHintDate`,`sFinishDate`,`sLastActivityDate`,`sContestStartDate`,`bRanked`,`sAllLangProg`)       VALUES (NEW.`ID`,@curVersion,NEW.`idGroup`,NEW.`idItem`,NEW.`idUserCreator`,NEW.`iOrder`,NEW.`iScore`,NEW.`iScoreComputed`,NEW.`iScoreReeval`,NEW.`iScoreDiffManual`,NEW.`sScoreDiffComment`,NEW.`nbSubmissionsAttempts`,NEW.`nbTasksTried`,NEW.`nbChildrenValidated`,NEW.`bValidated`,NEW.`bFinished`,NEW.`bKeyObtained`,NEW.`nbTasksWithHelp`,NEW.`sHintsRequested`,NEW.`nbHintsCached`,NEW.`nbCorrectionsRead`,NEW.`iPrecision`,NEW.`iAutonomy`,NEW.`sStartDate`,NEW.`sValidationDate`,NEW.`sBestAnswerDate`,NEW.`sLastAnswerDate`,NEW.`sThreadStartDate`,NEW.`sLastHintDate`,NEW.`sFinishDate`,NEW.`sLastActivityDate`,NEW.`sContestStartDate`,NEW.`bRanked`,NEW.`sAllLangProg`) ; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_groups_attempts` BEFORE DELETE ON `groups_attempts` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_attempts` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_groups_attempts` (`ID`,`iVersion`,`idGroup`,`idItem`,`idUserCreator`,`iOrder`,`iScore`,`iScoreComputed`,`iScoreReeval`,`iScoreDiffManual`,`sScoreDiffComment`,`nbSubmissionsAttempts`,`nbTasksTried`,`nbChildrenValidated`,`bValidated`,`bFinished`,`bKeyObtained`,`nbTasksWithHelp`,`sHintsRequested`,`nbHintsCached`,`nbCorrectionsRead`,`iPrecision`,`iAutonomy`,`sStartDate`,`sValidationDate`,`sBestAnswerDate`,`sLastAnswerDate`,`sThreadStartDate`,`sLastHintDate`,`sFinishDate`,`sLastActivityDate`,`sContestStartDate`,`bRanked`,`sAllLangProg`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idGroup`,OLD.`idItem`,OLD.`idUserCreator`,OLD.`iOrder`,OLD.`iScore`,OLD.`iScoreComputed`,OLD.`iScoreReeval`,OLD.`iScoreDiffManual`,OLD.`sScoreDiffComment`,OLD.`nbSubmissionsAttempts`,OLD.`nbTasksTried`,OLD.`nbChildrenValidated`,OLD.`bValidated`,OLD.`bFinished`,OLD.`bKeyObtained`,OLD.`nbTasksWithHelp`,OLD.`sHintsRequested`,OLD.`nbHintsCached`,OLD.`nbCorrectionsRead`,OLD.`iPrecision`,OLD.`iAutonomy`,OLD.`sStartDate`,OLD.`sValidationDate`,OLD.`sBestAnswerDate`,OLD.`sLastAnswerDate`,OLD.`sThreadStartDate`,OLD.`sLastHintDate`,OLD.`sFinishDate`,OLD.`sLastActivityDate`,OLD.`sContestStartDate`,OLD.`bRanked`,OLD.`sAllLangProg`, 1); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `groups_groups`
--

DROP TABLE IF EXISTS `groups_groups`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `groups_groups` (
  `ID` bigint(20) NOT NULL AUTO_INCREMENT,
  `idGroupParent` bigint(20) NOT NULL,
  `idGroupChild` bigint(20) NOT NULL,
  `iChildOrder` int(11) NOT NULL DEFAULT '0',
  `sType` enum('invitationSent','requestSent','invitationAccepted','requestAccepted','invitationRefused','requestRefused','removed','left','direct') NOT NULL DEFAULT 'direct',
  `sRole` enum('manager','owner','member','observer') NOT NULL DEFAULT 'member',
  `idUserInviting` int(11) DEFAULT NULL,
  `sStatusDate` datetime DEFAULT NULL,
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `parentchild` (`idGroupParent`,`idGroupChild`),
  KEY `iVersion` (`iVersion`),
  KEY `idGroupChild` (`idGroupChild`),
  KEY `idGroupParent` (`idGroupParent`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `groups_groups`
--

LOCK TABLES `groups_groups` WRITE;
ALTER TABLE `groups_groups` DISABLE KEYS;
ALTER TABLE `groups_groups` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_groups_groups` BEFORE INSERT ON `groups_groups` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; INSERT IGNORE INTO `groups_propagate` (ID, sAncestorsComputationState) VALUES (NEW.idGroupChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo' ; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_groups_groups` AFTER INSERT ON `groups_groups` FOR EACH ROW BEGIN INSERT INTO `history_groups_groups` (`ID`,`iVersion`,`idGroupParent`,`idGroupChild`,`iChildOrder`,`sType`,`sRole`,`sStatusDate`,`idUserInviting`) VALUES (NEW.`ID`,@curVersion,NEW.`idGroupParent`,NEW.`idGroupChild`,NEW.`iChildOrder`,NEW.`sType`,NEW.`sRole`,NEW.`sStatusDate`,NEW.`idUserInviting`); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idGroupParent` <=> NEW.`idGroupParent` AND OLD.`idGroupChild` <=> NEW.`idGroupChild` AND OLD.`iChildOrder` <=> NEW.`iChildOrder` AND OLD.`sType` <=> NEW.`sType` AND OLD.`sRole` <=> NEW.`sRole` AND OLD.`sStatusDate` <=> NEW.`sStatusDate` AND OLD.`idUserInviting` <=> NEW.`idUserInviting`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_groups_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_groups_groups` (`ID`,`iVersion`,`idGroupParent`,`idGroupChild`,`iChildOrder`,`sType`,`sRole`,`sStatusDate`,`idUserInviting`)       VALUES (NEW.`ID`,@curVersion,NEW.`idGroupParent`,NEW.`idGroupChild`,NEW.`iChildOrder`,NEW.`sType`,NEW.`sRole`,NEW.`sStatusDate`,NEW.`idUserInviting`) ; END IF; IF (OLD.idGroupChild != NEW.idGroupChild OR OLD.idGroupParent != NEW.idGroupParent OR OLD.sType != NEW.sType) THEN INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idGroupChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idGroupParent, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) (SELECT `groups_ancestors`.`idGroupChild`, 'todo' FROM `groups_ancestors` WHERE `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupChild`) ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; DELETE `groups_ancestors` from `groups_ancestors` WHERE `groups_ancestors`.`idGroupChild` = OLD.`idGroupChild` and `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupParent`;DELETE `bridges` FROM `groups_ancestors` `child_descendants` JOIN `groups_ancestors` `parent_ancestors` JOIN `groups_ancestors` `bridges` ON (`bridges`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor` AND `bridges`.`idGroupChild` = `child_descendants`.`idGroupChild`) WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent` AND `child_descendants`.`idGroupAncestor` = OLD.`idGroupChild`; DELETE `child_ancestors` FROM `groups_ancestors` `child_ancestors` JOIN  `groups_ancestors` `parent_ancestors` ON (`child_ancestors`.`idGroupChild` = OLD.`idGroupChild` AND `child_ancestors`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor`) WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent`; DELETE `parent_ancestors` FROM `groups_ancestors` `parent_ancestors` JOIN  `groups_ancestors` `child_ancestors` ON (`parent_ancestors`.`idGroupAncestor` = OLD.`idGroupParent` AND `child_ancestors`.`idGroupChild` = `parent_ancestors`.`idGroupChild`) WHERE `child_ancestors`.`idGroupAncestor` = OLD.`idGroupChild`  ; END IF; IF (OLD.idGroupChild != NEW.idGroupChild OR OLD.idGroupParent != NEW.idGroupParent OR OLD.sType != NEW.sType) THEN INSERT IGNORE INTO `groups_propagate` (ID, sAncestorsComputationState) VALUES (NEW.idGroupChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'  ; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_groups_groups` (`ID`,`iVersion`,`idGroupParent`,`idGroupChild`,`iChildOrder`,`sType`,`sRole`,`sStatusDate`,`idUserInviting`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idGroupParent`,OLD.`idGroupChild`,OLD.`iChildOrder`,OLD.`sType`,OLD.`sRole`,OLD.`sStatusDate`,OLD.`idUserInviting`, 1); INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idGroupChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idGroupParent, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) (SELECT `groups_ancestors`.`idGroupChild`, 'todo' FROM `groups_ancestors` WHERE `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupChild`) ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; DELETE `groups_ancestors` from `groups_ancestors` WHERE `groups_ancestors`.`idGroupChild` = OLD.`idGroupChild` and `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupParent`;DELETE `bridges` FROM `groups_ancestors` `child_descendants` JOIN `groups_ancestors` `parent_ancestors` JOIN `groups_ancestors` `bridges` ON (`bridges`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor` AND `bridges`.`idGroupChild` = `child_descendants`.`idGroupChild`) WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent` AND `child_descendants`.`idGroupAncestor` = OLD.`idGroupChild`; DELETE `child_ancestors` FROM `groups_ancestors` `child_ancestors` JOIN  `groups_ancestors` `parent_ancestors` ON (`child_ancestors`.`idGroupChild` = OLD.`idGroupChild` AND `child_ancestors`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor`) WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent`; DELETE `parent_ancestors` FROM `groups_ancestors` `parent_ancestors` JOIN  `groups_ancestors` `child_ancestors` ON (`parent_ancestors`.`idGroupAncestor` = OLD.`idGroupParent` AND `child_ancestors`.`idGroupChild` = `parent_ancestors`.`idGroupChild`) WHERE `child_ancestors`.`idGroupAncestor` = OLD.`idGroupChild` ; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `groups_items`
--

DROP TABLE IF EXISTS `groups_items`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `groups_items` (
  `ID` bigint(20) NOT NULL,
  `idGroup` bigint(20) NOT NULL,
  `idItem` bigint(20) NOT NULL,
  `idUserCreated` bigint(20) NOT NULL,
  `sPartialAccessDate` datetime DEFAULT NULL,
  `sAccessReason` varchar(200) DEFAULT NULL,
  `sFullAccessDate` datetime DEFAULT NULL,
  `sAccessSolutionsDate` datetime DEFAULT NULL,
  `bOwnerAccess` tinyint(1) NOT NULL DEFAULT '0',
  `bManagerAccess` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'not for inherited access',
  `sCachedFullAccessDate` datetime DEFAULT NULL,
  `sCachedPartialAccessDate` datetime DEFAULT NULL,
  `sCachedAccessSolutionsDate` datetime DEFAULT NULL,
  `sCachedGrayedAccessDate` datetime DEFAULT NULL,
  `sCachedAccessReason` varchar(200) DEFAULT NULL,
  `bCachedFullAccess` tinyint(1) NOT NULL DEFAULT '0',
  `bCachedPartialAccess` tinyint(1) NOT NULL DEFAULT '0',
  `bCachedAccessSolutions` tinyint(1) NOT NULL DEFAULT '0',
  `bCachedGrayedAccess` tinyint(1) NOT NULL DEFAULT '0',
  `bCachedManagerAccess` tinyint(1) NOT NULL DEFAULT '0',
  `sPropagateAccess` enum('self','children','done') NOT NULL DEFAULT 'self',
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `idItem` (`idItem`,`idGroup`),
  KEY `iVersion` (`iVersion`),
  KEY `idGroup` (`idGroup`) COMMENT 'idGroup',
  KEY `idItemtem` (`idItem`),
  KEY `fullAccess` (`bCachedFullAccess`,`sCachedFullAccessDate`),
  KEY `accessSolutions` (`bCachedAccessSolutions`,`sCachedAccessSolutionsDate`),
  KEY `sPropagateAccess` (`sPropagateAccess`),
  KEY `partialAccess` (`bCachedPartialAccess`,`sCachedPartialAccessDate`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `groups_items`
--

LOCK TABLES `groups_items` WRITE;
ALTER TABLE `groups_items` DISABLE KEYS;
ALTER TABLE `groups_items` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_groups_items` BEFORE INSERT ON `groups_items` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; SET NEW.`sPropagateAccess`='self' ; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_groups_items` AFTER INSERT ON `groups_items` FOR EACH ROW BEGIN INSERT INTO `history_groups_items` (`ID`,`iVersion`,`idGroup`,`idItem`,`idUserCreated`,`sPartialAccessDate`,`sFullAccessDate`,`sAccessReason`,`sAccessSolutionsDate`,`bOwnerAccess`,`bManagerAccess`,`sCachedPartialAccessDate`,`sCachedFullAccessDate`,`sCachedAccessSolutionsDate`,`sCachedGrayedAccessDate`,`bCachedFullAccess`,`bCachedPartialAccess`,`bCachedAccessSolutions`,`bCachedGrayedAccess`,`bCachedManagerAccess`,`sPropagateAccess`) VALUES (NEW.`ID`,@curVersion,NEW.`idGroup`,NEW.`idItem`,NEW.`idUserCreated`,NEW.`sPartialAccessDate`,NEW.`sFullAccessDate`,NEW.`sAccessReason`,NEW.`sAccessSolutionsDate`,NEW.`bOwnerAccess`,NEW.`bManagerAccess`,NEW.`sCachedPartialAccessDate`,NEW.`sCachedFullAccessDate`,NEW.`sCachedAccessSolutionsDate`,NEW.`sCachedGrayedAccessDate`,NEW.`bCachedFullAccess`,NEW.`bCachedPartialAccess`,NEW.`bCachedAccessSolutions`,NEW.`bCachedGrayedAccess`,NEW.`bCachedManagerAccess`,NEW.`sPropagateAccess`); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_groups_items` BEFORE UPDATE ON `groups_items` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idGroup` <=> NEW.`idGroup` AND OLD.`idItem` <=> NEW.`idItem` AND OLD.`idUserCreated` <=> NEW.`idUserCreated` AND OLD.`sPartialAccessDate` <=> NEW.`sPartialAccessDate` AND OLD.`sFullAccessDate` <=> NEW.`sFullAccessDate` AND OLD.`sAccessReason` <=> NEW.`sAccessReason` AND OLD.`sAccessSolutionsDate` <=> NEW.`sAccessSolutionsDate` AND OLD.`bOwnerAccess` <=> NEW.`bOwnerAccess` AND OLD.`bManagerAccess` <=> NEW.`bManagerAccess` AND OLD.`sCachedPartialAccessDate` <=> NEW.`sCachedPartialAccessDate` AND OLD.`sCachedFullAccessDate` <=> NEW.`sCachedFullAccessDate` AND OLD.`sCachedAccessSolutionsDate` <=> NEW.`sCachedAccessSolutionsDate` AND OLD.`sCachedGrayedAccessDate` <=> NEW.`sCachedGrayedAccessDate` AND OLD.`bCachedFullAccess` <=> NEW.`bCachedFullAccess` AND OLD.`bCachedPartialAccess` <=> NEW.`bCachedPartialAccess` AND OLD.`bCachedAccessSolutions` <=> NEW.`bCachedAccessSolutions` AND OLD.`bCachedGrayedAccess` <=> NEW.`bCachedGrayedAccess` AND OLD.`bCachedManagerAccess` <=> NEW.`bCachedManagerAccess`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_groups_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_groups_items` (`ID`,`iVersion`,`idGroup`,`idItem`,`idUserCreated`,`sPartialAccessDate`,`sFullAccessDate`,`sAccessReason`,`sAccessSolutionsDate`,`bOwnerAccess`,`bManagerAccess`,`sCachedPartialAccessDate`,`sCachedFullAccessDate`,`sCachedAccessSolutionsDate`,`sCachedGrayedAccessDate`,`bCachedFullAccess`,`bCachedPartialAccess`,`bCachedAccessSolutions`,`bCachedGrayedAccess`,`bCachedManagerAccess`,`sPropagateAccess`)       VALUES (NEW.`ID`,@curVersion,NEW.`idGroup`,NEW.`idItem`,NEW.`idUserCreated`,NEW.`sPartialAccessDate`,NEW.`sFullAccessDate`,NEW.`sAccessReason`,NEW.`sAccessSolutionsDate`,NEW.`bOwnerAccess`,NEW.`bManagerAccess`,NEW.`sCachedPartialAccessDate`,NEW.`sCachedFullAccessDate`,NEW.`sCachedAccessSolutionsDate`,NEW.`sCachedGrayedAccessDate`,NEW.`bCachedFullAccess`,NEW.`bCachedPartialAccess`,NEW.`bCachedAccessSolutions`,NEW.`bCachedGrayedAccess`,NEW.`bCachedManagerAccess`,NEW.`sPropagateAccess`) ; END IF; IF NOT (NEW.`sFullAccessDate` <=> OLD.`sFullAccessDate`AND NEW.`sPartialAccessDate` <=> OLD.`sPartialAccessDate`AND NEW.`sAccessSolutionsDate` <=> OLD.`sAccessSolutionsDate`AND NEW.`bManagerAccess` <=> OLD.`bManagerAccess`AND NEW.`sAccessReason` <=> OLD.`sAccessReason`)THEN SET NEW.`sPropagateAccess` = 'self'; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_groups_items` BEFORE DELETE ON `groups_items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_groups_items` (`ID`,`iVersion`,`idGroup`,`idItem`,`idUserCreated`,`sPartialAccessDate`,`sFullAccessDate`,`sAccessReason`,`sAccessSolutionsDate`,`bOwnerAccess`,`bManagerAccess`,`sCachedPartialAccessDate`,`sCachedFullAccessDate`,`sCachedAccessSolutionsDate`,`sCachedGrayedAccessDate`,`bCachedFullAccess`,`bCachedPartialAccess`,`bCachedAccessSolutions`,`bCachedGrayedAccess`,`bCachedManagerAccess`,`sPropagateAccess`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idGroup`,OLD.`idItem`,OLD.`idUserCreated`,OLD.`sPartialAccessDate`,OLD.`sFullAccessDate`,OLD.`sAccessReason`,OLD.`sAccessSolutionsDate`,OLD.`bOwnerAccess`,OLD.`bManagerAccess`,OLD.`sCachedPartialAccessDate`,OLD.`sCachedFullAccessDate`,OLD.`sCachedAccessSolutionsDate`,OLD.`sCachedGrayedAccessDate`,OLD.`bCachedFullAccess`,OLD.`bCachedPartialAccess`,OLD.`bCachedAccessSolutions`,OLD.`bCachedGrayedAccess`,OLD.`bCachedManagerAccess`,OLD.`sPropagateAccess`, 1); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_delete_groups_items` AFTER DELETE ON `groups_items` FOR EACH ROW BEGIN DELETE FROM groups_items_propagate where ID = OLD.ID ; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `groups_items_propagate`
--

DROP TABLE IF EXISTS `groups_items_propagate`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `groups_items_propagate` (
  `ID` bigint(20) NOT NULL,
  `sPropagateAccess` enum('self','children','done') NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `sPropagateAccess` (`sPropagateAccess`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `groups_items_propagate`
--

LOCK TABLES `groups_items_propagate` WRITE;
ALTER TABLE `groups_items_propagate` DISABLE KEYS;
ALTER TABLE `groups_items_propagate` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `groups_login_prefixes`
--

DROP TABLE IF EXISTS `groups_login_prefixes`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `groups_login_prefixes` (
  `ID` bigint(20) NOT NULL AUTO_INCREMENT,
  `idGroup` bigint(20) NOT NULL,
  `prefix` varchar(100) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL,
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `prefix` (`prefix`),
  KEY `idGroup` (`idGroup`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `groups_login_prefixes`
--

LOCK TABLES `groups_login_prefixes` WRITE;
ALTER TABLE `groups_login_prefixes` DISABLE KEYS;
ALTER TABLE `groups_login_prefixes` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_groups_login_prefixes` BEFORE INSERT ON `groups_login_prefixes` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_groups_login_prefixes` AFTER INSERT ON `groups_login_prefixes` FOR EACH ROW BEGIN INSERT INTO `history_groups_login_prefixes` (`ID`,`iVersion`,`idGroup`,`prefix`) VALUES (NEW.`ID`,@curVersion,NEW.`idGroup`,NEW.`prefix`); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_groups_login_prefixes` BEFORE UPDATE ON `groups_login_prefixes` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idGroup` <=> NEW.`idGroup` AND OLD.`prefix` <=> NEW.`prefix`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_groups_login_prefixes` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_groups_login_prefixes` (`ID`,`iVersion`,`idGroup`,`prefix`)       VALUES (NEW.`ID`,@curVersion,NEW.`idGroup`,NEW.`prefix`) ; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_groups_login_prefixes` BEFORE DELETE ON `groups_login_prefixes` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_login_prefixes` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_groups_login_prefixes` (`ID`,`iVersion`,`idGroup`,`prefix`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idGroup`,OLD.`prefix`, 1); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `groups_propagate`
--

DROP TABLE IF EXISTS `groups_propagate`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `groups_propagate` (
  `ID` bigint(20) NOT NULL,
  `sAncestorsComputationState` enum('todo','done','processing','') NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `sAncestorsComputationState` (`sAncestorsComputationState`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `groups_propagate`
--

LOCK TABLES `groups_propagate` WRITE;
ALTER TABLE `groups_propagate` DISABLE KEYS;
ALTER TABLE `groups_propagate` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_filters`
--

DROP TABLE IF EXISTS `history_filters`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_filters` (
  `historyID` int(11) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `idUser` bigint(20) NOT NULL,
  `sName` varchar(45) NOT NULL DEFAULT '',
  `bSelected` tinyint(1) NOT NULL DEFAULT '0',
  `bStarred` tinyint(1) DEFAULT NULL,
  `sStartDate` datetime DEFAULT NULL,
  `sEndDate` datetime DEFAULT NULL,
  `bArchived` tinyint(1) DEFAULT NULL,
  `bParticipated` tinyint(1) DEFAULT NULL,
  `bUnread` tinyint(1) DEFAULT NULL,
  `idItem` bigint(20) DEFAULT NULL,
  `idGroup` bigint(20) DEFAULT NULL,
  `olderThan` int(11) DEFAULT NULL,
  `newerThan` int(11) DEFAULT NULL,
  `sUsersSearch` varchar(200) DEFAULT NULL,
  `sBodySearch` varchar(100) DEFAULT NULL,
  `bImportant` tinyint(1) DEFAULT NULL,
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`historyID`),
  KEY `user_idx` (`idUser`),
  KEY `iVersion` (`iVersion`),
  KEY `iNextVersion` (`iNextVersion`),
  KEY `bDeleted` (`bDeleted`),
  KEY `ID` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_filters`
--

LOCK TABLES `history_filters` WRITE;
ALTER TABLE `history_filters` DISABLE KEYS;
ALTER TABLE `history_filters` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_groups`
--

DROP TABLE IF EXISTS `history_groups`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_groups` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `sName` varchar(200) NOT NULL DEFAULT '',
  `iGrade` int(4) NOT NULL DEFAULT '-2',
  `sGradeDetails` varchar(50) DEFAULT NULL,
  `sDescription` text,
  `sDateCreated` datetime DEFAULT NULL,
  `bOpened` tinyint(1) NOT NULL,
  `bFreeAccess` tinyint(1) NOT NULL,
  `idTeamItem` bigint(20) DEFAULT NULL,
  `iTeamParticipating` tinyint(1) NOT NULL DEFAULT '0',
  `sPassword` varchar(50) DEFAULT NULL,
  `sPasswordTimer` time DEFAULT NULL,
  `sPasswordEnd` datetime DEFAULT NULL,
  `sRedirectPath` text,
  `bOpenContest` tinyint(1) NOT NULL DEFAULT '0',
  `sType` enum('Root','Class','Team','Club','Friends','Other','UserSelf','UserAdmin','RootSelf','RootAdmin') NOT NULL,
  `bSendEmails` tinyint(1) NOT NULL,
  `bAncestorsComputed` tinyint(1) NOT NULL DEFAULT '0',
  `sAncestorsComputationState` enum('done','processing','todo') NOT NULL DEFAULT 'todo',
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  `lockUserDeletionDate` date DEFAULT NULL,
  PRIMARY KEY (`historyID`),
  KEY `iVersion` (`iVersion`),
  KEY `ID` (`ID`),
  KEY `iNextVersion` (`iNextVersion`),
  KEY `bDeleted` (`bDeleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_groups`
--

LOCK TABLES `history_groups` WRITE;
ALTER TABLE `history_groups` DISABLE KEYS;
ALTER TABLE `history_groups` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_groups_ancestors`
--

DROP TABLE IF EXISTS `history_groups_ancestors`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_groups_ancestors` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `idGroupAncestor` bigint(20) NOT NULL,
  `idGroupChild` bigint(20) NOT NULL,
  `bIsSelf` tinyint(1) NOT NULL DEFAULT '0',
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`historyID`),
  KEY `iVersion` (`iVersion`),
  KEY `iNextVersion` (`iNextVersion`),
  KEY `bDeleted` (`bDeleted`),
  KEY `idGroupAncestor` (`idGroupAncestor`,`idGroupChild`),
  KEY `ancestor` (`idGroupAncestor`),
  KEY `descendant` (`idGroupChild`),
  KEY `ID` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_groups_ancestors`
--

LOCK TABLES `history_groups_ancestors` WRITE;
ALTER TABLE `history_groups_ancestors` DISABLE KEYS;
ALTER TABLE `history_groups_ancestors` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_groups_attempts`
--

DROP TABLE IF EXISTS `history_groups_attempts`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_groups_attempts` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `idGroup` bigint(20) NOT NULL,
  `idItem` bigint(20) NOT NULL,
  `idUserCreator` bigint(20) DEFAULT NULL,
  `iOrder` int(11) NOT NULL,
  `iScore` float NOT NULL DEFAULT '0',
  `iScoreComputed` float NOT NULL DEFAULT '0',
  `iScoreReeval` float DEFAULT '0',
  `iScoreDiffManual` float NOT NULL DEFAULT '0',
  `sScoreDiffComment` varchar(200) NOT NULL DEFAULT '',
  `nbSubmissionsAttempts` int(11) NOT NULL DEFAULT '0',
  `nbTasksTried` int(11) NOT NULL DEFAULT '0',
  `nbTasksSolved` int(11) NOT NULL DEFAULT '0',
  `nbChildrenValidated` int(11) NOT NULL DEFAULT '0',
  `bValidated` tinyint(1) NOT NULL DEFAULT '0',
  `bFinished` tinyint(1) NOT NULL DEFAULT '0',
  `bKeyObtained` tinyint(1) NOT NULL DEFAULT '0',
  `nbTasksWithHelp` int(11) NOT NULL DEFAULT '0',
  `sHintsRequested` mediumtext,
  `nbHintsCached` int(11) NOT NULL DEFAULT '0',
  `nbCorrectionsRead` int(11) NOT NULL DEFAULT '0',
  `iPrecision` int(11) NOT NULL DEFAULT '0',
  `iAutonomy` int(11) NOT NULL DEFAULT '0',
  `sStartDate` datetime DEFAULT NULL,
  `sValidationDate` datetime DEFAULT NULL,
  `sFinishDate` datetime DEFAULT NULL,
  `sLastActivityDate` datetime DEFAULT NULL,
  `sThreadStartDate` datetime DEFAULT NULL,
  `sBestAnswerDate` datetime DEFAULT NULL,
  `sLastAnswerDate` datetime DEFAULT NULL,
  `sLastHintDate` datetime DEFAULT NULL,
  `sAdditionalTime` datetime DEFAULT NULL,
  `sContestStartDate` datetime DEFAULT NULL,
  `bRanked` tinyint(1) NOT NULL DEFAULT '0',
  `sAllLangProg` varchar(200) DEFAULT NULL,
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL,
  PRIMARY KEY (`historyID`),
  KEY `iVersion` (`iVersion`),
  KEY `idItem` (`idItem`),
  KEY `GroupItem` (`idGroup`,`idItem`),
  KEY `idGroup` (`idGroup`),
  KEY `ID` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_groups_attempts`
--

LOCK TABLES `history_groups_attempts` WRITE;
ALTER TABLE `history_groups_attempts` DISABLE KEYS;
ALTER TABLE `history_groups_attempts` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_groups_groups`
--

DROP TABLE IF EXISTS `history_groups_groups`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_groups_groups` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `idGroupParent` bigint(20) NOT NULL,
  `idGroupChild` bigint(20) NOT NULL,
  `iChildOrder` int(11) NOT NULL,
  `sType` enum('invitationSent','requestSent','invitationAccepted','requestAccepted','invitationRefused','requestRefused','removed','left','direct') NOT NULL DEFAULT 'direct',
  `sRole` enum('manager','owner','member','observer') NOT NULL DEFAULT 'member',
  `idUserInviting` int(11) DEFAULT NULL,
  `sStatusDate` datetime DEFAULT NULL,
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`historyID`),
  KEY `iVersion` (`iVersion`),
  KEY `ID` (`ID`),
  KEY `iNextVersion` (`iNextVersion`),
  KEY `bDeleted` (`bDeleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_groups_groups`
--

LOCK TABLES `history_groups_groups` WRITE;
ALTER TABLE `history_groups_groups` DISABLE KEYS;
ALTER TABLE `history_groups_groups` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_groups_items`
--

DROP TABLE IF EXISTS `history_groups_items`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_groups_items` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `idGroup` bigint(20) NOT NULL,
  `idItem` bigint(20) NOT NULL,
  `idUserCreated` bigint(20) NOT NULL,
  `sPartialAccessDate` datetime DEFAULT NULL,
  `sAccessReason` varchar(200) DEFAULT NULL,
  `sFullAccessDate` datetime DEFAULT NULL,
  `sAccessSolutionsDate` datetime DEFAULT NULL,
  `bOwnerAccess` tinyint(1) NOT NULL DEFAULT '0',
  `bManagerAccess` tinyint(1) NOT NULL DEFAULT '0',
  `sCachedFullAccessDate` datetime DEFAULT NULL,
  `sCachedPartialAccessDate` datetime DEFAULT NULL,
  `sCachedAccessSolutionsDate` datetime DEFAULT NULL,
  `sCachedGrayedAccessDate` datetime DEFAULT NULL,
  `sCachedAccessReason` varchar(200) DEFAULT NULL,
  `bCachedFullAccess` tinyint(1) NOT NULL DEFAULT '0',
  `bCachedPartialAccess` tinyint(1) NOT NULL DEFAULT '0',
  `bCachedAccessSolutions` tinyint(1) NOT NULL DEFAULT '0',
  `bCachedGrayedAccess` tinyint(1) NOT NULL DEFAULT '0',
  `bCachedManagerAccess` tinyint(1) NOT NULL DEFAULT '0',
  `sPropagateAccess` enum('self','children','done') NOT NULL DEFAULT 'self',
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`historyID`),
  KEY `iVersion` (`iVersion`),
  KEY `ID` (`ID`),
  KEY `itemGroup` (`idItem`,`idGroup`),
  KEY `iNextVersion` (`iNextVersion`),
  KEY `bDeleted` (`bDeleted`),
  KEY `idItem` (`idItem`),
  KEY `idGroup` (`idGroup`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_groups_items`
--

LOCK TABLES `history_groups_items` WRITE;
ALTER TABLE `history_groups_items` DISABLE KEYS;
ALTER TABLE `history_groups_items` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_groups_login_prefixes`
--

DROP TABLE IF EXISTS `history_groups_login_prefixes`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_groups_login_prefixes` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `idGroup` bigint(20) NOT NULL,
  `prefix` varchar(100) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL,
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(4) NOT NULL DEFAULT '0',
  PRIMARY KEY (`historyID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_groups_login_prefixes`
--

LOCK TABLES `history_groups_login_prefixes` WRITE;
ALTER TABLE `history_groups_login_prefixes` DISABLE KEYS;
ALTER TABLE `history_groups_login_prefixes` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_items`
--

DROP TABLE IF EXISTS `history_items`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_items` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `sUrl` varchar(200) DEFAULT NULL,
  `sOptions` TEXT NOT NULL,
  `idPlatform` int(11) DEFAULT NULL,
  `sTextId` varchar(200) DEFAULT NULL,
  `sRepositoryPath` text,
  `sType` enum('Root','CustomProgressRoot','OfficialProgressRoot','CustomContestRoot','OfficialContestRoot','DomainRoot','Category','Level','Chapter','GenericChapter','StaticChapter','Section','Task','Course','ContestChapter','LimitedTimeChapter','Presentation') NOT NULL,
  `bTitleBarVisible` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `bTransparentFolder` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `bDisplayDetailsInParent` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT 'when true, display a large icon, the subtitle, and more within the parent chapter',
  `bCustomChapter` tinyint(3) unsigned DEFAULT '0' COMMENT 'true if this is a chapter where users can add their own content. access to this chapter will not be propagated to its children',
  `bDisplayChildrenAsTabs` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `bUsesAPI` tinyint(1) NOT NULL DEFAULT '1',
  `bReadOnly` tinyint(1) NOT NULL DEFAULT '0',
  `sFullScreen` enum('forceYes','forceNo','default','') NOT NULL DEFAULT 'default',
  `bShowDifficulty` tinyint(1) NOT NULL,
  `bShowSource` tinyint(1) NOT NULL,
  `bHintsAllowed` tinyint(1) NOT NULL,
  `bFixedRanks` tinyint(1) NOT NULL DEFAULT '0',
  `sValidationType` enum('None','All','AllButOne','Categories','One') NOT NULL DEFAULT 'All',
  `iValidationMin` int(11) DEFAULT NULL,
  `sPreparationState` enum('NotReady','Reviewing','Ready') NOT NULL DEFAULT 'NotReady',
  `idItemUnlocked` text,
  `iScoreMinUnlock` int(11) NOT NULL DEFAULT '100',
  `sSupportedLangProg` varchar(200) DEFAULT NULL,
  `idDefaultLanguage` bigint(20) DEFAULT '1',
  `sTeamMode` enum('All','Half','One','None') DEFAULT NULL,
  `bTeamsEditable` tinyint(1) NOT NULL,
  `idTeamInGroup` bigint(20) DEFAULT NULL,
  `iTeamMaxMembers` int(11) NOT NULL DEFAULT '0',
  `bHasAttempts` tinyint(1) NOT NULL DEFAULT '0',
  `sAccessOpenDate` datetime DEFAULT NULL,
  `sDuration` time DEFAULT NULL,
  `sEndContestDate` datetime DEFAULT NULL,
  `bShowUserInfos` tinyint(1) NOT NULL DEFAULT '0',
  `sContestPhase` enum('Running','Analysis','Closed') NOT NULL,
  `iLevel` int(11) DEFAULT NULL,
  `bNoScore` tinyint(1) NOT NULL,
  `groupCodeEnter` tinyint(1) DEFAULT '0' COMMENT 'Offer users to enter through a group code',
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`historyID`),
  KEY `ID` (`ID`),
  KEY `iVersion` (`iVersion`),
  KEY `iNextVersion` (`iNextVersion`),
  KEY `bDeleted` (`bDeleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_items`
--

LOCK TABLES `history_items` WRITE;
ALTER TABLE `history_items` DISABLE KEYS;
ALTER TABLE `history_items` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_items_ancestors`
--

DROP TABLE IF EXISTS `history_items_ancestors`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_items_ancestors` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `idItemAncestor` bigint(20) NOT NULL,
  `idItemChild` bigint(20) NOT NULL,
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`historyID`),
  KEY `iVersion` (`iVersion`),
  KEY `iNextVersion` (`iNextVersion`),
  KEY `bDeleted` (`bDeleted`),
  KEY `idItemAncestor` (`idItemAncestor`,`idItemChild`),
  KEY `idItemAncestortor` (`idItemAncestor`),
  KEY `idItemChild` (`idItemChild`),
  KEY `ID` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_items_ancestors`
--

LOCK TABLES `history_items_ancestors` WRITE;
ALTER TABLE `history_items_ancestors` DISABLE KEYS;
ALTER TABLE `history_items_ancestors` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_items_items`
--

DROP TABLE IF EXISTS `history_items_items`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_items_items` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `idItemParent` bigint(20) NOT NULL,
  `idItemChild` bigint(20) NOT NULL,
  `iChildOrder` int(11) NOT NULL,
  `sCategory` enum('Undefined','Discovery','Application','Validation','Challenge') NOT NULL DEFAULT 'Undefined',
  `bAlwaysVisible` tinyint(1) NOT NULL DEFAULT '0',
  `bAccessRestricted` tinyint(1) NOT NULL DEFAULT '1',
  `iDifficulty` int(11) NOT NULL,
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`historyID`),
  KEY `ID` (`ID`),
  KEY `iVersion` (`iVersion`),
  KEY `idItemParent` (`idItemParent`),
  KEY `idItemChild` (`idItemChild`),
  KEY `iNextVersion` (`iNextVersion`),
  KEY `bDeleted` (`bDeleted`),
  KEY `parentChild` (`idItemParent`,`idItemChild`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_items_items`
--

LOCK TABLES `history_items_items` WRITE;
ALTER TABLE `history_items_items` DISABLE KEYS;
ALTER TABLE `history_items_items` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_items_strings`
--

DROP TABLE IF EXISTS `history_items_strings`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_items_strings` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `idItem` bigint(20) NOT NULL,
  `idLanguage` bigint(20) NOT NULL,
  `sTranslator` varchar(100) DEFAULT NULL,
  `sTitle` varchar(200) DEFAULT NULL,
  `sImageUrl` text,
  `sSubtitle` varchar(200) DEFAULT NULL,
  `sDescription` text,
  `sEduComment` text,
  `sRankingComment` text,
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`historyID`),
  KEY `ID` (`ID`),
  KEY `iVersion` (`iVersion`),
  KEY `itemLanguage` (`idItem`,`idLanguage`),
  KEY `idItem` (`idItem`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_items_strings`
--

LOCK TABLES `history_items_strings` WRITE;
ALTER TABLE `history_items_strings` DISABLE KEYS;
ALTER TABLE `history_items_strings` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_languages`
--

DROP TABLE IF EXISTS `history_languages`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_languages` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `sName` varchar(100) NOT NULL DEFAULT '',
  `sCode` varchar(2) NOT NULL DEFAULT '',
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`historyID`),
  KEY `ID` (`ID`),
  KEY `iVersion` (`iVersion`),
  KEY `sCode` (`sCode`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_languages`
--

LOCK TABLES `history_languages` WRITE;
ALTER TABLE `history_languages` DISABLE KEYS;
ALTER TABLE `history_languages` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_messages`
--

DROP TABLE IF EXISTS `history_messages`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_messages` (
  `history_ID` int(11) NOT NULL AUTO_INCREMENT,
  `ID` int(11) NOT NULL,
  `idThread` int(11) NOT NULL,
  `idUser` bigint(20) DEFAULT NULL,
  `sSubmissionDate` datetime DEFAULT NULL,
  `bPublished` tinyint(1) NOT NULL DEFAULT '1',
  `sTitle` varchar(200) DEFAULT '',
  `sBody` varchar(2000) DEFAULT '',
  `bTrainersOnly` tinyint(1) NOT NULL DEFAULT '0',
  `bArchived` tinyint(1) DEFAULT '0',
  `bPersistant` tinyint(1) DEFAULT NULL,
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_ID`),
  KEY `thread` (`idThread`),
  KEY `iVersion` (`iVersion`),
  KEY `iNextVersion` (`iNextVersion`),
  KEY `bDeleted` (`bDeleted`),
  KEY `ID` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_messages`
--

LOCK TABLES `history_messages` WRITE;
ALTER TABLE `history_messages` DISABLE KEYS;
ALTER TABLE `history_messages` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_threads`
--

DROP TABLE IF EXISTS `history_threads`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_threads` (
  `history_ID` int(11) NOT NULL AUTO_INCREMENT,
  `ID` int(11) NOT NULL,
  `sType` enum('Help','Bug','General') NOT NULL,
  `idUserCreated` bigint(20) NOT NULL,
  `idItem` bigint(20) DEFAULT NULL,
  `sLastActivityDate` datetime NOT NULL,
  `sTitle` varchar(200) DEFAULT NULL,
  `bAdminHelpAsked` tinyint(1) NOT NULL DEFAULT '0',
  `bHidden` tinyint(1) NOT NULL DEFAULT '0',
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_ID`),
  KEY `iVersion` (`iVersion`),
  KEY `iNextVersion` (`iNextVersion`),
  KEY `bDeleted` (`bDeleted`),
  KEY `ID` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_threads`
--

LOCK TABLES `history_threads` WRITE;
ALTER TABLE `history_threads` DISABLE KEYS;
ALTER TABLE `history_threads` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_users`
--

DROP TABLE IF EXISTS `history_users`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_users` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `loginID` bigint(20) DEFAULT NULL,
  `sLogin` varchar(100) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '',
  `sOpenIdIdentity` varchar(255) DEFAULT NULL COMMENT 'User''s Open Id Identity',
  `sPasswordMd5` varchar(100) DEFAULT NULL,
  `sSalt` varchar(32) DEFAULT NULL,
  `sRecover` varchar(50) DEFAULT NULL,
  `sRegistrationDate` datetime DEFAULT NULL,
  `sEmail` varchar(100) DEFAULT NULL,
  `bEmailVerified` tinyint(1) NOT NULL DEFAULT '0',
  `sFirstName` varchar(100) DEFAULT NULL COMMENT 'User''s first name',
  `sLastName` varchar(100) DEFAULT NULL COMMENT 'User''s last name',
  `sStudentId` text,
  `sCountryCode` char(3) NOT NULL DEFAULT '',
  `sTimeZone` varchar(100) DEFAULT NULL,
  `sBirthDate` date DEFAULT NULL COMMENT 'User''s birth date',
  `iGraduationYear` int(11) NOT NULL DEFAULT '0' COMMENT 'User''s high school graduation year',
  `iGrade` int(11) DEFAULT NULL,
  `sSex` enum('Male','Female') DEFAULT NULL,
  `sAddress` mediumtext COMMENT 'User''s address',
  `sZipcode` longtext COMMENT 'User''s postal code',
  `sCity` longtext COMMENT 'User''s city',
  `sLandLineNumber` longtext COMMENT 'User''s phone number',
  `sCellPhoneNumber` longtext COMMENT 'User''s mobil phone number',
  `sDefaultLanguage` char(3) NOT NULL DEFAULT 'fr' COMMENT 'User''s default language',
  `bNotifyNews` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'TODO',
  `sNotify` enum('Never','Answers','Concerned') NOT NULL DEFAULT 'Answers',
  `bPublicFirstName` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Publicly show user''s first name',
  `bPublicLastName` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Publicly show user''s first name',
  `sFreeText` mediumtext,
  `sWebSite` varchar(100) DEFAULT NULL,
  `bPhotoAutoload` tinyint(1) NOT NULL DEFAULT '0',
  `sLangProg` varchar(30) DEFAULT 'Python',
  `sLastLoginDate` datetime DEFAULT NULL,
  `sLastActivityDate` datetime DEFAULT NULL COMMENT 'User''s last activity time on the website',
  `sLastIP` varchar(16) DEFAULT NULL,
  `bBasicEditorMode` tinyint(4) NOT NULL DEFAULT '1',
  `nbSpacesForTab` int(11) NOT NULL DEFAULT '3',
  `iMemberState` tinyint(4) NOT NULL DEFAULT '0',
  `idUserGodfather` int(11) DEFAULT NULL,
  `iStepLevelInSite` int(11) NOT NULL DEFAULT '0' COMMENT 'User''s level',
  `bIsAdmin` tinyint(4) NOT NULL DEFAULT '0',
  `bNoRanking` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'TODO',
  `nbHelpGiven` int(11) NOT NULL DEFAULT '0' COMMENT 'TODO',
  `idGroupSelf` bigint(20) DEFAULT NULL,
  `idGroupOwned` bigint(20) DEFAULT NULL,
  `idGroupAccess` bigint(20) DEFAULT NULL,
  `sNotificationReadDate` datetime DEFAULT NULL,
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  `loginModulePrefix` varchar(100) DEFAULT NULL,
  `creatorID` bigint(20) DEFAULT NULL COMMENT 'which user created a given login with the login generation tool',
  `allowSubgroups` tinyint(4) DEFAULT NULL COMMENT 'Allow to create subgroups',
  PRIMARY KEY (`historyID`),
  KEY `ID` (`ID`),
  KEY `iVersion` (`iVersion`),
  KEY `sCountryCode` (`sCountryCode`),
  KEY `idUserGodfather` (`idUserGodfather`),
  KEY `sLangProg` (`sLangProg`),
  KEY `iNextVersion` (`iNextVersion`),
  KEY `bDeleted` (`bDeleted`),
  KEY `idGroupSelf` (`idGroupSelf`),
  KEY `idGroupOwned` (`idGroupOwned`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_users`
--

LOCK TABLES `history_users` WRITE;
ALTER TABLE `history_users` DISABLE KEYS;
ALTER TABLE `history_users` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_users_items`
--

DROP TABLE IF EXISTS `history_users_items`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_users_items` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `idUser` bigint(20) NOT NULL,
  `idItem` bigint(20) NOT NULL,
  `idAttemptActive` bigint(20) DEFAULT NULL,
  `iScore` float NOT NULL DEFAULT '0',
  `iScoreComputed` float NOT NULL DEFAULT '0',
  `iScoreReeval` float DEFAULT '0',
  `iScoreDiffManual` float NOT NULL DEFAULT '0',
  `sScoreDiffComment` varchar(200) DEFAULT NULL,
  `nbSubmissionsAttempts` int(11) NOT NULL,
  `nbTasksTried` int(11) NOT NULL,
  `nbTasksSolved` int(11) NOT NULL DEFAULT '0',
  `nbChildrenValidated` int(11) NOT NULL,
  `bValidated` int(11) NOT NULL,
  `bFinished` int(11) NOT NULL,
  `bKeyObtained` tinyint(1) NOT NULL DEFAULT '0',
  `nbTasksWithHelp` int(11) NOT NULL,
  `sHintsRequested` mediumtext,
  `nbHintsCached` int(11) NOT NULL,
  `nbCorrectionsRead` int(11) NOT NULL,
  `iPrecision` int(11) NOT NULL,
  `iAutonomy` int(11) NOT NULL,
  `sStartDate` datetime DEFAULT NULL,
  `sValidationDate` datetime DEFAULT NULL,
  `sFinishDate` datetime DEFAULT NULL,
  `sLastActivityDate` datetime DEFAULT NULL,
  `sThreadStartDate` datetime DEFAULT NULL,
  `sBestAnswerDate` datetime DEFAULT NULL,
  `sLastAnswerDate` datetime DEFAULT NULL,
  `sLastHintDate` datetime DEFAULT NULL,
  `sAdditionalTime` time DEFAULT NULL,
  `sContestStartDate` datetime DEFAULT NULL,
  `bRanked` tinyint(1) NOT NULL,
  `sAllLangProg` varchar(200) DEFAULT NULL,
  `sState` mediumtext,
  `sAnswer` mediumtext,
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  `bPlatformDataRemoved` tinyint(4) NOT NULL DEFAULT '0',
  PRIMARY KEY (`historyID`),
  KEY `ID` (`ID`),
  KEY `iVersion` (`iVersion`),
  KEY `itemUser` (`idItem`,`idUser`),
  KEY `iNextVersion` (`iNextVersion`),
  KEY `bDeleted` (`bDeleted`),
  KEY `idItem` (`idItem`),
  KEY `idUser` (`idUser`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_users_items`
--

LOCK TABLES `history_users_items` WRITE;
ALTER TABLE `history_users_items` DISABLE KEYS;
ALTER TABLE `history_users_items` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `history_users_threads`
--

DROP TABLE IF EXISTS `history_users_threads`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `history_users_threads` (
  `history_ID` int(11) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `idUser` bigint(20) NOT NULL,
  `idThread` bigint(20) NOT NULL,
  `sLastReadDate` datetime DEFAULT NULL,
  `bParticipated` tinyint(1) NOT NULL DEFAULT '0',
  `sLastWriteDate` datetime DEFAULT NULL,
  `bStarred` tinyint(1) DEFAULT NULL,
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`history_ID`),
  KEY `userThread` (`idUser`,`idThread`),
  KEY `user` (`idUser`),
  KEY `iVersion` (`iVersion`),
  KEY `iNextVersion` (`iNextVersion`),
  KEY `bDeleted` (`bDeleted`),
  KEY `ID` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `history_users_threads`
--

LOCK TABLES `history_users_threads` WRITE;
ALTER TABLE `history_users_threads` DISABLE KEYS;
ALTER TABLE `history_users_threads` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `items`
--

DROP TABLE IF EXISTS `items`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `items` (
  `ID` bigint(20) NOT NULL,
  `sUrl` varchar(200) DEFAULT NULL,
  `sOptions` TEXT NOT NULL,
  `idPlatform` int(11) DEFAULT NULL,
  `sTextId` varchar(200) DEFAULT NULL,
  `sRepositoryPath` text,
  `sType` enum('Root','Category','Chapter','Task','Course') NOT NULL,
  `bTitleBarVisible` tinyint(3) unsigned NOT NULL DEFAULT '1',
  `bTransparentFolder` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `bDisplayDetailsInParent` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT 'when true, display a large icon, the subtitle, and more within the parent chapter',
  `bCustomChapter` tinyint(3) unsigned DEFAULT '0' COMMENT 'true if this is a chapter where users can add their own content. access to this chapter will not be propagated to its children',
  `bDisplayChildrenAsTabs` tinyint(3) unsigned NOT NULL DEFAULT '0',
  `bUsesAPI` tinyint(1) NOT NULL DEFAULT '1',
  `bReadOnly` tinyint(1) NOT NULL DEFAULT '0',
  `sFullScreen` enum('forceYes','forceNo','default','') NOT NULL DEFAULT 'default',
  `bShowDifficulty` tinyint(1) NOT NULL DEFAULT '0',
  `bShowSource` tinyint(1) NOT NULL DEFAULT '0',
  `bHintsAllowed` tinyint(1) NOT NULL DEFAULT '0',
  `bFixedRanks` tinyint(1) NOT NULL DEFAULT '0',
  `sValidationType` enum('None','All','AllButOne','Categories','One','Manual') NOT NULL DEFAULT 'All',
  `iValidationMin` int(11) DEFAULT NULL,
  `sPreparationState` enum('NotReady','Reviewing','Ready') NOT NULL DEFAULT 'NotReady',
  `idItemUnlocked` text,
  `iScoreMinUnlock` int(11) NOT NULL DEFAULT '100',
  `sSupportedLangProg` varchar(200) DEFAULT NULL,
  `idDefaultLanguage` bigint(20) DEFAULT '1',
  `sTeamMode` enum('All','Half','One','None') DEFAULT NULL,
  `bTeamsEditable` tinyint(1) NOT NULL,
  `idTeamInGroup` bigint(20) DEFAULT NULL,
  `iTeamMaxMembers` int(11) NOT NULL DEFAULT '0',
  `bHasAttempts` tinyint(1) NOT NULL DEFAULT '0',
  `sAccessOpenDate` datetime DEFAULT NULL,
  `sDuration` time DEFAULT NULL,
  `sEndContestDate` datetime DEFAULT NULL,
  `bShowUserInfos` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'always show user infos in title bar of all descendants',
  `sContestPhase` enum('Running','Analysis','Closed') NOT NULL,
  `iLevel` int(11) DEFAULT NULL,
  `bNoScore` tinyint(1) NOT NULL,
  `groupCodeEnter` tinyint(1) DEFAULT '0' COMMENT 'Offer users to enter through a group code',
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `iVersion` (`iVersion`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `items`
--

LOCK TABLES `items` WRITE;
ALTER TABLE `items` DISABLE KEYS;
ALTER TABLE `items` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_items` BEFORE INSERT ON `items` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; SELECT platforms.ID INTO @platformID FROM platforms WHERE NEW.sUrl REGEXP platforms.sRegexp ORDER BY platforms.iPriority DESC LIMIT 1 ; SET NEW.idPlatform=@platformID ; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_items` AFTER INSERT ON `items` FOR EACH ROW BEGIN INSERT INTO `history_items` (`ID`,`iVersion`,`sUrl`,`sOptions`,`idPlatform`,`sTextId`,`sRepositoryPath`,`sType`,`bUsesAPI`,`bReadOnly`,`sFullScreen`,`bShowDifficulty`,`bShowSource`,`bHintsAllowed`,`bFixedRanks`,`sValidationType`,`iValidationMin`,`sPreparationState`,`idItemUnlocked`,`iScoreMinUnlock`,`sSupportedLangProg`,`idDefaultLanguage`,`sTeamMode`,`bTeamsEditable`,`idTeamInGroup`,`iTeamMaxMembers`,`bHasAttempts`,`sAccessOpenDate`,`sDuration`,`sEndContestDate`,`bShowUserInfos`,`sContestPhase`,`iLevel`,`bNoScore`,`bTitleBarVisible`,`bTransparentFolder`,`bDisplayDetailsInParent`,`bDisplayChildrenAsTabs`,`bCustomChapter`,`groupCodeEnter`) VALUES (NEW.`ID`,@curVersion,NEW.`sUrl`,NEW.`sOptions`,NEW.`idPlatform`,NEW.`sTextId`,NEW.`sRepositoryPath`,NEW.`sType`,NEW.`bUsesAPI`,NEW.`bReadOnly`,NEW.`sFullScreen`,NEW.`bShowDifficulty`,NEW.`bShowSource`,NEW.`bHintsAllowed`,NEW.`bFixedRanks`,NEW.`sValidationType`,NEW.`iValidationMin`,NEW.`sPreparationState`,NEW.`idItemUnlocked`,NEW.`iScoreMinUnlock`,NEW.`sSupportedLangProg`,NEW.`idDefaultLanguage`,NEW.`sTeamMode`,NEW.`bTeamsEditable`,NEW.`idTeamInGroup`,NEW.`iTeamMaxMembers`,NEW.`bHasAttempts`,NEW.`sAccessOpenDate`,NEW.`sDuration`,NEW.`sEndContestDate`,NEW.`bShowUserInfos`,NEW.`sContestPhase`,NEW.`iLevel`,NEW.`bNoScore`,NEW.`bTitleBarVisible`,NEW.`bTransparentFolder`,NEW.`bDisplayDetailsInParent`,NEW.`bDisplayChildrenAsTabs`,NEW.`bCustomChapter`,NEW.`groupCodeEnter`); INSERT IGNORE INTO `items_propagate` (`ID`, `sAncestorsComputationState`) VALUES (NEW.`ID`, 'todo') ; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_items` BEFORE UPDATE ON `items` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`sUrl` <=> NEW.`sUrl` AND OLD.`sOptions` <=> NEW.`sOptions` AND OLD.`idPlatform` <=> NEW.`idPlatform` AND OLD.`sTextId` <=> NEW.`sTextId` AND OLD.`sRepositoryPath` <=> NEW.`sRepositoryPath` AND OLD.`sType` <=> NEW.`sType` AND OLD.`bUsesAPI` <=> NEW.`bUsesAPI` AND OLD.`bReadOnly` <=> NEW.`bReadOnly` AND OLD.`sFullScreen` <=> NEW.`sFullScreen` AND OLD.`bShowDifficulty` <=> NEW.`bShowDifficulty` AND OLD.`bShowSource` <=> NEW.`bShowSource` AND OLD.`bHintsAllowed` <=> NEW.`bHintsAllowed` AND OLD.`bFixedRanks` <=> NEW.`bFixedRanks` AND OLD.`sValidationType` <=> NEW.`sValidationType` AND OLD.`iValidationMin` <=> NEW.`iValidationMin` AND OLD.`sPreparationState` <=> NEW.`sPreparationState` AND OLD.`idItemUnlocked` <=> NEW.`idItemUnlocked` AND OLD.`iScoreMinUnlock` <=> NEW.`iScoreMinUnlock` AND OLD.`sSupportedLangProg` <=> NEW.`sSupportedLangProg` AND OLD.`idDefaultLanguage` <=> NEW.`idDefaultLanguage` AND OLD.`sTeamMode` <=> NEW.`sTeamMode` AND OLD.`bTeamsEditable` <=> NEW.`bTeamsEditable` AND OLD.`idTeamInGroup` <=> NEW.`idTeamInGroup` AND OLD.`iTeamMaxMembers` <=> NEW.`iTeamMaxMembers` AND OLD.`bHasAttempts` <=> NEW.`bHasAttempts` AND OLD.`sAccessOpenDate` <=> NEW.`sAccessOpenDate` AND OLD.`sDuration` <=> NEW.`sDuration` AND OLD.`sEndContestDate` <=> NEW.`sEndContestDate` AND OLD.`bShowUserInfos` <=> NEW.`bShowUserInfos` AND OLD.`sContestPhase` <=> NEW.`sContestPhase` AND OLD.`iLevel` <=> NEW.`iLevel` AND OLD.`bNoScore` <=> NEW.`bNoScore` AND OLD.`bTitleBarVisible` <=> NEW.`bTitleBarVisible` AND OLD.`bTransparentFolder` <=> NEW.`bTransparentFolder` AND OLD.`bDisplayDetailsInParent` <=> NEW.`bDisplayDetailsInParent` AND OLD.`bDisplayChildrenAsTabs` <=> NEW.`bDisplayChildrenAsTabs` AND OLD.`bCustomChapter` <=> NEW.`bCustomChapter` AND OLD.`groupCodeEnter` <=> NEW.`groupCodeEnter`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_items` (`ID`,`iVersion`,`sUrl`,`sOptions`,`idPlatform`,`sTextId`,`sRepositoryPath`,`sType`,`bUsesAPI`,`bReadOnly`,`sFullScreen`,`bShowDifficulty`,`bShowSource`,`bHintsAllowed`,`bFixedRanks`,`sValidationType`,`iValidationMin`,`sPreparationState`,`idItemUnlocked`,`iScoreMinUnlock`,`sSupportedLangProg`,`idDefaultLanguage`,`sTeamMode`,`bTeamsEditable`,`idTeamInGroup`,`iTeamMaxMembers`,`bHasAttempts`,`sAccessOpenDate`,`sDuration`,`sEndContestDate`,`bShowUserInfos`,`sContestPhase`,`iLevel`,`bNoScore`,`bTitleBarVisible`,`bTransparentFolder`,`bDisplayDetailsInParent`,`bDisplayChildrenAsTabs`,`bCustomChapter`,`groupCodeEnter`)       VALUES (NEW.`ID`,@curVersion,NEW.`sUrl`,NEW.`sOptions`,NEW.`idPlatform`,NEW.`sTextId`,NEW.`sRepositoryPath`,NEW.`sType`,NEW.`bUsesAPI`,NEW.`bReadOnly`,NEW.`sFullScreen`,NEW.`bShowDifficulty`,NEW.`bShowSource`,NEW.`bHintsAllowed`,NEW.`bFixedRanks`,NEW.`sValidationType`,NEW.`iValidationMin`,NEW.`sPreparationState`,NEW.`idItemUnlocked`,NEW.`iScoreMinUnlock`,NEW.`sSupportedLangProg`,NEW.`idDefaultLanguage`,NEW.`sTeamMode`,NEW.`bTeamsEditable`,NEW.`idTeamInGroup`,NEW.`iTeamMaxMembers`,NEW.`bHasAttempts`,NEW.`sAccessOpenDate`,NEW.`sDuration`,NEW.`sEndContestDate`,NEW.`bShowUserInfos`,NEW.`sContestPhase`,NEW.`iLevel`,NEW.`bNoScore`,NEW.`bTitleBarVisible`,NEW.`bTransparentFolder`,NEW.`bDisplayDetailsInParent`,NEW.`bDisplayChildrenAsTabs`,NEW.`bCustomChapter`,NEW.`groupCodeEnter`) ; END IF; SELECT platforms.ID INTO @platformID FROM platforms WHERE NEW.sUrl REGEXP platforms.sRegexp ORDER BY platforms.iPriority DESC LIMIT 1 ; SET NEW.idPlatform=@platformID ; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_items` BEFORE DELETE ON `items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_items` (`ID`,`iVersion`,`sUrl`,`sOptions`,`idPlatform`,`sTextId`,`sRepositoryPath`,`sType`,`bUsesAPI`,`bReadOnly`,`sFullScreen`,`bShowDifficulty`,`bShowSource`,`bHintsAllowed`,`bFixedRanks`,`sValidationType`,`iValidationMin`,`sPreparationState`,`idItemUnlocked`,`iScoreMinUnlock`,`sSupportedLangProg`,`idDefaultLanguage`,`sTeamMode`,`bTeamsEditable`,`idTeamInGroup`,`iTeamMaxMembers`,`bHasAttempts`,`sAccessOpenDate`,`sDuration`,`sEndContestDate`,`bShowUserInfos`,`sContestPhase`,`iLevel`,`bNoScore`,`bTitleBarVisible`,`bTransparentFolder`,`bDisplayDetailsInParent`,`bDisplayChildrenAsTabs`,`bCustomChapter`,`groupCodeEnter`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`sUrl`,OLD.`sOptions`,OLD.`idPlatform`,OLD.`sTextId`,OLD.`sRepositoryPath`,OLD.`sType`,OLD.`bUsesAPI`,OLD.`bReadOnly`,OLD.`sFullScreen`,OLD.`bShowDifficulty`,OLD.`bShowSource`,OLD.`bHintsAllowed`,OLD.`bFixedRanks`,OLD.`sValidationType`,OLD.`iValidationMin`,OLD.`sPreparationState`,OLD.`idItemUnlocked`,OLD.`iScoreMinUnlock`,OLD.`sSupportedLangProg`,OLD.`idDefaultLanguage`,OLD.`sTeamMode`,OLD.`bTeamsEditable`,OLD.`idTeamInGroup`,OLD.`iTeamMaxMembers`,OLD.`bHasAttempts`,OLD.`sAccessOpenDate`,OLD.`sDuration`,OLD.`sEndContestDate`,OLD.`bShowUserInfos`,OLD.`sContestPhase`,OLD.`iLevel`,OLD.`bNoScore`,OLD.`bTitleBarVisible`,OLD.`bTransparentFolder`,OLD.`bDisplayDetailsInParent`,OLD.`bDisplayChildrenAsTabs`,OLD.`bCustomChapter`,OLD.`groupCodeEnter`, 1); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_delete_items` AFTER DELETE ON `items` FOR EACH ROW BEGIN DELETE FROM items_propagate where ID = OLD.ID ; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `items_ancestors`
--

DROP TABLE IF EXISTS `items_ancestors`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `items_ancestors` (
  `ID` bigint(20) NOT NULL AUTO_INCREMENT,
  `idItemAncestor` bigint(20) NOT NULL,
  `idItemChild` bigint(20) NOT NULL,
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `idItemAncestor` (`idItemAncestor`,`idItemChild`),
  KEY `idItemAncestortor` (`idItemAncestor`),
  KEY `idItemChild` (`idItemChild`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `items_ancestors`
--

LOCK TABLES `items_ancestors` WRITE;
ALTER TABLE `items_ancestors` DISABLE KEYS;
ALTER TABLE `items_ancestors` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_items_ancestors` BEFORE INSERT ON `items_ancestors` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_items_ancestors` AFTER INSERT ON `items_ancestors` FOR EACH ROW BEGIN INSERT INTO `history_items_ancestors` (`ID`,`iVersion`,`idItemAncestor`,`idItemChild`) VALUES (NEW.`ID`,@curVersion,NEW.`idItemAncestor`,NEW.`idItemChild`); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_items_ancestors` BEFORE UPDATE ON `items_ancestors` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idItemAncestor` <=> NEW.`idItemAncestor` AND OLD.`idItemChild` <=> NEW.`idItemChild`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_items_ancestors` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_items_ancestors` (`ID`,`iVersion`,`idItemAncestor`,`idItemChild`)       VALUES (NEW.`ID`,@curVersion,NEW.`idItemAncestor`,NEW.`idItemChild`) ; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_items_ancestors` BEFORE DELETE ON `items_ancestors` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items_ancestors` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_items_ancestors` (`ID`,`iVersion`,`idItemAncestor`,`idItemChild`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idItemAncestor`,OLD.`idItemChild`, 1); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `items_items`
--

DROP TABLE IF EXISTS `items_items`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `items_items` (
  `ID` bigint(20) NOT NULL,
  `idItemParent` bigint(20) NOT NULL,
  `idItemChild` bigint(20) NOT NULL,
  `iChildOrder` int(11) NOT NULL,
  `sCategory` enum('Undefined','Discovery','Application','Validation','Challenge') NOT NULL DEFAULT 'Undefined',
  `bAlwaysVisible` tinyint(1) NOT NULL DEFAULT '0',
  `bAccessRestricted` tinyint(1) NOT NULL DEFAULT '1',
  `iDifficulty` int(11) NOT NULL,
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `idItemParent` (`idItemParent`),
  KEY `idItemChild` (`idItemChild`),
  KEY `iVersion` (`iVersion`),
  KEY `parentChild` (`idItemParent`,`idItemChild`),
  KEY `parentVersion` (`idItemParent`,`iVersion`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `items_items`
--

LOCK TABLES `items_items` WRITE;
ALTER TABLE `items_items` DISABLE KEYS;
ALTER TABLE `items_items` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_items_items` BEFORE INSERT ON `items_items` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; INSERT IGNORE INTO `items_propagate` (ID, sAncestorsComputationState) VALUES (NEW.idItemChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo' ; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_items_items` AFTER INSERT ON `items_items` FOR EACH ROW BEGIN INSERT INTO `history_items_items` (`ID`,`iVersion`,`idItemParent`,`idItemChild`,`iChildOrder`,`sCategory`,`bAccessRestricted`,`bAlwaysVisible`,`iDifficulty`) VALUES (NEW.`ID`,@curVersion,NEW.`idItemParent`,NEW.`idItemChild`,NEW.`iChildOrder`,NEW.`sCategory`,NEW.`bAccessRestricted`,NEW.`bAlwaysVisible`,NEW.`iDifficulty`); INSERT IGNORE INTO `groups_items_propagate` SELECT `ID`, 'children' as `sPropagateAccess` FROM `groups_items` WHERE `groups_items`.`idItem` = NEW.`idItemParent` ON DUPLICATE KEY UPDATE sPropagateAccess='children' ; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_items_items` BEFORE UPDATE ON `items_items` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idItemParent` <=> NEW.`idItemParent` AND OLD.`idItemChild` <=> NEW.`idItemChild` AND OLD.`iChildOrder` <=> NEW.`iChildOrder` AND OLD.`sCategory` <=> NEW.`sCategory` AND OLD.`bAccessRestricted` <=> NEW.`bAccessRestricted` AND OLD.`bAlwaysVisible` <=> NEW.`bAlwaysVisible` AND OLD.`iDifficulty` <=> NEW.`iDifficulty`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_items_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_items_items` (`ID`,`iVersion`,`idItemParent`,`idItemChild`,`iChildOrder`,`sCategory`,`bAccessRestricted`,`bAlwaysVisible`,`iDifficulty`)       VALUES (NEW.`ID`,@curVersion,NEW.`idItemParent`,NEW.`idItemChild`,NEW.`iChildOrder`,NEW.`sCategory`,NEW.`bAccessRestricted`,NEW.`bAlwaysVisible`,NEW.`iDifficulty`) ; END IF; IF (OLD.idItemChild != NEW.idItemChild OR OLD.idItemParent != NEW.idItemParent) THEN INSERT IGNORE INTO `items_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idItemChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `items_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idItemParent, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `items_propagate` (`ID`, `sAncestorsComputationState`) (SELECT `items_ancestors`.`idItemChild`, 'todo' FROM `items_ancestors` WHERE `items_ancestors`.`idItemAncestor` = OLD.`idItemChild`) ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; DELETE `items_ancestors` from `items_ancestors` WHERE `items_ancestors`.`idItemChild` = OLD.`idItemChild` and `items_ancestors`.`idItemAncestor` = OLD.`idItemParent`;DELETE `bridges` FROM `items_ancestors` `child_descendants` JOIN `items_ancestors` `parent_ancestors` JOIN `items_ancestors` `bridges` ON (`bridges`.`idItemAncestor` = `parent_ancestors`.`idItemAncestor` AND `bridges`.`idItemChild` = `child_descendants`.`idItemChild`) WHERE `parent_ancestors`.`idItemChild` = OLD.`idItemParent` AND `child_descendants`.`idItemAncestor` = OLD.`idItemChild`; DELETE `child_ancestors` FROM `items_ancestors` `child_ancestors` JOIN  `items_ancestors` `parent_ancestors` ON (`child_ancestors`.`idItemChild` = OLD.`idItemChild` AND `child_ancestors`.`idItemAncestor` = `parent_ancestors`.`idItemAncestor`) WHERE `parent_ancestors`.`idItemChild` = OLD.`idItemParent`; DELETE `parent_ancestors` FROM `items_ancestors` `parent_ancestors` JOIN  `items_ancestors` `child_ancestors` ON (`parent_ancestors`.`idItemAncestor` = OLD.`idItemParent` AND `child_ancestors`.`idItemChild` = `parent_ancestors`.`idItemChild`) WHERE `child_ancestors`.`idItemAncestor` = OLD.`idItemChild`  ; END IF; IF (OLD.idItemChild != NEW.idItemChild OR OLD.idItemParent != NEW.idItemParent) THEN INSERT IGNORE INTO `items_propagate` (ID, sAncestorsComputationState) VALUES (NEW.idItemChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'  ; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_update_items_items` AFTER UPDATE ON `items_items` FOR EACH ROW BEGIN INSERT IGNORE INTO `groups_items_propagate` SELECT `ID`, 'children' as `sPropagateAccess` FROM `groups_items` WHERE `groups_items`.`idItem` = NEW.`idItemParent` OR `groups_items`.`idItem` = OLD.`idItemParent` ON DUPLICATE KEY UPDATE sPropagateAccess='children' ; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_items_items` BEFORE DELETE ON `items_items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_items_items` (`ID`,`iVersion`,`idItemParent`,`idItemChild`,`iChildOrder`,`sCategory`,`bAccessRestricted`,`bAlwaysVisible`,`iDifficulty`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idItemParent`,OLD.`idItemChild`,OLD.`iChildOrder`,OLD.`sCategory`,OLD.`bAccessRestricted`,OLD.`bAlwaysVisible`,OLD.`iDifficulty`, 1); INSERT IGNORE INTO `items_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idItemChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `items_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idItemParent, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `items_propagate` (`ID`, `sAncestorsComputationState`) (SELECT `items_ancestors`.`idItemChild`, 'todo' FROM `items_ancestors` WHERE `items_ancestors`.`idItemAncestor` = OLD.`idItemChild`) ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; DELETE `items_ancestors` from `items_ancestors` WHERE `items_ancestors`.`idItemChild` = OLD.`idItemChild` and `items_ancestors`.`idItemAncestor` = OLD.`idItemParent`;DELETE `bridges` FROM `items_ancestors` `child_descendants` JOIN `items_ancestors` `parent_ancestors` JOIN `items_ancestors` `bridges` ON (`bridges`.`idItemAncestor` = `parent_ancestors`.`idItemAncestor` AND `bridges`.`idItemChild` = `child_descendants`.`idItemChild`) WHERE `parent_ancestors`.`idItemChild` = OLD.`idItemParent` AND `child_descendants`.`idItemAncestor` = OLD.`idItemChild`; DELETE `child_ancestors` FROM `items_ancestors` `child_ancestors` JOIN  `items_ancestors` `parent_ancestors` ON (`child_ancestors`.`idItemChild` = OLD.`idItemChild` AND `child_ancestors`.`idItemAncestor` = `parent_ancestors`.`idItemAncestor`) WHERE `parent_ancestors`.`idItemChild` = OLD.`idItemParent`; DELETE `parent_ancestors` FROM `items_ancestors` `parent_ancestors` JOIN  `items_ancestors` `child_ancestors` ON (`parent_ancestors`.`idItemAncestor` = OLD.`idItemParent` AND `child_ancestors`.`idItemChild` = `parent_ancestors`.`idItemChild`) WHERE `child_ancestors`.`idItemAncestor` = OLD.`idItemChild` ; INSERT IGNORE INTO `groups_items_propagate` SELECT `ID`, 'children' as `sPropagateAccess` FROM `groups_items` WHERE `groups_items`.`idItem` = OLD.`idItemParent` ON DUPLICATE KEY UPDATE sPropagateAccess='children' ; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `items_propagate`
--

DROP TABLE IF EXISTS `items_propagate`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `items_propagate` (
  `ID` bigint(20) NOT NULL,
  `sAncestorsComputationState` enum('todo','done','processing','') NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `sAncestorsComputationDate` (`sAncestorsComputationState`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `items_propagate`
--

LOCK TABLES `items_propagate` WRITE;
ALTER TABLE `items_propagate` DISABLE KEYS;
ALTER TABLE `items_propagate` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `items_strings`
--

DROP TABLE IF EXISTS `items_strings`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `items_strings` (
  `ID` bigint(20) NOT NULL,
  `idItem` bigint(20) NOT NULL,
  `idLanguage` bigint(20) NOT NULL,
  `sTranslator` varchar(100) DEFAULT NULL,
  `sTitle` varchar(200) DEFAULT NULL,
  `sImageUrl` text,
  `sSubtitle` varchar(200) DEFAULT NULL,
  `sDescription` text,
  `sEduComment` text,
  `sRankingComment` text,
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `idItem` (`idItem`,`idLanguage`),
  KEY `iVersion` (`iVersion`),
  KEY `idItemAlone` (`idItem`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `items_strings`
--

LOCK TABLES `items_strings` WRITE;
ALTER TABLE `items_strings` DISABLE KEYS;
ALTER TABLE `items_strings` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_items_strings` BEFORE INSERT ON `items_strings` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_items_strings` AFTER INSERT ON `items_strings` FOR EACH ROW BEGIN INSERT INTO `history_items_strings` (`ID`,`iVersion`,`idItem`,`idLanguage`,`sTranslator`,`sTitle`,`sImageUrl`,`sSubtitle`,`sDescription`,`sEduComment`,`sRankingComment`) VALUES (NEW.`ID`,@curVersion,NEW.`idItem`,NEW.`idLanguage`,NEW.`sTranslator`,NEW.`sTitle`,NEW.`sImageUrl`,NEW.`sSubtitle`,NEW.`sDescription`,NEW.`sEduComment`,NEW.`sRankingComment`); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_items_strings` BEFORE UPDATE ON `items_strings` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idItem` <=> NEW.`idItem` AND OLD.`idLanguage` <=> NEW.`idLanguage` AND OLD.`sTranslator` <=> NEW.`sTranslator` AND OLD.`sTitle` <=> NEW.`sTitle` AND OLD.`sImageUrl` <=> NEW.`sImageUrl` AND OLD.`sSubtitle` <=> NEW.`sSubtitle` AND OLD.`sDescription` <=> NEW.`sDescription` AND OLD.`sEduComment` <=> NEW.`sEduComment` AND OLD.`sRankingComment` <=> NEW.`sRankingComment`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_items_strings` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_items_strings` (`ID`,`iVersion`,`idItem`,`idLanguage`,`sTranslator`,`sTitle`,`sImageUrl`,`sSubtitle`,`sDescription`,`sEduComment`,`sRankingComment`)       VALUES (NEW.`ID`,@curVersion,NEW.`idItem`,NEW.`idLanguage`,NEW.`sTranslator`,NEW.`sTitle`,NEW.`sImageUrl`,NEW.`sSubtitle`,NEW.`sDescription`,NEW.`sEduComment`,NEW.`sRankingComment`) ; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_items_strings` BEFORE DELETE ON `items_strings` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_items_strings` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_items_strings` (`ID`,`iVersion`,`idItem`,`idLanguage`,`sTranslator`,`sTitle`,`sImageUrl`,`sSubtitle`,`sDescription`,`sEduComment`,`sRankingComment`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idItem`,OLD.`idLanguage`,OLD.`sTranslator`,OLD.`sTitle`,OLD.`sImageUrl`,OLD.`sSubtitle`,OLD.`sDescription`,OLD.`sEduComment`,OLD.`sRankingComment`, 1); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `languages`
--

DROP TABLE IF EXISTS `languages`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `languages` (
  `ID` bigint(20) NOT NULL,
  `sName` varchar(100) NOT NULL DEFAULT '',
  `sCode` varchar(2) NOT NULL DEFAULT '',
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `iVersion` (`iVersion`),
  KEY `sCode` (`sCode`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `languages`
--

LOCK TABLES `languages` WRITE;
ALTER TABLE `languages` DISABLE KEYS;
ALTER TABLE `languages` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_languages` BEFORE INSERT ON `languages` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_languages` AFTER INSERT ON `languages` FOR EACH ROW BEGIN INSERT INTO `history_languages` (`ID`,`iVersion`,`sName`,`sCode`) VALUES (NEW.`ID`,@curVersion,NEW.`sName`,NEW.`sCode`); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_languages` BEFORE UPDATE ON `languages` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`sName` <=> NEW.`sName` AND OLD.`sCode` <=> NEW.`sCode`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_languages` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_languages` (`ID`,`iVersion`,`sName`,`sCode`)       VALUES (NEW.`ID`,@curVersion,NEW.`sName`,NEW.`sCode`) ; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_languages` BEFORE DELETE ON `languages` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_languages` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_languages` (`ID`,`iVersion`,`sName`,`sCode`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`sName`,OLD.`sCode`, 1); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `messages`
--

DROP TABLE IF EXISTS `messages`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `messages` (
  `ID` bigint(20) NOT NULL,
  `idThread` bigint(20) DEFAULT NULL,
  `idUser` bigint(20) DEFAULT NULL,
  `sSubmissionDate` datetime DEFAULT NULL,
  `bPublished` tinyint(1) NOT NULL DEFAULT '1',
  `sTitle` varchar(200) DEFAULT '',
  `sBody` varchar(2000) DEFAULT '',
  `bTrainersOnly` tinyint(1) NOT NULL DEFAULT '0',
  `bArchived` tinyint(1) DEFAULT '0',
  `bPersistant` tinyint(1) DEFAULT NULL,
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `iVersion` (`iVersion`),
  KEY `idThread` (`idThread`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `messages`
--

LOCK TABLES `messages` WRITE;
ALTER TABLE `messages` DISABLE KEYS;
ALTER TABLE `messages` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_messages` BEFORE INSERT ON `messages` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_messages` AFTER INSERT ON `messages` FOR EACH ROW BEGIN INSERT INTO `history_messages` (`ID`,`iVersion`,`idThread`,`idUser`,`sSubmissionDate`,`bPublished`,`sTitle`,`sBody`,`bTrainersOnly`,`bArchived`,`bPersistant`) VALUES (NEW.`ID`,@curVersion,NEW.`idThread`,NEW.`idUser`,NEW.`sSubmissionDate`,NEW.`bPublished`,NEW.`sTitle`,NEW.`sBody`,NEW.`bTrainersOnly`,NEW.`bArchived`,NEW.`bPersistant`); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_messages` BEFORE UPDATE ON `messages` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idThread` <=> NEW.`idThread` AND OLD.`idUser` <=> NEW.`idUser` AND OLD.`sSubmissionDate` <=> NEW.`sSubmissionDate` AND OLD.`bPublished` <=> NEW.`bPublished` AND OLD.`sTitle` <=> NEW.`sTitle` AND OLD.`sBody` <=> NEW.`sBody` AND OLD.`bTrainersOnly` <=> NEW.`bTrainersOnly` AND OLD.`bArchived` <=> NEW.`bArchived` AND OLD.`bPersistant` <=> NEW.`bPersistant`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_messages` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_messages` (`ID`,`iVersion`,`idThread`,`idUser`,`sSubmissionDate`,`bPublished`,`sTitle`,`sBody`,`bTrainersOnly`,`bArchived`,`bPersistant`)       VALUES (NEW.`ID`,@curVersion,NEW.`idThread`,NEW.`idUser`,NEW.`sSubmissionDate`,NEW.`bPublished`,NEW.`sTitle`,NEW.`sBody`,NEW.`bTrainersOnly`,NEW.`bArchived`,NEW.`bPersistant`) ; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_messages` BEFORE DELETE ON `messages` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_messages` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_messages` (`ID`,`iVersion`,`idThread`,`idUser`,`sSubmissionDate`,`bPublished`,`sTitle`,`sBody`,`bTrainersOnly`,`bArchived`,`bPersistant`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idThread`,OLD.`idUser`,OLD.`sSubmissionDate`,OLD.`bPublished`,OLD.`sTitle`,OLD.`sBody`,OLD.`bTrainersOnly`,OLD.`bArchived`,OLD.`bPersistant`, 1); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `platforms`
--

DROP TABLE IF EXISTS `platforms`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `platforms` (
  `ID` int(11) NOT NULL AUTO_INCREMENT,
  `sName` varchar(50) NOT NULL DEFAULT '',
  `sBaseUrl` varchar(200) DEFAULT NULL,
  `sPublicKey` varchar(512) NOT NULL DEFAULT '',
  `bUsesTokens` tinyint(1) NOT NULL,
  `sRegexp` text,
  `iPriority` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `platforms`
--

LOCK TABLES `platforms` WRITE;
ALTER TABLE `platforms` DISABLE KEYS;
ALTER TABLE `platforms` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `schema_revision`
--

DROP TABLE IF EXISTS `schema_revision`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `schema_revision` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `executed_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `file` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=105 DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `schema_revision`
--

LOCK TABLES `schema_revision` WRITE;
ALTER TABLE `schema_revision` DISABLE KEYS;
INSERT INTO `schema_revision` VALUES (1,'2018-10-24 06:11:11','1.0/revision-002/synchro_versions.sql'),(2,'2018-10-24 06:11:11','1.0/revision-003/history_tables.sql'),(3,'2018-10-24 06:11:11','1.0/revision-004/historyID_autoincrement.sql'),(4,'2018-10-24 06:11:11','1.0/revision-008/root_category.sql'),(5,'2018-10-24 06:11:11','1.0/revision-016/fix_item_sType.sql'),(6,'2018-10-24 06:11:11','1.0/revision-016/groups_tables.sql'),(7,'2018-10-24 06:11:11','1.0/revision-017/missing_fields.sql'),(8,'2018-10-24 06:11:11','1.0/revision-018/fix_history_groups.sql'),(9,'2018-10-24 06:11:11','1.0/revision-020/groups_sType.sql'),(10,'2018-10-24 06:11:11','1.0/revision-028/access_modes.sql'),(11,'2018-10-24 06:11:11','1.0/revision-029/access_solutions.sql'),(12,'2018-10-24 06:11:11','1.0/revision-029/validationType.sql'),(13,'2018-10-24 06:11:11','1.0/revision-031/drop_group_owners.sql'),(14,'2018-10-24 06:11:11','1.0/revision-031/users.sql'),(15,'2018-10-24 06:11:11','1.0/revision-032/isAdmin_fix.sql'),(16,'2018-10-24 06:11:11','1.0/revision-032/sNotify.sql'),(17,'2018-10-24 06:11:11','1.0/revision-032/user_groups.sql'),(18,'2018-10-24 06:11:11','1.0/revision-044/users_items.sql'),(19,'2018-10-24 06:11:11','1.0/revision-045/ancestors.sql'),(20,'2018-10-24 06:11:11','1.0/revision-055/index_ancestors.sql'),(21,'2018-10-24 06:11:11','1.0/revision-057/groups_items_access.sql'),(22,'2018-10-24 06:11:11','1.0/revision-060/items_ancestors_access.sql'),(23,'2018-10-24 06:11:11','1.0/revision-061/access_dates.sql'),(24,'2018-10-24 06:11:11','1.0/revision-066/nextVersion_null.sql'),(25,'2018-10-24 06:11:11','1.0/revision-139/completing_items_table.sql'),(26,'2018-10-24 06:11:11','1.0/revision-161/history_items.sql'),(27,'2018-10-24 06:11:11','1.0/revision-161/history_items_ancestors.sql'),(28,'2018-10-24 06:11:11','1.0/revision-212/user_item_index.sql'),(29,'2018-10-24 06:11:11','1.0/revision-227/groups_items_propagation.sql'),(30,'2018-10-24 06:11:11','1.0/revision-246/0-users_items_computations.sql'),(31,'2018-10-24 06:11:11','1.0/revision-246/items_groups_ancestors.sql'),(32,'2018-10-24 06:11:11','1.0/revision-246/optimizations.sql'),(33,'2018-10-24 06:11:11','1.0/revision-268/users_and_items.sql'),(34,'2018-10-24 06:11:11','1.0/revision-277/forum.sql'),(35,'2018-10-24 06:11:11','1.0/revision-321/manual_validation.sql'),(36,'2018-10-24 06:11:11','1.0/revision-321/platforms.sql'),(37,'2018-10-24 06:11:11','1.0/revision-321/users_answers.sql'),(38,'2018-10-24 06:11:11','1.0/revision-532/bugfixes.sql'),(39,'2018-10-24 06:11:11','1.0/revision-533/user_index.sql'),(40,'2018-10-24 06:11:11','1.0/revision-534/groups_items_index.sql'),(41,'2018-10-24 06:11:11','1.0/revision-536/index_item_item.sql'),(42,'2018-10-24 06:11:11','1.0/revision-537/default_propagation_users_items.sql'),(43,'2018-10-24 06:11:11','1.0/revision-538/task_url_index.sql'),(44,'2018-10-24 06:11:11','1.0/revision-539/item.bFullScreen.sql'),(45,'2018-10-24 06:11:11','1.0/revision-540/groups.sql'),(46,'2018-10-24 06:11:11','1.0/revision-541/items.sIconUrl.sql'),(47,'2018-10-24 06:11:11','1.0/revision-543/users.sRegistrationDate.sql'),(48,'2018-10-24 06:11:11','1.0/revision-544/limittedTimeContests.sql'),(49,'2018-10-24 06:11:11','1.0/revision-545/contest_adjustment.sql'),(50,'2018-10-24 06:11:11','1.0/revision-546/more_groups_groups_indexes.sql'),(51,'2018-10-24 06:11:11','1.0/revision-546/more_indexes.sql'),(52,'2018-10-24 06:11:11','1.0/revision-547/groups_defaults.sql'),(53,'2018-10-24 06:11:11','1.0/revision-548/items_bShowUserInfos.sql'),(54,'2018-10-24 06:11:11','1.0/revision-549/users_defaults.sql'),(55,'2018-10-24 06:11:32','1.1/revision-001/bugfix.sql'),(56,'2018-10-24 06:11:32','1.1/revision-002/platforms.sql'),(57,'2018-10-24 06:11:32','1.1/revision-003/drop_tmp.sql'),(58,'2018-10-24 06:11:33','1.1/revision-004/groups_stextid.sql'),(59,'2018-10-24 06:11:33','1.1/revision-004/unlocks.sql'),(60,'2018-10-24 06:11:33','1.1/revision-005/iScoreReeval.sql'),(61,'2018-10-24 06:11:33','1.1/revision-006/items_bReadOnly.sql'),(62,'2018-10-24 06:11:34','1.1/revision-007/fix_default_values.sql'),(63,'2018-10-24 06:11:34','1.1/revision-007/fix_history_bdeleted.sql'),(64,'2018-10-24 06:11:34','1.1/revision-007/fix_history_groups.sql'),(65,'2018-10-24 06:11:34','1.1/revision-007/fix_history_users.sql'),(66,'2018-10-24 06:11:34','1.1/revision-008/schema_revision.sql'),(67,'2018-10-24 06:11:34','1.1/revision-009/bFixedRanks.sql'),(68,'2018-10-24 06:11:34','1.1/revision-010/error_log.sql'),(69,'2018-10-24 06:11:35','1.1/revision-011/reducing_item_stype.sql'),(70,'2018-10-24 06:11:35','1.1/revision-012/fix_items_sDuration.sql'),(71,'2018-10-24 06:11:35','1.1/revision-012/users_items_sAnswer.sql'),(72,'2018-10-24 06:11:35','1.1/revision-013/items_idDefaultLanguage.sql'),(73,'2018-10-24 06:11:35','1.1/revision-014/default_values.sql'),(74,'2018-10-24 06:11:35','1.1/revision-014/default_values_bugfix.sql'),(75,'2018-10-24 06:11:35','1.1/revision-014/groups_fields.sql'),(76,'2018-10-24 06:11:35','1.1/revision-014/groups_login_prefixes.sql'),(77,'2018-10-24 06:11:36','1.1/revision-014/lm_prefix.sql'),(78,'2018-10-24 06:11:36','1.1/revision-015/groups_null_fileds.sql'),(79,'2018-10-24 06:11:36','1.1/revision-015/nulls.sql'),(80,'2018-10-24 06:11:36','1.1/revision-015/update_root_groups_textid.sql'),(81,'2018-10-24 06:11:36','1.1/revision-015/users.allowSubgroups.sql'),(82,'2018-10-24 06:11:36','1.1/revision-016/groupCodeEnter.sql'),(83,'2018-10-24 06:11:36','1.1/revision-016/items.sql'),(84,'2018-10-24 06:11:36','1.1/revision-016/text_varchar_bugfix.sql'),(85,'2018-10-24 06:11:37','1.1/revision-017/sHintsRequested.sql'),(86,'2018-10-24 06:11:37','1.1/revision-018/groups_add_teams.sql'),(87,'2018-10-24 06:11:37','1.1/revision-019/history_groups_add_teams.sql'),(88,'2018-10-24 06:11:37','1.1/revision-020/groups_add_teams.sql'),(89,'2018-10-24 06:11:37','1.1/revision-021/graduation_grade.sql'),(90,'2018-10-24 06:11:37','1.1/revision-021/groups_login_prefixes.sql'),(91,'2018-10-24 06:11:37','1.1/revision-021/user_creator_id.sql'),(92,'2018-10-24 06:11:38','1.1/revision-022/attempts.sql'),(93,'2018-10-24 06:11:38','1.1/revision-022/history_group_login_prefixes_index.sql'),(94,'2018-10-24 06:11:38','1.1/revision-023/indexes.sql'),(95,'2018-10-24 06:11:38','1.1/revision-024/indexes.sql'),(96,'2018-10-24 06:11:38','1.1/revision-024/remove_history_users_answers.sql'),(97,'2018-10-24 06:11:38','1.1/revision-025/bTeamsEditable.sql'),(98,'2018-10-24 06:11:38','1.1/revision-026/sBestAnswerDate.sql'),(99,'2018-10-24 06:11:38','1.1/revision-027/badges.sql'),(100,'2018-10-24 06:11:38','1.1/revision-028/lockUserDeletionDate.sql'),(101,'2018-10-24 06:11:39','1.1/revision-028/nulls.sql'),(102,'2018-10-24 06:11:39','1.1/revision-029/items_sRepositoryPath.sql'),(103,'2018-10-24 06:11:39','1.1/revision-029/platform_baseUrl.sql'),(104,'2018-10-24 06:11:39','1.1/revision-029/user_items_platform_data.sql'),(105,'2019-09-24 22:10:44','1.1/revision-030/items_strings_sImageUrl.sql'),(106,'2019-09-24 22:10:44','1.1/revision-031/items_strings_sImageUrl.sql'),(107,'2019-09-24 22:10:44','1.1/revision-031/users_items_sAdditionalTime.sql'),(108,'2019-09-24 22:10:44','1.1/revision-032/history_groups_attempts_key_ID.sql'),(110,'2019-09-24 22:10:44','1.1/revision-033/fields_fix.sql'),(111,'2019-09-24 22:10:44','1.1/revision-033/fields_fix_type.sql'),(112,'2019-09-24 22:10:44','1.1/revision-033/keys.sql'),(113,'2019-09-24 22:10:44','1.1/revision-033/remove_history_fields.sql');
ALTER TABLE `schema_revision` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `synchro_version`
--

DROP TABLE IF EXISTS `synchro_version`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `synchro_version` (
  `ID` tinyint(1) NOT NULL,
  `iVersion` int(11) NOT NULL,
  `iLastServerVersion` int(11) NOT NULL,
  `iLastClientVersion` int(11) NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `iVersion` (`iVersion`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `synchro_version`
--

LOCK TABLES `synchro_version` WRITE;
ALTER TABLE `synchro_version` DISABLE KEYS;
INSERT INTO `synchro_version` VALUES (0,0,0,0);
ALTER TABLE `synchro_version` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `threads`
--

DROP TABLE IF EXISTS `threads`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `threads` (
  `ID` bigint(20) NOT NULL,
  `sType` enum('Help','Bug','General') NOT NULL,
  `sLastActivityDate` datetime DEFAULT NULL,
  `idUserCreated` bigint(20) NOT NULL,
  `idItem` bigint(20) DEFAULT NULL,
  `sTitle` varchar(200) DEFAULT NULL,
  `bAdminHelpAsked` tinyint(1) NOT NULL DEFAULT '0',
  `bHidden` tinyint(1) NOT NULL DEFAULT '0',
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `iVersion` (`iVersion`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `threads`
--

LOCK TABLES `threads` WRITE;
ALTER TABLE `threads` DISABLE KEYS;
ALTER TABLE `threads` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_threads` BEFORE INSERT ON `threads` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_threads` AFTER INSERT ON `threads` FOR EACH ROW BEGIN INSERT INTO `history_threads` (`ID`,`iVersion`,`sType`,`idUserCreated`,`idItem`,`sTitle`,`bAdminHelpAsked`,`bHidden`,`sLastActivityDate`) VALUES (NEW.`ID`,@curVersion,NEW.`sType`,NEW.`idUserCreated`,NEW.`idItem`,NEW.`sTitle`,NEW.`bAdminHelpAsked`,NEW.`bHidden`,NEW.`sLastActivityDate`); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_threads` BEFORE UPDATE ON `threads` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`sType` <=> NEW.`sType` AND OLD.`idUserCreated` <=> NEW.`idUserCreated` AND OLD.`idItem` <=> NEW.`idItem` AND OLD.`sTitle` <=> NEW.`sTitle` AND OLD.`bAdminHelpAsked` <=> NEW.`bAdminHelpAsked` AND OLD.`bHidden` <=> NEW.`bHidden` AND OLD.`sLastActivityDate` <=> NEW.`sLastActivityDate`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_threads` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_threads` (`ID`,`iVersion`,`sType`,`idUserCreated`,`idItem`,`sTitle`,`bAdminHelpAsked`,`bHidden`,`sLastActivityDate`)       VALUES (NEW.`ID`,@curVersion,NEW.`sType`,NEW.`idUserCreated`,NEW.`idItem`,NEW.`sTitle`,NEW.`bAdminHelpAsked`,NEW.`bHidden`,NEW.`sLastActivityDate`) ; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_threads` BEFORE DELETE ON `threads` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_threads` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_threads` (`ID`,`iVersion`,`sType`,`idUserCreated`,`idItem`,`sTitle`,`bAdminHelpAsked`,`bHidden`,`sLastActivityDate`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`sType`,OLD.`idUserCreated`,OLD.`idItem`,OLD.`sTitle`,OLD.`bAdminHelpAsked`,OLD.`bHidden`,OLD.`sLastActivityDate`, 1); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `users` (
  `ID` bigint(20) NOT NULL,
  `loginID` bigint(20) DEFAULT NULL COMMENT 'the ''userId'' returned by login platform',
  `tempUser` tinyint(1) NOT NULL,
  `sLogin` varchar(100) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '',
  `sOpenIdIdentity` varchar(255) DEFAULT NULL COMMENT 'User''s Open Id Identity',
  `sPasswordMd5` varchar(100) DEFAULT NULL,
  `sSalt` varchar(32) DEFAULT NULL,
  `sRecover` varchar(50) DEFAULT NULL,
  `sRegistrationDate` datetime DEFAULT NULL,
  `sEmail` varchar(100) DEFAULT NULL,
  `bEmailVerified` tinyint(1) NOT NULL DEFAULT '0',
  `sFirstName` varchar(100) DEFAULT NULL COMMENT 'User''s first name',
  `sLastName` varchar(100) DEFAULT NULL COMMENT 'User''s last name',
  `sStudentId` text,
  `sCountryCode` char(3) NOT NULL DEFAULT '',
  `sTimeZone` varchar(100) DEFAULT NULL,
  `sBirthDate` date DEFAULT NULL COMMENT 'User''s birth date',
  `iGraduationYear` int(11) NOT NULL DEFAULT '0' COMMENT 'User''s high school graduation year',
  `iGrade` int(11) DEFAULT NULL,
  `sSex` enum('Male','Female') DEFAULT NULL,
  `sAddress` mediumtext COMMENT 'User''s address',
  `sZipcode` longtext COMMENT 'User''s postal code',
  `sCity` longtext COMMENT 'User''s city',
  `sLandLineNumber` longtext COMMENT 'User''s phone number',
  `sCellPhoneNumber` longtext COMMENT 'User''s mobil phone number',
  `sDefaultLanguage` char(3) NOT NULL DEFAULT 'fr' COMMENT 'User''s default language',
  `bNotifyNews` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'TODO',
  `sNotify` enum('Never','Answers','Concerned') NOT NULL DEFAULT 'Answers',
  `bPublicFirstName` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Publicly show user''s first name',
  `bPublicLastName` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Publicly show user''s first name',
  `sFreeText` mediumtext,
  `sWebSite` varchar(100) DEFAULT NULL,
  `bPhotoAutoload` tinyint(1) NOT NULL DEFAULT '0',
  `sLangProg` varchar(30) DEFAULT 'Python',
  `sLastLoginDate` datetime DEFAULT NULL,
  `sLastActivityDate` datetime DEFAULT NULL COMMENT 'User''s last activity time on the website',
  `sLastIP` varchar(16) DEFAULT NULL,
  `bBasicEditorMode` tinyint(4) NOT NULL DEFAULT '1',
  `nbSpacesForTab` int(11) NOT NULL DEFAULT '3',
  `iMemberState` tinyint(4) NOT NULL DEFAULT '0',
  `idUserGodfather` int(11) DEFAULT NULL,
  `iStepLevelInSite` int(11) NOT NULL DEFAULT '0' COMMENT 'User''s level',
  `bIsAdmin` tinyint(4) NOT NULL DEFAULT '0',
  `bNoRanking` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'TODO',
  `nbHelpGiven` int(11) NOT NULL DEFAULT '0' COMMENT 'TODO',
  `idGroupSelf` bigint(20) DEFAULT NULL,
  `idGroupOwned` bigint(20) DEFAULT NULL,
  `idGroupAccess` bigint(20) DEFAULT NULL,
  `sNotificationReadDate` datetime DEFAULT NULL,
  `iVersion` bigint(20) NOT NULL,
  `loginModulePrefix` varchar(100) DEFAULT NULL COMMENT 'Set to enable login module accounts manager',
  `creatorID` bigint(20) DEFAULT NULL COMMENT 'which user created a given login with the login generation tool',
  `allowSubgroups` tinyint(4) DEFAULT NULL COMMENT 'Allow to create subgroups',
  PRIMARY KEY (`ID`),
  UNIQUE KEY `sLogin` (`sLogin`),
  UNIQUE KEY `idGroupSelf` (`idGroupSelf`),
  UNIQUE KEY `idGroupOwned` (`idGroupOwned`),
  KEY `iVersion` (`iVersion`),
  KEY `sCountryCode` (`sCountryCode`),
  KEY `idUserGodfather` (`idUserGodfather`),
  KEY `sLangProg` (`sLangProg`),
  KEY `loginID` (`loginID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
ALTER TABLE `users` DISABLE KEYS;
ALTER TABLE `users` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_users` BEFORE INSERT ON `users` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_users` AFTER INSERT ON `users` FOR EACH ROW BEGIN INSERT INTO `history_users` (`ID`,`iVersion`,`sLogin`,`sOpenIdIdentity`,`sPasswordMd5`,`sSalt`,`sRecover`,`sRegistrationDate`,`sEmail`,`bEmailVerified`,`sFirstName`,`sLastName`,`sCountryCode`,`sTimeZone`,`sBirthDate`,`iGraduationYear`,`iGrade`,`sSex`,`sStudentId`,`sAddress`,`sZipcode`,`sCity`,`sLandLineNumber`,`sCellPhoneNumber`,`sDefaultLanguage`,`bNotifyNews`,`sNotify`,`bPublicFirstName`,`bPublicLastName`,`sFreeText`,`sWebSite`,`bPhotoAutoload`,`sLangProg`,`sLastLoginDate`,`sLastActivityDate`,`sLastIP`,`bBasicEditorMode`,`nbSpacesForTab`,`iMemberState`,`idUserGodfather`,`iStepLevelInSite`,`bIsAdmin`,`bNoRanking`,`nbHelpGiven`,`idGroupSelf`,`idGroupOwned`,`idGroupAccess`,`sNotificationReadDate`,`loginModulePrefix`,`allowSubgroups`) VALUES (NEW.`ID`,@curVersion,NEW.`sLogin`,NEW.`sOpenIdIdentity`,NEW.`sPasswordMd5`,NEW.`sSalt`,NEW.`sRecover`,NEW.`sRegistrationDate`,NEW.`sEmail`,NEW.`bEmailVerified`,NEW.`sFirstName`,NEW.`sLastName`,NEW.`sCountryCode`,NEW.`sTimeZone`,NEW.`sBirthDate`,NEW.`iGraduationYear`,NEW.`iGrade`,NEW.`sSex`,NEW.`sStudentId`,NEW.`sAddress`,NEW.`sZipcode`,NEW.`sCity`,NEW.`sLandLineNumber`,NEW.`sCellPhoneNumber`,NEW.`sDefaultLanguage`,NEW.`bNotifyNews`,NEW.`sNotify`,NEW.`bPublicFirstName`,NEW.`bPublicLastName`,NEW.`sFreeText`,NEW.`sWebSite`,NEW.`bPhotoAutoload`,NEW.`sLangProg`,NEW.`sLastLoginDate`,NEW.`sLastActivityDate`,NEW.`sLastIP`,NEW.`bBasicEditorMode`,NEW.`nbSpacesForTab`,NEW.`iMemberState`,NEW.`idUserGodfather`,NEW.`iStepLevelInSite`,NEW.`bIsAdmin`,NEW.`bNoRanking`,NEW.`nbHelpGiven`,NEW.`idGroupSelf`,NEW.`idGroupOwned`,NEW.`idGroupAccess`,NEW.`sNotificationReadDate`,NEW.`loginModulePrefix`,NEW.`allowSubgroups`); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_users` BEFORE UPDATE ON `users` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`sLogin` <=> NEW.`sLogin` AND OLD.`sOpenIdIdentity` <=> NEW.`sOpenIdIdentity` AND OLD.`sPasswordMd5` <=> NEW.`sPasswordMd5` AND OLD.`sSalt` <=> NEW.`sSalt` AND OLD.`sRecover` <=> NEW.`sRecover` AND OLD.`sRegistrationDate` <=> NEW.`sRegistrationDate` AND OLD.`sEmail` <=> NEW.`sEmail` AND OLD.`bEmailVerified` <=> NEW.`bEmailVerified` AND OLD.`sFirstName` <=> NEW.`sFirstName` AND OLD.`sLastName` <=> NEW.`sLastName` AND OLD.`sCountryCode` <=> NEW.`sCountryCode` AND OLD.`sTimeZone` <=> NEW.`sTimeZone` AND OLD.`sBirthDate` <=> NEW.`sBirthDate` AND OLD.`iGraduationYear` <=> NEW.`iGraduationYear` AND OLD.`iGrade` <=> NEW.`iGrade` AND OLD.`sSex` <=> NEW.`sSex` AND OLD.`sStudentId` <=> NEW.`sStudentId` AND OLD.`sAddress` <=> NEW.`sAddress` AND OLD.`sZipcode` <=> NEW.`sZipcode` AND OLD.`sCity` <=> NEW.`sCity` AND OLD.`sLandLineNumber` <=> NEW.`sLandLineNumber` AND OLD.`sCellPhoneNumber` <=> NEW.`sCellPhoneNumber` AND OLD.`sDefaultLanguage` <=> NEW.`sDefaultLanguage` AND OLD.`bNotifyNews` <=> NEW.`bNotifyNews` AND OLD.`sNotify` <=> NEW.`sNotify` AND OLD.`bPublicFirstName` <=> NEW.`bPublicFirstName` AND OLD.`bPublicLastName` <=> NEW.`bPublicLastName` AND OLD.`sFreeText` <=> NEW.`sFreeText` AND OLD.`sWebSite` <=> NEW.`sWebSite` AND OLD.`bPhotoAutoload` <=> NEW.`bPhotoAutoload` AND OLD.`sLangProg` <=> NEW.`sLangProg` AND OLD.`sLastLoginDate` <=> NEW.`sLastLoginDate` AND OLD.`sLastActivityDate` <=> NEW.`sLastActivityDate` AND OLD.`sLastIP` <=> NEW.`sLastIP` AND OLD.`bBasicEditorMode` <=> NEW.`bBasicEditorMode` AND OLD.`nbSpacesForTab` <=> NEW.`nbSpacesForTab` AND OLD.`iMemberState` <=> NEW.`iMemberState` AND OLD.`idUserGodfather` <=> NEW.`idUserGodfather` AND OLD.`iStepLevelInSite` <=> NEW.`iStepLevelInSite` AND OLD.`bIsAdmin` <=> NEW.`bIsAdmin` AND OLD.`bNoRanking` <=> NEW.`bNoRanking` AND OLD.`nbHelpGiven` <=> NEW.`nbHelpGiven` AND OLD.`idGroupSelf` <=> NEW.`idGroupSelf` AND OLD.`idGroupOwned` <=> NEW.`idGroupOwned` AND OLD.`idGroupAccess` <=> NEW.`idGroupAccess` AND OLD.`sNotificationReadDate` <=> NEW.`sNotificationReadDate` AND OLD.`loginModulePrefix` <=> NEW.`loginModulePrefix` AND OLD.`allowSubgroups` <=> NEW.`allowSubgroups`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_users` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_users` (`ID`,`iVersion`,`sLogin`,`sOpenIdIdentity`,`sPasswordMd5`,`sSalt`,`sRecover`,`sRegistrationDate`,`sEmail`,`bEmailVerified`,`sFirstName`,`sLastName`,`sCountryCode`,`sTimeZone`,`sBirthDate`,`iGraduationYear`,`iGrade`,`sSex`,`sStudentId`,`sAddress`,`sZipcode`,`sCity`,`sLandLineNumber`,`sCellPhoneNumber`,`sDefaultLanguage`,`bNotifyNews`,`sNotify`,`bPublicFirstName`,`bPublicLastName`,`sFreeText`,`sWebSite`,`bPhotoAutoload`,`sLangProg`,`sLastLoginDate`,`sLastActivityDate`,`sLastIP`,`bBasicEditorMode`,`nbSpacesForTab`,`iMemberState`,`idUserGodfather`,`iStepLevelInSite`,`bIsAdmin`,`bNoRanking`,`nbHelpGiven`,`idGroupSelf`,`idGroupOwned`,`idGroupAccess`,`sNotificationReadDate`,`loginModulePrefix`,`allowSubgroups`)    VALUES (NEW.`ID`,@curVersion,NEW.`sLogin`,NEW.`sOpenIdIdentity`,NEW.`sPasswordMd5`,NEW.`sSalt`,NEW.`sRecover`,NEW.`sRegistrationDate`,NEW.`sEmail`,NEW.`bEmailVerified`,NEW.`sFirstName`,NEW.`sLastName`,NEW.`sCountryCode`,NEW.`sTimeZone`,NEW.`sBirthDate`,NEW.`iGraduationYear`,NEW.`iGrade`,NEW.`sSex`,NEW.`sStudentId`,NEW.`sAddress`,NEW.`sZipcode`,NEW.`sCity`,NEW.`sLandLineNumber`,NEW.`sCellPhoneNumber`,NEW.`sDefaultLanguage`,NEW.`bNotifyNews`,NEW.`sNotify`,NEW.`bPublicFirstName`,NEW.`bPublicLastName`,NEW.`sFreeText`,NEW.`sWebSite`,NEW.`bPhotoAutoload`,NEW.`sLangProg`,NEW.`sLastLoginDate`,NEW.`sLastActivityDate`,NEW.`sLastIP`,NEW.`bBasicEditorMode`,NEW.`nbSpacesForTab`,NEW.`iMemberState`,NEW.`idUserGodfather`,NEW.`iStepLevelInSite`,NEW.`bIsAdmin`,NEW.`bNoRanking`,NEW.`nbHelpGiven`,NEW.`idGroupSelf`,NEW.`idGroupOwned`,NEW.`idGroupAccess`,NEW.`sNotificationReadDate`,NEW.`loginModulePrefix`,NEW.`allowSubgroups`) ; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_users` BEFORE DELETE ON `users` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_users` (`ID`,`iVersion`,`sLogin`,`sOpenIdIdentity`,`sPasswordMd5`,`sSalt`,`sRecover`,`sRegistrationDate`,`sEmail`,`bEmailVerified`,`sFirstName`,`sLastName`,`sCountryCode`,`sTimeZone`,`sBirthDate`,`iGraduationYear`,`iGrade`,`sSex`,`sStudentId`,`sAddress`,`sZipcode`,`sCity`,`sLandLineNumber`,`sCellPhoneNumber`,`sDefaultLanguage`,`bNotifyNews`,`sNotify`,`bPublicFirstName`,`bPublicLastName`,`sFreeText`,`sWebSite`,`bPhotoAutoload`,`sLangProg`,`sLastLoginDate`,`sLastActivityDate`,`sLastIP`,`bBasicEditorMode`,`nbSpacesForTab`,`iMemberState`,`idUserGodfather`,`iStepLevelInSite`,`bIsAdmin`,`bNoRanking`,`nbHelpGiven`,`idGroupSelf`,`idGroupOwned`,`idGroupAccess`,`sNotificationReadDate`,`loginModulePrefix`,`allowSubgroups`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`sLogin`,OLD.`sOpenIdIdentity`,OLD.`sPasswordMd5`,OLD.`sSalt`,OLD.`sRecover`,OLD.`sRegistrationDate`,OLD.`sEmail`,OLD.`bEmailVerified`,OLD.`sFirstName`,OLD.`sLastName`,OLD.`sCountryCode`,OLD.`sTimeZone`,OLD.`sBirthDate`,OLD.`iGraduationYear`,OLD.`iGrade`,OLD.`sSex`,OLD.`sStudentId`,OLD.`sAddress`,OLD.`sZipcode`,OLD.`sCity`,OLD.`sLandLineNumber`,OLD.`sCellPhoneNumber`,OLD.`sDefaultLanguage`,OLD.`bNotifyNews`,OLD.`sNotify`,OLD.`bPublicFirstName`,OLD.`bPublicLastName`,OLD.`sFreeText`,OLD.`sWebSite`,OLD.`bPhotoAutoload`,OLD.`sLangProg`,OLD.`sLastLoginDate`,OLD.`sLastActivityDate`,OLD.`sLastIP`,OLD.`bBasicEditorMode`,OLD.`nbSpacesForTab`,OLD.`iMemberState`,OLD.`idUserGodfather`,OLD.`iStepLevelInSite`,OLD.`bIsAdmin`,OLD.`bNoRanking`,OLD.`nbHelpGiven`,OLD.`idGroupSelf`,OLD.`idGroupOwned`,OLD.`idGroupAccess`,OLD.`sNotificationReadDate`,OLD.`loginModulePrefix`,OLD.`allowSubgroups`, 1); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `users_answers`
--

DROP TABLE IF EXISTS `users_answers`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `users_answers` (
  `ID` bigint(20) NOT NULL AUTO_INCREMENT,
  `idUser` bigint(20) NOT NULL,
  `idItem` bigint(20) NOT NULL,
  `idAttempt` bigint(20) DEFAULT NULL,
  `sName` varchar(200) DEFAULT NULL,
  `sType` enum('Submission','Saved','Current') NOT NULL DEFAULT 'Submission',
  `sState` mediumtext,
  `sAnswer` mediumtext,
  `sLangProg` varchar(50) DEFAULT NULL,
  `sSubmissionDate` datetime NOT NULL,
  `iScore` float DEFAULT NULL,
  `bValidated` tinyint(1) DEFAULT NULL,
  `sGradingDate` datetime DEFAULT NULL,
  `idUserGrader` int(11) DEFAULT NULL,
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `idUser` (`idUser`),
  KEY `idItem` (`idItem`),
  KEY `idAttempt` (`idAttempt`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `users_answers`
--

LOCK TABLES `users_answers` WRITE;
ALTER TABLE `users_answers` DISABLE KEYS;
ALTER TABLE `users_answers` ENABLE KEYS;
UNLOCK TABLES;

--
-- Table structure for table `users_items`
--

DROP TABLE IF EXISTS `users_items`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `users_items` (
  `ID` bigint(20) NOT NULL AUTO_INCREMENT,
  `idUser` bigint(20) NOT NULL,
  `idItem` bigint(20) NOT NULL,
  `idAttemptActive` bigint(20) DEFAULT NULL,
  `iScore` float NOT NULL DEFAULT '0',
  `iScoreComputed` float NOT NULL DEFAULT '0',
  `iScoreReeval` float DEFAULT '0',
  `iScoreDiffManual` float NOT NULL DEFAULT '0',
  `sScoreDiffComment` varchar(200) NOT NULL DEFAULT '',
  `nbSubmissionsAttempts` int(11) NOT NULL DEFAULT '0',
  `nbTasksTried` int(11) NOT NULL DEFAULT '0',
  `nbTasksSolved` int(11) NOT NULL DEFAULT '0',
  `nbChildrenValidated` int(11) NOT NULL DEFAULT '0',
  `bValidated` tinyint(1) NOT NULL DEFAULT '0',
  `bFinished` tinyint(1) NOT NULL DEFAULT '0',
  `bKeyObtained` tinyint(1) NOT NULL DEFAULT '0',
  `nbTasksWithHelp` int(11) NOT NULL DEFAULT '0',
  `sHintsRequested` mediumtext,
  `nbHintsCached` int(11) NOT NULL DEFAULT '0',
  `nbCorrectionsRead` int(11) NOT NULL DEFAULT '0',
  `iPrecision` int(11) NOT NULL DEFAULT '0',
  `iAutonomy` int(11) NOT NULL DEFAULT '0',
  `sStartDate` datetime DEFAULT NULL,
  `sValidationDate` datetime DEFAULT NULL,
  `sFinishDate` datetime DEFAULT NULL,
  `sLastActivityDate` datetime DEFAULT NULL,
  `sThreadStartDate` datetime DEFAULT NULL,
  `sBestAnswerDate` datetime DEFAULT NULL,
  `sLastAnswerDate` datetime DEFAULT NULL,
  `sLastHintDate` datetime DEFAULT NULL,
  `sAdditionalTime` time DEFAULT NULL,
  `sContestStartDate` datetime DEFAULT NULL,
  `bRanked` tinyint(1) NOT NULL DEFAULT '0',
  `sAllLangProg` varchar(200) DEFAULT NULL,
  `iVersion` bigint(20) NOT NULL,
  `sAncestorsComputationState` enum('done','processing','todo','temp') NOT NULL DEFAULT 'todo',
  `sState` mediumtext,
  `sAnswer` mediumtext,
  `bPlatformDataRemoved` tinyint(4) NOT NULL DEFAULT '0',
  PRIMARY KEY (`ID`),
  UNIQUE KEY `UserItem` (`idUser`,`idItem`),
  KEY `iVersion` (`iVersion`),
  KEY `sAncestorsComputationState` (`sAncestorsComputationState`),
  KEY `idItem` (`idItem`),
  KEY `idUser` (`idUser`),
  KEY `idAttemptActive` (`idAttemptActive`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `users_items`
--

LOCK TABLES `users_items` WRITE;
ALTER TABLE `users_items` DISABLE KEYS;
ALTER TABLE `users_items` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_users_items` BEFORE INSERT ON `users_items` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_users_items` AFTER INSERT ON `users_items` FOR EACH ROW BEGIN INSERT INTO `history_users_items` (`ID`,`iVersion`,`idUser`,`idItem`,`idAttemptActive`,`iScore`,`iScoreComputed`,`iScoreReeval`,`iScoreDiffManual`,`sScoreDiffComment`,`nbSubmissionsAttempts`,`nbTasksTried`,`nbChildrenValidated`,`bValidated`,`bFinished`,`bKeyObtained`,`nbTasksWithHelp`,`sHintsRequested`,`nbHintsCached`,`nbCorrectionsRead`,`iPrecision`,`iAutonomy`,`sStartDate`,`sValidationDate`,`sBestAnswerDate`,`sLastAnswerDate`,`sThreadStartDate`,`sLastHintDate`,`sFinishDate`,`sLastActivityDate`,`sContestStartDate`,`bRanked`,`sAllLangProg`,`sState`,`sAnswer`) VALUES (NEW.`ID`,@curVersion,NEW.`idUser`,NEW.`idItem`,NEW.`idAttemptActive`,NEW.`iScore`,NEW.`iScoreComputed`,NEW.`iScoreReeval`,NEW.`iScoreDiffManual`,NEW.`sScoreDiffComment`,NEW.`nbSubmissionsAttempts`,NEW.`nbTasksTried`,NEW.`nbChildrenValidated`,NEW.`bValidated`,NEW.`bFinished`,NEW.`bKeyObtained`,NEW.`nbTasksWithHelp`,NEW.`sHintsRequested`,NEW.`nbHintsCached`,NEW.`nbCorrectionsRead`,NEW.`iPrecision`,NEW.`iAutonomy`,NEW.`sStartDate`,NEW.`sValidationDate`,NEW.`sBestAnswerDate`,NEW.`sLastAnswerDate`,NEW.`sThreadStartDate`,NEW.`sLastHintDate`,NEW.`sFinishDate`,NEW.`sLastActivityDate`,NEW.`sContestStartDate`,NEW.`bRanked`,NEW.`sAllLangProg`,NEW.`sState`,NEW.`sAnswer`); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_users_items` BEFORE UPDATE ON `users_items` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idUser` <=> NEW.`idUser` AND OLD.`idItem` <=> NEW.`idItem` AND OLD.`idAttemptActive` <=> NEW.`idAttemptActive` AND OLD.`iScore` <=> NEW.`iScore` AND OLD.`iScoreComputed` <=> NEW.`iScoreComputed` AND OLD.`iScoreReeval` <=> NEW.`iScoreReeval` AND OLD.`iScoreDiffManual` <=> NEW.`iScoreDiffManual` AND OLD.`sScoreDiffComment` <=> NEW.`sScoreDiffComment` AND OLD.`nbTasksTried` <=> NEW.`nbTasksTried` AND OLD.`nbChildrenValidated` <=> NEW.`nbChildrenValidated` AND OLD.`bValidated` <=> NEW.`bValidated` AND OLD.`bFinished` <=> NEW.`bFinished` AND OLD.`bKeyObtained` <=> NEW.`bKeyObtained` AND OLD.`nbTasksWithHelp` <=> NEW.`nbTasksWithHelp` AND OLD.`sHintsRequested` <=> NEW.`sHintsRequested` AND OLD.`nbHintsCached` <=> NEW.`nbHintsCached` AND OLD.`nbCorrectionsRead` <=> NEW.`nbCorrectionsRead` AND OLD.`iPrecision` <=> NEW.`iPrecision` AND OLD.`iAutonomy` <=> NEW.`iAutonomy` AND OLD.`sStartDate` <=> NEW.`sStartDate` AND OLD.`sValidationDate` <=> NEW.`sValidationDate` AND OLD.`sBestAnswerDate` <=> NEW.`sBestAnswerDate` AND OLD.`sLastAnswerDate` <=> NEW.`sLastAnswerDate` AND OLD.`sThreadStartDate` <=> NEW.`sThreadStartDate` AND OLD.`sLastHintDate` <=> NEW.`sLastHintDate` AND OLD.`sFinishDate` <=> NEW.`sFinishDate` AND OLD.`sContestStartDate` <=> NEW.`sContestStartDate` AND OLD.`bRanked` <=> NEW.`bRanked` AND OLD.`sAllLangProg` <=> NEW.`sAllLangProg` AND OLD.`sState` <=> NEW.`sState` AND OLD.`sAnswer` <=> NEW.`sAnswer`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_users_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_users_items` (`ID`,`iVersion`,`idUser`,`idItem`,`idAttemptActive`,`iScore`,`iScoreComputed`,`iScoreReeval`,`iScoreDiffManual`,`sScoreDiffComment`,`nbSubmissionsAttempts`,`nbTasksTried`,`nbChildrenValidated`,`bValidated`,`bFinished`,`bKeyObtained`,`nbTasksWithHelp`,`sHintsRequested`,`nbHintsCached`,`nbCorrectionsRead`,`iPrecision`,`iAutonomy`,`sStartDate`,`sValidationDate`,`sBestAnswerDate`,`sLastAnswerDate`,`sThreadStartDate`,`sLastHintDate`,`sFinishDate`,`sLastActivityDate`,`sContestStartDate`,`bRanked`,`sAllLangProg`,`sState`,`sAnswer`)       VALUES (NEW.`ID`,@curVersion,NEW.`idUser`,NEW.`idItem`,NEW.`idAttemptActive`,NEW.`iScore`,NEW.`iScoreComputed`,NEW.`iScoreReeval`,NEW.`iScoreDiffManual`,NEW.`sScoreDiffComment`,NEW.`nbSubmissionsAttempts`,NEW.`nbTasksTried`,NEW.`nbChildrenValidated`,NEW.`bValidated`,NEW.`bFinished`,NEW.`bKeyObtained`,NEW.`nbTasksWithHelp`,NEW.`sHintsRequested`,NEW.`nbHintsCached`,NEW.`nbCorrectionsRead`,NEW.`iPrecision`,NEW.`iAutonomy`,NEW.`sStartDate`,NEW.`sValidationDate`,NEW.`sBestAnswerDate`,NEW.`sLastAnswerDate`,NEW.`sThreadStartDate`,NEW.`sLastHintDate`,NEW.`sFinishDate`,NEW.`sLastActivityDate`,NEW.`sContestStartDate`,NEW.`bRanked`,NEW.`sAllLangProg`,NEW.`sState`,NEW.`sAnswer`) ; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_users_items` BEFORE DELETE ON `users_items` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users_items` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_users_items` (`ID`,`iVersion`,`idUser`,`idItem`,`idAttemptActive`,`iScore`,`iScoreComputed`,`iScoreReeval`,`iScoreDiffManual`,`sScoreDiffComment`,`nbSubmissionsAttempts`,`nbTasksTried`,`nbChildrenValidated`,`bValidated`,`bFinished`,`bKeyObtained`,`nbTasksWithHelp`,`sHintsRequested`,`nbHintsCached`,`nbCorrectionsRead`,`iPrecision`,`iAutonomy`,`sStartDate`,`sValidationDate`,`sBestAnswerDate`,`sLastAnswerDate`,`sThreadStartDate`,`sLastHintDate`,`sFinishDate`,`sLastActivityDate`,`sContestStartDate`,`bRanked`,`sAllLangProg`,`sState`,`sAnswer`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idUser`,OLD.`idItem`,OLD.`idAttemptActive`,OLD.`iScore`,OLD.`iScoreComputed`,OLD.`iScoreReeval`,OLD.`iScoreDiffManual`,OLD.`sScoreDiffComment`,OLD.`nbSubmissionsAttempts`,OLD.`nbTasksTried`,OLD.`nbChildrenValidated`,OLD.`bValidated`,OLD.`bFinished`,OLD.`bKeyObtained`,OLD.`nbTasksWithHelp`,OLD.`sHintsRequested`,OLD.`nbHintsCached`,OLD.`nbCorrectionsRead`,OLD.`iPrecision`,OLD.`iAutonomy`,OLD.`sStartDate`,OLD.`sValidationDate`,OLD.`sBestAnswerDate`,OLD.`sLastAnswerDate`,OLD.`sThreadStartDate`,OLD.`sLastHintDate`,OLD.`sFinishDate`,OLD.`sLastActivityDate`,OLD.`sContestStartDate`,OLD.`bRanked`,OLD.`sAllLangProg`,OLD.`sState`,OLD.`sAnswer`, 1); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Table structure for table `users_threads`
--

DROP TABLE IF EXISTS `users_threads`;
SET @saved_cs_client     = @@character_set_client;
 SET character_set_client = utf8mb4;
CREATE TABLE `users_threads` (
  `ID` bigint(20) NOT NULL,
  `idUser` bigint(20) NOT NULL,
  `idThread` bigint(20) NOT NULL,
  `sLastReadDate` datetime DEFAULT NULL,
  `bParticipated` tinyint(1) NOT NULL DEFAULT '0',
  `sLastWriteDate` datetime DEFAULT NULL,
  `bStarred` tinyint(1) DEFAULT NULL,
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `userThread` (`idUser`,`idThread`),
  KEY `users_idx` (`idUser`),
  KEY `iVersion` (`iVersion`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
SET character_set_client = @saved_cs_client;

--
-- Dumping data for table `users_threads`
--

LOCK TABLES `users_threads` WRITE;
ALTER TABLE `users_threads` DISABLE KEYS;
ALTER TABLE `users_threads` ENABLE KEYS;
UNLOCK TABLES;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_insert_users_threads` BEFORE INSERT ON `users_threads` FOR EACH ROW BEGIN IF (NEW.ID IS NULL OR NEW.ID = 0) THEN SET NEW.ID = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;SET NEW.iVersion = @curVersion; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `after_insert_users_threads` AFTER INSERT ON `users_threads` FOR EACH ROW BEGIN INSERT INTO `history_users_threads` (`ID`,`iVersion`,`idUser`,`idThread`,`sLastReadDate`,`sLastWriteDate`,`bStarred`) VALUES (NEW.`ID`,@curVersion,NEW.`idUser`,NEW.`idThread`,NEW.`sLastReadDate`,NEW.`sLastWriteDate`,NEW.`bStarred`); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_update_users_threads` BEFORE UPDATE ON `users_threads` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idUser` <=> NEW.`idUser` AND OLD.`idThread` <=> NEW.`idThread` AND OLD.`sLastReadDate` <=> NEW.`sLastReadDate` AND OLD.`sLastWriteDate` <=> NEW.`sLastWriteDate` AND OLD.`bStarred` <=> NEW.`bStarred`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_users_threads` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_users_threads` (`ID`,`iVersion`,`idUser`,`idThread`,`sLastReadDate`,`sLastWriteDate`,`bStarred`)       VALUES (NEW.`ID`,@curVersion,NEW.`idUser`,NEW.`idThread`,NEW.`sLastReadDate`,NEW.`sLastWriteDate`,NEW.`bStarred`) ; END IF; END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;
SET @saved_cs_client      = @@character_set_client;
SET @saved_cs_results     = @@character_set_results;
SET @saved_col_connection = @@collation_connection;
SET character_set_client  = utf8;
SET character_set_results = utf8;
SET collation_connection  = utf8_general_ci;
SET @saved_sql_mode       = @@sql_mode;
SET sql_mode              = 'NO_ENGINE_SUBSTITUTION';
DELIMITER ;;
CREATE TRIGGER `before_delete_users_threads` BEFORE DELETE ON `users_threads` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_users_threads` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_users_threads` (`ID`,`iVersion`,`idUser`,`idThread`,`sLastReadDate`,`sLastWriteDate`,`bStarred`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idUser`,OLD.`idThread`,OLD.`sLastReadDate`,OLD.`sLastWriteDate`,OLD.`bStarred`, 1); END ;;
DELIMITER ;
SET sql_mode              = @saved_sql_mode;
SET character_set_client  = @saved_cs_client;
SET character_set_results = @saved_cs_results;
SET collation_connection  = @saved_col_connection;

--
-- Dumping events for database 'algorea_db'
--

--
-- Dumping routines for database 'algorea_db'
--
SET TIME_ZONE=@OLD_TIME_ZONE;

SET SQL_MODE=@OLD_SQL_MODE;
SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS;
SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS;
SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT;
SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS;
SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION;
SET SQL_NOTES=@OLD_SQL_NOTES;

-- Dump completed on 2019-09-11 19:47:24

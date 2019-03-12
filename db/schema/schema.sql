-- MySQL dump 10.13  Distrib 5.7.23, for osx10.14 (x86_64)
--
-- Host: localhost    Database: algorea_db
-- ------------------------------------------------------
-- Server version	5.7.23

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `badges`
--

DROP TABLE IF EXISTS `badges`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `badges` (
  `ID` bigint(20) NOT NULL AUTO_INCREMENT,
  `idUser` bigint(20) NOT NULL,  `name` text,
  `code` text NOT NULL,
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `badges`
--

LOCK TABLES `badges` WRITE;
/*!40000 ALTER TABLE `badges` DISABLE KEYS */;
/*!40000 ALTER TABLE `badges` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `error_log`
--

DROP TABLE IF EXISTS `error_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `error_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `url` text COLLATE utf8_unicode_ci NOT NULL,
  `browser` text COLLATE utf8_unicode_ci NOT NULL,
  `details` text COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `error_log`
--

LOCK TABLES `error_log` WRITE;
/*!40000 ALTER TABLE `error_log` DISABLE KEYS */;
/*!40000 ALTER TABLE `error_log` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `filters`
--

DROP TABLE IF EXISTS `filters`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `filters`
--

LOCK TABLES `filters` WRITE;
/*!40000 ALTER TABLE `filters` DISABLE KEYS */;
/*!40000 ALTER TABLE `filters` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `gorp_migrations`
--

DROP TABLE IF EXISTS `gorp_migrations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `gorp_migrations` (
  `id` varchar(255) NOT NULL,
  `applied_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `gorp_migrations`
--

LOCK TABLES `gorp_migrations` WRITE;
/*!40000 ALTER TABLE `gorp_migrations` DISABLE KEYS */;
/*!40000 ALTER TABLE `gorp_migrations` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `groups`
--

DROP TABLE IF EXISTS `groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `groups`
--

LOCK TABLES `groups` WRITE;
/*!40000 ALTER TABLE `groups` DISABLE KEYS */;
/*!40000 ALTER TABLE `groups` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `groups_ancestors`
--

DROP TABLE IF EXISTS `groups_ancestors`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `groups_ancestors`
--

LOCK TABLES `groups_ancestors` WRITE;
/*!40000 ALTER TABLE `groups_ancestors` DISABLE KEYS */;
/*!40000 ALTER TABLE `groups_ancestors` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `groups_attempts`
--

DROP TABLE IF EXISTS `groups_attempts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `groups_attempts`
--

LOCK TABLES `groups_attempts` WRITE;
/*!40000 ALTER TABLE `groups_attempts` DISABLE KEYS */;
/*!40000 ALTER TABLE `groups_attempts` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `groups_groups`
--

DROP TABLE IF EXISTS `groups_groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `groups_groups`
--

LOCK TABLES `groups_groups` WRITE;
/*!40000 ALTER TABLE `groups_groups` DISABLE KEYS */;
/*!40000 ALTER TABLE `groups_groups` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `groups_items`
--

DROP TABLE IF EXISTS `groups_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
  KEY `sPropagateAccess` (`sPropagateAccess`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `groups_items`
--

LOCK TABLES `groups_items` WRITE;
/*!40000 ALTER TABLE `groups_items` DISABLE KEYS */;
/*!40000 ALTER TABLE `groups_items` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `groups_items_propagate`
--

DROP TABLE IF EXISTS `groups_items_propagate`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `groups_items_propagate` (
  `ID` bigint(20) NOT NULL,
  `sPropagateAccess` enum('self','children','done') NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `sPropagateAccess` (`sPropagateAccess`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `groups_items_propagate`
--

LOCK TABLES `groups_items_propagate` WRITE;
/*!40000 ALTER TABLE `groups_items_propagate` DISABLE KEYS */;
/*!40000 ALTER TABLE `groups_items_propagate` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `groups_login_prefixes`
--

DROP TABLE IF EXISTS `groups_login_prefixes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `groups_login_prefixes` (
  `ID` bigint(20) NOT NULL AUTO_INCREMENT,
  `idGroup` bigint(20) NOT NULL,
  `prefix` varchar(100) COLLATE utf8_unicode_ci NOT NULL,
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  UNIQUE KEY `prefix` (`prefix`),
  KEY `idGroup` (`idGroup`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `groups_login_prefixes`
--

LOCK TABLES `groups_login_prefixes` WRITE;
/*!40000 ALTER TABLE `groups_login_prefixes` DISABLE KEYS */;
/*!40000 ALTER TABLE `groups_login_prefixes` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `groups_propagate`
--

DROP TABLE IF EXISTS `groups_propagate`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `groups_propagate` (
  `ID` bigint(20) NOT NULL,
  `sAncestorsComputationState` enum('todo','done','processing','') NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `sAncestorsComputationState` (`sAncestorsComputationState`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `groups_propagate`
--

LOCK TABLES `groups_propagate` WRITE;
/*!40000 ALTER TABLE `groups_propagate` DISABLE KEYS */;
/*!40000 ALTER TABLE `groups_propagate` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_filters`
--

DROP TABLE IF EXISTS `history_filters`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `history_filters` (
  `history_ID` int(11) NOT NULL AUTO_INCREMENT,
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
  PRIMARY KEY (`history_ID`),
  KEY `user_idx` (`idUser`),
  KEY `iVersion` (`iVersion`),
  KEY `iNextVersion` (`iNextVersion`),
  KEY `bDeleted` (`bDeleted`),
  KEY `ID` (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_filters`
--

LOCK TABLES `history_filters` WRITE;
/*!40000 ALTER TABLE `history_filters` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_filters` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_groups`
--

DROP TABLE IF EXISTS `history_groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_groups`
--

LOCK TABLES `history_groups` WRITE;
/*!40000 ALTER TABLE `history_groups` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_groups` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_groups_ancestors`
--

DROP TABLE IF EXISTS `history_groups_ancestors`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_groups_ancestors`
--

LOCK TABLES `history_groups_ancestors` WRITE;
/*!40000 ALTER TABLE `history_groups_ancestors` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_groups_ancestors` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_groups_attempts`
--

DROP TABLE IF EXISTS `history_groups_attempts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
  `sPropagationState` enum('done','processing','todo','temp') NOT NULL DEFAULT 'done',
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL,
  PRIMARY KEY (`historyID`),
  KEY `iVersion` (`iVersion`),
  KEY `sAncestorsComputationState` (`sPropagationState`),
  KEY `idItem` (`idItem`),
  KEY `GroupItem` (`idGroup`,`idItem`),
  KEY `idGroup` (`idGroup`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_groups_attempts`
--

LOCK TABLES `history_groups_attempts` WRITE;
/*!40000 ALTER TABLE `history_groups_attempts` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_groups_attempts` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_groups_groups`
--

DROP TABLE IF EXISTS `history_groups_groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_groups_groups`
--

LOCK TABLES `history_groups_groups` WRITE;
/*!40000 ALTER TABLE `history_groups_groups` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_groups_groups` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_groups_items`
--

DROP TABLE IF EXISTS `history_groups_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_groups_items`
--

LOCK TABLES `history_groups_items` WRITE;
/*!40000 ALTER TABLE `history_groups_items` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_groups_items` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_groups_login_prefixes`
--

DROP TABLE IF EXISTS `history_groups_login_prefixes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `history_groups_login_prefixes` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `idGroup` bigint(20) NOT NULL,
  `prefix` varchar(100) COLLATE utf8_unicode_ci NOT NULL,
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(4) NOT NULL DEFAULT '0',
  PRIMARY KEY (`historyID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_groups_login_prefixes`
--

LOCK TABLES `history_groups_login_prefixes` WRITE;
/*!40000 ALTER TABLE `history_groups_login_prefixes` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_groups_login_prefixes` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_items`
--

DROP TABLE IF EXISTS `history_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `history_items` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `sUrl` varchar(200) DEFAULT NULL,
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
  `sAncestorsComputationState` enum('done','processing','todo') NOT NULL DEFAULT 'todo',
  `sAncestorsAccessComputationState` enum('todo','processing','done') NOT NULL,
  `iVersion` bigint(20) NOT NULL,
  `iNextVersion` bigint(20) DEFAULT NULL,
  `bDeleted` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`historyID`),
  KEY `ID` (`ID`),
  KEY `iVersion` (`iVersion`),
  KEY `iNextVersion` (`iNextVersion`),
  KEY `bDeleted` (`bDeleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_items`
--

LOCK TABLES `history_items` WRITE;
/*!40000 ALTER TABLE `history_items` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_items` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_items_ancestors`
--

DROP TABLE IF EXISTS `history_items_ancestors`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_items_ancestors`
--

LOCK TABLES `history_items_ancestors` WRITE;
/*!40000 ALTER TABLE `history_items_ancestors` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_items_ancestors` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_items_items`
--

DROP TABLE IF EXISTS `history_items_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_items_items`
--

LOCK TABLES `history_items_items` WRITE;
/*!40000 ALTER TABLE `history_items_items` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_items_items` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_items_strings`
--

DROP TABLE IF EXISTS `history_items_strings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `history_items_strings` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
  `idItem` bigint(20) NOT NULL,
  `idLanguage` bigint(20) NOT NULL,
  `sTranslator` varchar(100) DEFAULT NULL,
  `sTitle` varchar(200) NOT NULL DEFAULT '',
  `sImageUrl` varchar(100) DEFAULT NULL,
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_items_strings`
--

LOCK TABLES `history_items_strings` WRITE;
/*!40000 ALTER TABLE `history_items_strings` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_items_strings` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_languages`
--

DROP TABLE IF EXISTS `history_languages`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_languages`
--

LOCK TABLES `history_languages` WRITE;
/*!40000 ALTER TABLE `history_languages` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_languages` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_messages`
--

DROP TABLE IF EXISTS `history_messages`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
  `bReadByCandidate` tinyint(1) DEFAULT NULL,
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_messages`
--

LOCK TABLES `history_messages` WRITE;
/*!40000 ALTER TABLE `history_messages` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_messages` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_threads`
--

DROP TABLE IF EXISTS `history_threads`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_threads`
--

LOCK TABLES `history_threads` WRITE;
/*!40000 ALTER TABLE `history_threads` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_threads` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_users`
--

DROP TABLE IF EXISTS `history_users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `history_users` (
  `historyID` bigint(20) NOT NULL AUTO_INCREMENT,
  `ID` bigint(20) NOT NULL,
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_users`
--

LOCK TABLES `history_users` WRITE;
/*!40000 ALTER TABLE `history_users` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_users_items`
--

DROP TABLE IF EXISTS `history_users_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
  `sAdditionalTime` datetime DEFAULT NULL,
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_users_items`
--

LOCK TABLES `history_users_items` WRITE;
/*!40000 ALTER TABLE `history_users_items` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_users_items` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `history_users_threads`
--

DROP TABLE IF EXISTS `history_users_threads`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `history_users_threads`
--

LOCK TABLES `history_users_threads` WRITE;
/*!40000 ALTER TABLE `history_users_threads` DISABLE KEYS */;
/*!40000 ALTER TABLE `history_users_threads` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `items`
--

DROP TABLE IF EXISTS `items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `items` (
  `ID` bigint(20) NOT NULL,
  `sUrl` varchar(200) DEFAULT NULL,
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `items`
--

LOCK TABLES `items` WRITE;
/*!40000 ALTER TABLE `items` DISABLE KEYS */;
/*!40000 ALTER TABLE `items` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `items_ancestors`
--

DROP TABLE IF EXISTS `items_ancestors`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `items_ancestors`
--

LOCK TABLES `items_ancestors` WRITE;
/*!40000 ALTER TABLE `items_ancestors` DISABLE KEYS */;
/*!40000 ALTER TABLE `items_ancestors` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `items_items`
--

DROP TABLE IF EXISTS `items_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `items_items`
--

LOCK TABLES `items_items` WRITE;
/*!40000 ALTER TABLE `items_items` DISABLE KEYS */;
/*!40000 ALTER TABLE `items_items` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `items_propagate`
--

DROP TABLE IF EXISTS `items_propagate`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `items_propagate` (
  `ID` bigint(20) NOT NULL,
  `sAncestorsComputationState` enum('todo','done','processing','') NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `sAncestorsComputationDate` (`sAncestorsComputationState`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `items_propagate`
--

LOCK TABLES `items_propagate` WRITE;
/*!40000 ALTER TABLE `items_propagate` DISABLE KEYS */;
/*!40000 ALTER TABLE `items_propagate` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `items_strings`
--

DROP TABLE IF EXISTS `items_strings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `items_strings` (
  `ID` bigint(20) NOT NULL,
  `idItem` bigint(20) NOT NULL,
  `idLanguage` bigint(20) NOT NULL,
  `sTranslator` varchar(100) DEFAULT NULL,
  `sTitle` varchar(200) NOT NULL DEFAULT '',
  `sImageUrl` varchar(100) DEFAULT NULL,
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `items_strings`
--

LOCK TABLES `items_strings` WRITE;
/*!40000 ALTER TABLE `items_strings` DISABLE KEYS */;
/*!40000 ALTER TABLE `items_strings` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `languages`
--

DROP TABLE IF EXISTS `languages`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `languages` (
  `ID` bigint(20) NOT NULL,
  `sName` varchar(100) NOT NULL DEFAULT '',
  `sCode` varchar(2) NOT NULL DEFAULT '',
  `iVersion` bigint(20) NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `iVersion` (`iVersion`),
  KEY `sCode` (`sCode`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `languages`
--

LOCK TABLES `languages` WRITE;
/*!40000 ALTER TABLE `languages` DISABLE KEYS */;
/*!40000 ALTER TABLE `languages` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `messages`
--

DROP TABLE IF EXISTS `messages`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `messages`
--

LOCK TABLES `messages` WRITE;
/*!40000 ALTER TABLE `messages` DISABLE KEYS */;
/*!40000 ALTER TABLE `messages` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `platforms`
--

DROP TABLE IF EXISTS `platforms`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `platforms`
--

LOCK TABLES `platforms` WRITE;
/*!40000 ALTER TABLE `platforms` DISABLE KEYS */;
/*!40000 ALTER TABLE `platforms` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `schema_revision`
--

DROP TABLE IF EXISTS `schema_revision`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `schema_revision` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `executed_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `file` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=105 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `schema_revision`
--

LOCK TABLES `schema_revision` WRITE;
/*!40000 ALTER TABLE `schema_revision` DISABLE KEYS */;
INSERT INTO `schema_revision` VALUES (1,'2018-10-24 06:11:11','1.0/revision-002/synchro_versions.sql'),(2,'2018-10-24 06:11:11','1.0/revision-003/history_tables.sql'),(3,'2018-10-24 06:11:11','1.0/revision-004/historyID_autoincrement.sql'),(4,'2018-10-24 06:11:11','1.0/revision-008/root_category.sql'),(5,'2018-10-24 06:11:11','1.0/revision-016/fix_item_sType.sql'),(6,'2018-10-24 06:11:11','1.0/revision-016/groups_tables.sql'),(7,'2018-10-24 06:11:11','1.0/revision-017/missing_fields.sql'),(8,'2018-10-24 06:11:11','1.0/revision-018/fix_history_groups.sql'),(9,'2018-10-24 06:11:11','1.0/revision-020/groups_sType.sql'),(10,'2018-10-24 06:11:11','1.0/revision-028/access_modes.sql'),(11,'2018-10-24 06:11:11','1.0/revision-029/access_solutions.sql'),(12,'2018-10-24 06:11:11','1.0/revision-029/validationType.sql'),(13,'2018-10-24 06:11:11','1.0/revision-031/drop_group_owners.sql'),(14,'2018-10-24 06:11:11','1.0/revision-031/users.sql'),(15,'2018-10-24 06:11:11','1.0/revision-032/isAdmin_fix.sql'),(16,'2018-10-24 06:11:11','1.0/revision-032/sNotify.sql'),(17,'2018-10-24 06:11:11','1.0/revision-032/user_groups.sql'),(18,'2018-10-24 06:11:11','1.0/revision-044/users_items.sql'),(19,'2018-10-24 06:11:11','1.0/revision-045/ancestors.sql'),(20,'2018-10-24 06:11:11','1.0/revision-055/index_ancestors.sql'),(21,'2018-10-24 06:11:11','1.0/revision-057/groups_items_access.sql'),(22,'2018-10-24 06:11:11','1.0/revision-060/items_ancestors_access.sql'),(23,'2018-10-24 06:11:11','1.0/revision-061/access_dates.sql'),(24,'2018-10-24 06:11:11','1.0/revision-066/nextVersion_null.sql'),(25,'2018-10-24 06:11:11','1.0/revision-139/completing_items_table.sql'),(26,'2018-10-24 06:11:11','1.0/revision-161/history_items.sql'),(27,'2018-10-24 06:11:11','1.0/revision-161/history_items_ancestors.sql'),(28,'2018-10-24 06:11:11','1.0/revision-212/user_item_index.sql'),(29,'2018-10-24 06:11:11','1.0/revision-227/groups_items_propagation.sql'),(30,'2018-10-24 06:11:11','1.0/revision-246/0-users_items_computations.sql'),(31,'2018-10-24 06:11:11','1.0/revision-246/items_groups_ancestors.sql'),(32,'2018-10-24 06:11:11','1.0/revision-246/optimizations.sql'),(33,'2018-10-24 06:11:11','1.0/revision-268/users_and_items.sql'),(34,'2018-10-24 06:11:11','1.0/revision-277/forum.sql'),(35,'2018-10-24 06:11:11','1.0/revision-321/manual_validation.sql'),(36,'2018-10-24 06:11:11','1.0/revision-321/platforms.sql'),(37,'2018-10-24 06:11:11','1.0/revision-321/users_answers.sql'),(38,'2018-10-24 06:11:11','1.0/revision-532/bugfixes.sql'),(39,'2018-10-24 06:11:11','1.0/revision-533/user_index.sql'),(40,'2018-10-24 06:11:11','1.0/revision-534/groups_items_index.sql'),(41,'2018-10-24 06:11:11','1.0/revision-536/index_item_item.sql'),(42,'2018-10-24 06:11:11','1.0/revision-537/default_propagation_users_items.sql'),(43,'2018-10-24 06:11:11','1.0/revision-538/task_url_index.sql'),(44,'2018-10-24 06:11:11','1.0/revision-539/item.bFullScreen.sql'),(45,'2018-10-24 06:11:11','1.0/revision-540/groups.sql'),(46,'2018-10-24 06:11:11','1.0/revision-541/items.sIconUrl.sql'),(47,'2018-10-24 06:11:11','1.0/revision-543/users.sRegistrationDate.sql'),(48,'2018-10-24 06:11:11','1.0/revision-544/limittedTimeContests.sql'),(49,'2018-10-24 06:11:11','1.0/revision-545/contest_adjustment.sql'),(50,'2018-10-24 06:11:11','1.0/revision-546/more_groups_groups_indexes.sql'),(51,'2018-10-24 06:11:11','1.0/revision-546/more_indexes.sql'),(52,'2018-10-24 06:11:11','1.0/revision-547/groups_defaults.sql'),(53,'2018-10-24 06:11:11','1.0/revision-548/items_bShowUserInfos.sql'),(54,'2018-10-24 06:11:11','1.0/revision-549/users_defaults.sql'),(55,'2018-10-24 06:11:32','1.1/revision-001/bugfix.sql'),(56,'2018-10-24 06:11:32','1.1/revision-002/platforms.sql'),(57,'2018-10-24 06:11:32','1.1/revision-003/drop_tmp.sql'),(58,'2018-10-24 06:11:33','1.1/revision-004/groups_stextid.sql'),(59,'2018-10-24 06:11:33','1.1/revision-004/unlocks.sql'),(60,'2018-10-24 06:11:33','1.1/revision-005/iScoreReeval.sql'),(61,'2018-10-24 06:11:33','1.1/revision-006/items_bReadOnly.sql'),(62,'2018-10-24 06:11:34','1.1/revision-007/fix_default_values.sql'),(63,'2018-10-24 06:11:34','1.1/revision-007/fix_history_bdeleted.sql'),(64,'2018-10-24 06:11:34','1.1/revision-007/fix_history_groups.sql'),(65,'2018-10-24 06:11:34','1.1/revision-007/fix_history_users.sql'),(66,'2018-10-24 06:11:34','1.1/revision-008/schema_revision.sql'),(67,'2018-10-24 06:11:34','1.1/revision-009/bFixedRanks.sql'),(68,'2018-10-24 06:11:34','1.1/revision-010/error_log.sql'),(69,'2018-10-24 06:11:35','1.1/revision-011/reducing_item_stype.sql'),(70,'2018-10-24 06:11:35','1.1/revision-012/fix_items_sDuration.sql'),(71,'2018-10-24 06:11:35','1.1/revision-012/users_items_sAnswer.sql'),(72,'2018-10-24 06:11:35','1.1/revision-013/items_idDefaultLanguage.sql'),(73,'2018-10-24 06:11:35','1.1/revision-014/default_values.sql'),(74,'2018-10-24 06:11:35','1.1/revision-014/default_values_bugfix.sql'),(75,'2018-10-24 06:11:35','1.1/revision-014/groups_fields.sql'),(76,'2018-10-24 06:11:35','1.1/revision-014/groups_login_prefixes.sql'),(77,'2018-10-24 06:11:36','1.1/revision-014/lm_prefix.sql'),(78,'2018-10-24 06:11:36','1.1/revision-015/groups_null_fileds.sql'),(79,'2018-10-24 06:11:36','1.1/revision-015/nulls.sql'),(80,'2018-10-24 06:11:36','1.1/revision-015/update_root_groups_textid.sql'),(81,'2018-10-24 06:11:36','1.1/revision-015/users.allowSubgroups.sql'),(82,'2018-10-24 06:11:36','1.1/revision-016/groupCodeEnter.sql'),(83,'2018-10-24 06:11:36','1.1/revision-016/items.sql'),(84,'2018-10-24 06:11:36','1.1/revision-016/text_varchar_bugfix.sql'),(85,'2018-10-24 06:11:37','1.1/revision-017/sHintsRequested.sql'),(86,'2018-10-24 06:11:37','1.1/revision-018/groups_add_teams.sql'),(87,'2018-10-24 06:11:37','1.1/revision-019/history_groups_add_teams.sql'),(88,'2018-10-24 06:11:37','1.1/revision-020/groups_add_teams.sql'),(89,'2018-10-24 06:11:37','1.1/revision-021/graduation_grade.sql'),(90,'2018-10-24 06:11:37','1.1/revision-021/groups_login_prefixes.sql'),(91,'2018-10-24 06:11:37','1.1/revision-021/user_creator_id.sql'),(92,'2018-10-24 06:11:38','1.1/revision-022/attempts.sql'),(93,'2018-10-24 06:11:38','1.1/revision-022/history_group_login_prefixes_index.sql'),(94,'2018-10-24 06:11:38','1.1/revision-023/indexes.sql'),(95,'2018-10-24 06:11:38','1.1/revision-024/indexes.sql'),(96,'2018-10-24 06:11:38','1.1/revision-024/remove_history_users_answers.sql'),(97,'2018-10-24 06:11:38','1.1/revision-025/bTeamsEditable.sql'),(98,'2018-10-24 06:11:38','1.1/revision-026/sBestAnswerDate.sql'),(99,'2018-10-24 06:11:38','1.1/revision-027/badges.sql'),(100,'2018-10-24 06:11:38','1.1/revision-028/lockUserDeletionDate.sql'),(101,'2018-10-24 06:11:39','1.1/revision-028/nulls.sql'),(102,'2018-10-24 06:11:39','1.1/revision-029/items_sRepositoryPath.sql'),(103,'2018-10-24 06:11:39','1.1/revision-029/platform_baseUrl.sql'),(104,'2018-10-24 06:11:39','1.1/revision-029/user_items_platform_data.sql');
/*!40000 ALTER TABLE `schema_revision` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `synchro_version`
--

DROP TABLE IF EXISTS `synchro_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `synchro_version` (
  `ID` tinyint(1) NOT NULL,
  `iVersion` int(11) NOT NULL,
  `iLastServerVersion` int(11) NOT NULL,
  `iLastClientVersion` int(11) NOT NULL,
  PRIMARY KEY (`ID`),
  KEY `iVersion` (`iVersion`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `synchro_version`
--

LOCK TABLES `synchro_version` WRITE;
/*!40000 ALTER TABLE `synchro_version` DISABLE KEYS */;
INSERT INTO `synchro_version` VALUES (0,0,0,0);
/*!40000 ALTER TABLE `synchro_version` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `threads`
--

DROP TABLE IF EXISTS `threads`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `threads`
--

LOCK TABLES `threads` WRITE;
/*!40000 ALTER TABLE `threads` DISABLE KEYS */;
/*!40000 ALTER TABLE `threads` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users_answers`
--

DROP TABLE IF EXISTS `users_answers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
  KEY `idItem` (`idItem`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users_answers`
--

LOCK TABLES `users_answers` WRITE;
/*!40000 ALTER TABLE `users_answers` DISABLE KEYS */;
/*!40000 ALTER TABLE `users_answers` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users_items`
--

DROP TABLE IF EXISTS `users_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
  `sAdditionalTime` datetime DEFAULT NULL,
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
  KEY `idUser` (`idUser`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users_items`
--

LOCK TABLES `users_items` WRITE;
/*!40000 ALTER TABLE `users_items` DISABLE KEYS */;
/*!40000 ALTER TABLE `users_items` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users_threads`
--

DROP TABLE IF EXISTS `users_threads`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
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
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users_threads`
--

LOCK TABLES `users_threads` WRITE;
/*!40000 ALTER TABLE `users_threads` DISABLE KEYS */;
/*!40000 ALTER TABLE `users_threads` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2018-10-24  8:45:43

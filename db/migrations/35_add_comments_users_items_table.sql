-- +migrate Up
ALTER TABLE `users_items`
  COMMENT 'Information about the activity of users on items',
  MODIFY COLUMN `idAttemptActive` bigint(20) DEFAULT NULL COMMENT 'Current attempt selected by this user.',
  MODIFY COLUMN `iScore` float NOT NULL DEFAULT '0' COMMENT 'Current score for this attempt ; can be a cached computation',
  MODIFY COLUMN `iScoreComputed` float NOT NULL DEFAULT '0' COMMENT 'Deprecated',
  MODIFY COLUMN `iScoreReeval` float DEFAULT '0' COMMENT 'Deprecated',
  MODIFY COLUMN `iScoreDiffManual` float NOT NULL DEFAULT '0' COMMENT 'How much did we manually add to the computed score',
  MODIFY COLUMN `sScoreDiffComment` varchar(200) NOT NULL DEFAULT '' COMMENT 'Why was the score manually changed ?',
  MODIFY COLUMN `nbSubmissionsAttempts` int(11) NOT NULL DEFAULT '0' COMMENT 'How many submissions in total for this item and its children?',
  MODIFY COLUMN `nbTasksTried` int(11) NOT NULL DEFAULT '0' COMMENT 'Deprecated',
  MODIFY COLUMN `nbTasksSolved` int(11) NOT NULL DEFAULT '0' COMMENT 'Deprecated',
  MODIFY COLUMN `nbChildrenValidated` int(11) NOT NULL DEFAULT '0' COMMENT 'Deprecated',
  MODIFY COLUMN `bValidated` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Deprecated',
  MODIFY COLUMN `bFinished` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Deprecated',
  MODIFY COLUMN `bKeyObtained` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the user obtained the key on this item. Changed to 1 if the user gets a score >= items.iScoreMinUnlock, will grant access to new item from items.idItemUnlocked. This information is propagated to users_items.',
  MODIFY COLUMN `nbTasksWithHelp` int(11) NOT NULL DEFAULT '0' COMMENT 'For how many of this item''s descendants tasks within this attempts, did the user ask for hints (or help on the forum - not implemented)?',
  MODIFY COLUMN `sHintsRequested` mediumtext COMMENT 'Deprecated',
  MODIFY COLUMN `nbHintsCached` int(11) NOT NULL DEFAULT '0' COMMENT 'Deprecated',
  MODIFY COLUMN `nbCorrectionsRead` int(11) NOT NULL DEFAULT '0' COMMENT 'Number of solutions the user read among the descendants of this item.',
  MODIFY COLUMN `iPrecision` int(11) NOT NULL DEFAULT '0' COMMENT 'Precision (based on a formula to be defined) of the user recently, when working on this item and its descendants.',
  MODIFY COLUMN `iAutonomy` int(11) NOT NULL DEFAULT '0' COMMENT 'Autonomy (based on a formula to be defined) of the user was recently, when working on this item and its descendants: how much help / hints he used.',
  MODIFY COLUMN `sStartDate` datetime DEFAULT NULL COMMENT 'Deprecated',
  MODIFY COLUMN `sValidationDate` datetime DEFAULT NULL COMMENT 'Deprecated',
  MODIFY COLUMN `sFinishDate` datetime DEFAULT NULL COMMENT 'Deprecated',
  MODIFY COLUMN `sLastActivityDate` datetime DEFAULT NULL COMMENT 'When was the last activity within this task.',
  MODIFY COLUMN `sThreadStartDate` datetime DEFAULT NULL COMMENT 'When was a discussion thread started by this user/group on the forum',
  MODIFY COLUMN `sBestAnswerDate` datetime DEFAULT NULL COMMENT 'Deprecated',
  MODIFY COLUMN `sLastAnswerDate` datetime DEFAULT NULL COMMENT 'Deprecated',
  MODIFY COLUMN `sLastHintDate` datetime DEFAULT NULL COMMENT 'Deprecated',
  MODIFY COLUMN `sContestStartDate` datetime DEFAULT NULL COMMENT 'Deprecated',
  MODIFY COLUMN `bRanked` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Deprecated',
  MODIFY COLUMN `sAllLangProg` varchar(200) DEFAULT NULL COMMENT 'List of programming languages used',
  MODIFY COLUMN `sAncestorsComputationState` enum('done','processing','todo','temp') NOT NULL DEFAULT 'todo' COMMENT 'Used to denote whether the ancestors data have to be recomputed (after this item''s score was changed for instance)',
  MODIFY COLUMN `sState` mediumtext COMMENT 'Deprecated',
  MODIFY COLUMN `sAnswer` mediumtext COMMENT 'Deprecated'
  ;

-- +migrate Down
ALTER TABLE `users_items`
  COMMENT '',
  MODIFY COLUMN `idAttemptActive` bigint(20) DEFAULT NULL,
  MODIFY COLUMN `iScore` float NOT NULL DEFAULT '0',
  MODIFY COLUMN `iScoreComputed` float NOT NULL DEFAULT '0',
  MODIFY COLUMN `iScoreReeval` float DEFAULT '0',
  MODIFY COLUMN `iScoreDiffManual` float NOT NULL DEFAULT '0',
  MODIFY COLUMN `sScoreDiffComment` varchar(200) NOT NULL DEFAULT '',
  MODIFY COLUMN `nbSubmissionsAttempts` int(11) NOT NULL DEFAULT '0',
  MODIFY COLUMN `nbTasksTried` int(11) NOT NULL DEFAULT '0',
  MODIFY COLUMN `nbTasksSolved` int(11) NOT NULL DEFAULT '0',
  MODIFY COLUMN `nbChildrenValidated` int(11) NOT NULL DEFAULT '0',
  MODIFY COLUMN `bValidated` tinyint(1) NOT NULL DEFAULT '0',
  MODIFY COLUMN `bFinished` tinyint(1) NOT NULL DEFAULT '0',
  MODIFY COLUMN `bKeyObtained` tinyint(1) NOT NULL DEFAULT '0',
  MODIFY COLUMN `nbTasksWithHelp` int(11) NOT NULL DEFAULT '0',
  MODIFY COLUMN `sHintsRequested` mediumtext,
  MODIFY COLUMN `nbHintsCached` int(11) NOT NULL DEFAULT '0',
  MODIFY COLUMN `nbCorrectionsRead` int(11) NOT NULL DEFAULT '0',
  MODIFY COLUMN `iPrecision` int(11) NOT NULL DEFAULT '0',
  MODIFY COLUMN `iAutonomy` int(11) NOT NULL DEFAULT '0',
  MODIFY COLUMN `sStartDate` datetime DEFAULT NULL,
  MODIFY COLUMN `sValidationDate` datetime DEFAULT NULL,
  MODIFY COLUMN `sFinishDate` datetime DEFAULT NULL,
  MODIFY COLUMN `sLastActivityDate` datetime DEFAULT NULL,
  MODIFY COLUMN `sThreadStartDate` datetime DEFAULT NULL,
  MODIFY COLUMN `sBestAnswerDate` datetime DEFAULT NULL,
  MODIFY COLUMN `sLastAnswerDate` datetime DEFAULT NULL,
  MODIFY COLUMN `sLastHintDate` datetime DEFAULT NULL,
  MODIFY COLUMN `sContestStartDate` datetime DEFAULT NULL,
  MODIFY COLUMN `bRanked` tinyint(1) NOT NULL DEFAULT '0',
  MODIFY COLUMN `sAllLangProg` varchar(200) DEFAULT NULL,
  MODIFY COLUMN `sAncestorsComputationState` enum('done','processing','todo','temp') NOT NULL DEFAULT 'todo',
  MODIFY COLUMN `sState` mediumtext,
  MODIFY COLUMN `sAnswer` mediumtext
  ;

-- +migrate Up
ALTER TABLE `groups`
  COMMENT 'A group can be either a user, a set of users, or a set of groups.',
  MODIFY COLUMN `sTextId` varchar(255) NOT NULL DEFAULT '' COMMENT 'Internal text ID for special groups. Used to refer o them and avoid breaking features if an admin renames the group',
  MODIFY COLUMN `iGrade` int(4) NOT NULL DEFAULT '-2' COMMENT 'For some types of groups, indicate which grade the users belong to.',
  MODIFY COLUMN `sGradeDetails` varchar(50) DEFAULT NULL COMMENT 'Explanations about the grade',
  MODIFY COLUMN `sDescription` text COMMENT 'Purpose of this group. Will be visible by its members. or by the public if the group is public.',
  MODIFY COLUMN `sDateCreated` datetime DEFAULT NULL,
  MODIFY COLUMN `bOpened` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Can users still join this group or request to join it?',
  MODIFY COLUMN `bFreeAccess` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Can users search for this group and ask to join it?',
  MODIFY COLUMN `idTeamItem` bigint(20) DEFAULT NULL COMMENT 'If this group is a team, what item is it attached to?',
  MODIFY COLUMN `iTeamParticipating` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Did the team start the item it is associated to (from idTeamItem)?',
  MODIFY COLUMN `sCode` varchar(50) DEFAULT NULL COMMENT 'Code that can be used to join the group (if it is opened)',
  MODIFY COLUMN `sCodeTimer` time DEFAULT NULL COMMENT 'How long after the first use of the password it will expire',
  MODIFY COLUMN `sCodeEnd` datetime DEFAULT NULL COMMENT 'When the password expires. Set when it is first used.',
  MODIFY COLUMN `sRedirectPath` text COMMENT 'Where the user should be sent when joining this group. For now it is a path to be used in the url.',
  MODIFY COLUMN `bOpenContest` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'If true and the group is associated through sRedirectPath with an item that is a contest, the contest should be started for this user as soon as he joins the group.',
  MODIFY COLUMN `lockUserDeletionDate` date DEFAULT NULL COMMENT 'Prevent users from this group to delete their own user themselves until this date',
  DROP COLUMN `sAncestorsComputationState`;


-- +migrate Down
ALTER TABLE `groups`
  COMMENT '',
  MODIFY COLUMN `sTextId` varchar(255) NOT NULL DEFAULT '',
  MODIFY COLUMN `iGrade` int(4) NOT NULL DEFAULT '-2',
  MODIFY COLUMN `sGradeDetails` varchar(50) DEFAULT NULL,
  MODIFY COLUMN `sDescription` text,
  MODIFY COLUMN `sDateCreated` datetime DEFAULT NULL,
  MODIFY COLUMN `bOpened` tinyint(1) NOT NULL DEFAULT '0',
  MODIFY COLUMN `bFreeAccess` tinyint(1) NOT NULL DEFAULT '0',
  MODIFY COLUMN `idTeamItem` bigint(20) DEFAULT NULL,
  MODIFY COLUMN `iTeamParticipating` tinyint(1) NOT NULL DEFAULT '0',
  MODIFY COLUMN `sCode` varchar(50) DEFAULT NULL,
  MODIFY COLUMN `sCodeTimer` time DEFAULT NULL,
  MODIFY COLUMN `sCodeEnd` datetime DEFAULT NULL,
  MODIFY COLUMN `sRedirectPath` text,
  MODIFY COLUMN `bOpenContest` tinyint(1) NOT NULL DEFAULT '0',
  MODIFY COLUMN `lockUserDeletionDate` date DEFAULT NULL,
  ADD COLUMN `sAncestorsComputationState` enum('done','processing','todo') NOT NULL DEFAULT 'todo' AFTER `bSendEmails`;

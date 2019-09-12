-- +migrate Up
ALTER TABLE `users_answers`
  COMMENT 'All the submissions made by users on tasks, as well as saved answers and the current answer',
  MODIFY COLUMN `sState` mediumtext COMMENT 'Saved state (sent by the task platform)',
  MODIFY COLUMN `sAnswer` mediumtext COMMENT 'Saved answer (sent by the task platform)',
  MODIFY COLUMN `sLangProg` varchar(50) DEFAULT NULL COMMENT 'Programming language of this submission',
  MODIFY COLUMN `sSubmissionDate` datetime NOT NULL COMMENT 'Submission time',
  MODIFY COLUMN `iScore` float DEFAULT NULL COMMENT 'Score obtained',
  MODIFY COLUMN `bValidated` tinyint(1) DEFAULT NULL COMMENT 'Whether it is considered "validated" (above validation threshold for this item)',
  MODIFY COLUMN `sGradingDate` datetime DEFAULT NULL COMMENT 'When was it last graded',
  MODIFY COLUMN `idUserGrader` int(11) DEFAULT NULL COMMENT 'Who did the last grading'
  ;

-- +migrate Down
ALTER TABLE `users_answers`
  COMMENT '',
  MODIFY COLUMN `sState` mediumtext,
  MODIFY COLUMN `sAnswer` mediumtext,
  MODIFY COLUMN `sLangProg` varchar(50) DEFAULT NULL,
  MODIFY COLUMN `sSubmissionDate` datetime NOT NULL,
  MODIFY COLUMN `iScore` float DEFAULT NULL,
  MODIFY COLUMN `bValidated` tinyint(1) DEFAULT NULL,
  MODIFY COLUMN `sGradingDate` datetime DEFAULT NULL,
  MODIFY COLUMN `idUserGrader` int(11) DEFAULT NULL
  ;

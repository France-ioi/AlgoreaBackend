-- +migrate Up
ALTER TABLE `users`
  COMMENT 'Users. A large part is obtained from the auth platform and may not be manually edited',
  MODIFY COLUMN `loginID` bigint(20) DEFAULT NULL COMMENT '"userId" returned by the auth platform',
  MODIFY COLUMN `tempUser` tinyint(1) NOT NULL COMMENT 'Whether it is a temporary user. If so, the user will be deleted soon.',
  MODIFY COLUMN `sLogin` varchar(100) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT 'login provided by the auth platform',
  MODIFY COLUMN `sOpenIdIdentity` varchar(255) DEFAULT NULL COMMENT 'User''s Open Id Identity',
  MODIFY COLUMN `sRegistrationDate` datetime DEFAULT NULL COMMENT 'When the user first connected to this platform',
  MODIFY COLUMN `sEmail` varchar(100) DEFAULT NULL COMMENT 'E-mail, provided by auth platform',
  MODIFY COLUMN `bEmailVerified` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether email has been verified, provided by auth platform',
  MODIFY COLUMN `sFirstName` varchar(100) DEFAULT NULL COMMENT 'First name, provided by auth platform',
  MODIFY COLUMN `sLastName` varchar(100) DEFAULT NULL COMMENT 'Last name, provided by auth platform',
  MODIFY COLUMN `sStudentId` text COMMENT 'A student id provided by the school, provided by auth platform',
  MODIFY COLUMN `sCountryCode` char(3) NOT NULL DEFAULT '' COMMENT '3-letter country code',
  MODIFY COLUMN `sTimeZone` varchar(100) DEFAULT NULL COMMENT 'Time zone, provided by auth platform',
  MODIFY COLUMN `sBirthDate` date DEFAULT NULL COMMENT 'Date of birth, provided by auth platform',
  MODIFY COLUMN `iGraduationYear` int(11) NOT NULL DEFAULT '0' COMMENT 'High school graduation year',
  MODIFY COLUMN `iGrade` int(11) DEFAULT NULL COMMENT 'School grade, provided by auth platform',
  MODIFY COLUMN `sSex` enum('Male','Female') DEFAULT NULL COMMENT 'Gender, provided by auth platform',
  MODIFY COLUMN `sAddress` mediumtext COMMENT 'Address, provided by auth platform',
  MODIFY COLUMN `sZipcode` longtext COMMENT 'Zip code, provided by auth platform',
  MODIFY COLUMN `sCity` longtext COMMENT 'City, provided by auth platform',
  MODIFY COLUMN `sLandLineNumber` longtext COMMENT 'Phone number, provided by auth platform',
  MODIFY COLUMN `sCellPhoneNumber` longtext COMMENT 'Mobile phone number, provided by auth platform',
  MODIFY COLUMN `sDefaultLanguage` char(3) NOT NULL DEFAULT 'fr' COMMENT 'Current language used to display content. Initial version provided by auth platform, then can be changed manually.',
  MODIFY COLUMN `bNotifyNews` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Whether the user accepts that we send emails about events related to the platform',
  MODIFY COLUMN `sNotify` enum('Never','Answers','Concerned') NOT NULL DEFAULT 'Answers' COMMENT 'When we should send an email to the user. Answers: when someone posts a message on a thread created by the user. Concerned: when someone post a message on a thread that the user participated in',
  MODIFY COLUMN `bPublicFirstName` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Whether show user''s first name in his public profile',
  MODIFY COLUMN `bPublicLastName` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Whether show user''s last name in his public profile',
  MODIFY COLUMN `sFreeText` mediumtext COMMENT 'Text provided by the user, to be displayed on his public profile',
  MODIFY COLUMN `sWebSite` varchar(100) DEFAULT NULL COMMENT 'Link to the user''s website, to be displayed on his public profile',
  MODIFY COLUMN `bPhotoAutoload` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Indicates that the user has a picture associated with his profile. Not used yet.',
  MODIFY COLUMN `sLangProg` varchar(30) DEFAULT 'Python' COMMENT 'Current programming language selected by the user (to display the corresponding version of tasks)',
  MODIFY COLUMN `sLastLoginDate` datetime DEFAULT NULL COMMENT 'When is the last time this user logged in on the platform',
  MODIFY COLUMN `sLastActivityDate` datetime DEFAULT NULL COMMENT 'Last activity time on the platform (any action)',
  MODIFY COLUMN `sLastIP` varchar(16) DEFAULT NULL COMMENT 'Last IP (to detect cheaters).',
  MODIFY COLUMN `bBasicEditorMode` tinyint(4) NOT NULL DEFAULT '1' COMMENT 'Which editor should be used in programming tasks.',
  MODIFY COLUMN `nbSpacesForTab` int(11) NOT NULL DEFAULT '3' COMMENT 'How many spaces for a tabulation, in programming tasks.',
  MODIFY COLUMN `iMemberState` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'On old website, indicates if the user is a member of France-ioi',
  MODIFY COLUMN `iStepLevelInSite` int(11) NOT NULL DEFAULT '0' COMMENT 'User''s level',
  MODIFY COLUMN `bIsAdmin` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Is the user an admin? Not used?',
  MODIFY COLUMN `bNoRanking` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Whether this user should not be listed when displaying the results of contests, or points obtained on the platform',
  MODIFY COLUMN `nbHelpGiven` int(11) NOT NULL DEFAULT '0' COMMENT 'How many times did the user help others (# of discussions)',
  MODIFY COLUMN `idGroupSelf` bigint(20) DEFAULT NULL COMMENT 'Group that represents this user',
  MODIFY COLUMN `idGroupOwned` bigint(20) DEFAULT NULL COMMENT 'Group that will contain groups that this user manages',
  MODIFY COLUMN `sNotificationReadDate` datetime DEFAULT NULL COMMENT 'When the user last read notifications',
  MODIFY COLUMN `loginModulePrefix` varchar(100) DEFAULT NULL COMMENT 'Set to enable login module accounts manager',
  MODIFY COLUMN `creatorID` bigint(20) DEFAULT NULL COMMENT 'Which user created a given login with the login generation tool',
  MODIFY COLUMN `allowSubgroups` tinyint(4) DEFAULT NULL COMMENT 'Allow to create subgroups';

-- +migrate Down
ALTER TABLE `users`
  COMMENT '',
  MODIFY COLUMN `loginID` bigint(20) DEFAULT NULL COMMENT 'the ''userId'' returned by login platform',
  MODIFY COLUMN `tempUser` tinyint(1) NOT NULL,
  MODIFY COLUMN `sLogin` varchar(100) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '',
  MODIFY COLUMN `sOpenIdIdentity` varchar(255) DEFAULT NULL COMMENT 'User''s Open Id Identity',
  MODIFY COLUMN `sRegistrationDate` datetime DEFAULT NULL,
  MODIFY COLUMN `sEmail` varchar(100) DEFAULT NULL,
  MODIFY COLUMN `bEmailVerified` tinyint(1) NOT NULL DEFAULT '0',
  MODIFY COLUMN `sFirstName` varchar(100) DEFAULT NULL COMMENT 'User''s first name',
  MODIFY COLUMN `sLastName` varchar(100) DEFAULT NULL COMMENT 'User''s last name',
  MODIFY COLUMN `sStudentId` text,
  MODIFY COLUMN `sCountryCode` char(3) NOT NULL DEFAULT '',
  MODIFY COLUMN `sTimeZone` varchar(100) DEFAULT NULL,
  MODIFY COLUMN `sBirthDate` date DEFAULT NULL COMMENT 'User''s birth date',
  MODIFY COLUMN `iGraduationYear` int(11) NOT NULL DEFAULT '0' COMMENT 'User''s high school graduation year',
  MODIFY COLUMN `iGrade` int(11) DEFAULT NULL,
  MODIFY COLUMN `sSex` enum('Male','Female') DEFAULT NULL,
  MODIFY COLUMN `sAddress` mediumtext COMMENT 'User''s address',
  MODIFY COLUMN `sZipcode` longtext COMMENT 'User''s postal code',
  MODIFY COLUMN `sCity` longtext COMMENT 'User''s city',
  MODIFY COLUMN `sLandLineNumber` longtext COMMENT 'User''s phone number',
  MODIFY COLUMN `sCellPhoneNumber` longtext COMMENT 'User''s mobil phone number',
  MODIFY COLUMN `sDefaultLanguage` char(3) NOT NULL DEFAULT 'fr' COMMENT 'User''s default language',
  MODIFY COLUMN `bNotifyNews` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'TODO',
  MODIFY COLUMN `sNotify` enum('Never','Answers','Concerned') NOT NULL DEFAULT 'Answers',
  MODIFY COLUMN `bPublicFirstName` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Publicly show user''s first name',
  MODIFY COLUMN `bPublicLastName` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'Publicly show user''s first name',
  MODIFY COLUMN `sFreeText` mediumtext,
  MODIFY COLUMN `sWebSite` varchar(100) DEFAULT NULL,
  MODIFY COLUMN `bPhotoAutoload` tinyint(1) NOT NULL DEFAULT '0',
  MODIFY COLUMN `sLangProg` varchar(30) DEFAULT 'Python',
  MODIFY COLUMN `sLastLoginDate` datetime DEFAULT NULL,
  MODIFY COLUMN `sLastActivityDate` datetime DEFAULT NULL COMMENT 'User''s last activity time on the website',
  MODIFY COLUMN `sLastIP` varchar(16) DEFAULT NULL,
  MODIFY COLUMN `bBasicEditorMode` tinyint(4) NOT NULL DEFAULT '1',
  MODIFY COLUMN `nbSpacesForTab` int(11) NOT NULL DEFAULT '3',
  MODIFY COLUMN `iMemberState` tinyint(4) NOT NULL DEFAULT '0',
  MODIFY COLUMN `iStepLevelInSite` int(11) NOT NULL DEFAULT '0' COMMENT 'User''s level',
  MODIFY COLUMN `bIsAdmin` tinyint(4) NOT NULL DEFAULT '0',
  MODIFY COLUMN `bNoRanking` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'TODO',
  MODIFY COLUMN `nbHelpGiven` int(11) NOT NULL DEFAULT '0' COMMENT 'TODO',
  MODIFY COLUMN `idGroupSelf` bigint(20) DEFAULT NULL,
  MODIFY COLUMN `idGroupOwned` bigint(20) DEFAULT NULL,
  MODIFY COLUMN `sNotificationReadDate` datetime DEFAULT NULL,
  MODIFY COLUMN `loginModulePrefix` varchar(100) DEFAULT NULL COMMENT 'Set to enable login module accounts manager',
  MODIFY COLUMN `creatorID` bigint(20) DEFAULT NULL COMMENT 'which user created a given login with the login generation tool',
  MODIFY COLUMN `allowSubgroups` tinyint(4) DEFAULT NULL COMMENT 'Allow to create subgroups';

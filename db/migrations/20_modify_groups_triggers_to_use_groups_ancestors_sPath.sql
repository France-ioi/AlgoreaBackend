-- +migrate Up
DROP TRIGGER `after_insert_groups`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50017 DEFINER=`algorea`@`%`*/ /*!50003 TRIGGER `after_insert_groups`
AFTER INSERT ON `groups` FOR EACH ROW BEGIN
  INSERT INTO `history_groups` (
    `ID`,`iVersion`,`sName`,`iGrade`,`sGradeDetails`,`sDescription`,`sDateCreated`,`bOpened`,`bFreeAccess`,`idTeamItem`,
    `iTeamParticipating`,`sPassword`,`sPasswordTimer`,`sPasswordEnd`,`sRedirectPath`,`bOpenContest`,`sType`,`bSendEmails`
  ) VALUES (
    NEW.`ID`,@curVersion,NEW.`sName`,NEW.`iGrade`,NEW.`sGradeDetails`,NEW.`sDescription`,NEW.`sDateCreated`,
    NEW.`bOpened`,NEW.`bFreeAccess`,NEW.`idTeamItem`,NEW.`iTeamParticipating`,NEW.`sPassword`,NEW.`sPasswordTimer`,
    NEW.`sPasswordEnd`,NEW.`sRedirectPath`,NEW.`bOpenContest`,NEW.`sType`,NEW.`bSendEmails`
  );
  INSERT INTO `groups_ancestors` (`idGroupAncestor`, `idGroupChild`, `bIsSelf`, `sPath`)
    VALUES (NEW.`ID`, NEW.`ID`, 1, CONCAT('/', NEW.`ID`, '/'));
END; */
-- +migrate StatementEnd
DROP TRIGGER `after_delete_groups`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50017 DEFINER=`algorea`@`%`*/ /*!50003 TRIGGER `after_delete_groups`
AFTER DELETE ON `groups` FOR EACH ROW BEGIN
  DELETE FROM `groups_groups` WHERE `idGroupParent` = OLD.`ID`;
  DELETE FROM `groups_groups` WHERE `idGroupChild` = OLD.`ID`;
END; */
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER `after_insert_groups`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50017 DEFINER=`algorea`@`%`*/ /*!50003 TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN INSERT INTO `history_groups` (`ID`,`iVersion`,`sName`,`iGrade`,`sGradeDetails`,`sDescription`,`sDateCreated`,`bOpened`,`bFreeAccess`,`idTeamItem`,`iTeamParticipating`,`sPassword`,`sPasswordTimer`,`sPasswordEnd`,`sRedirectPath`,`bOpenContest`,`sType`,`bSendEmails`) VALUES (NEW.`ID`,@curVersion,NEW.`sName`,NEW.`iGrade`,NEW.`sGradeDetails`,NEW.`sDescription`,NEW.`sDateCreated`,NEW.`bOpened`,NEW.`bFreeAccess`,NEW.`idTeamItem`,NEW.`iTeamParticipating`,NEW.`sPassword`,NEW.`sPasswordTimer`,NEW.`sPasswordEnd`,NEW.`sRedirectPath`,NEW.`bOpenContest`,NEW.`sType`,NEW.`bSendEmails`); INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (NEW.`ID`, 'todo') ; END; */
-- +migrate StatementEnd
DROP TRIGGER `after_delete_groups`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50017 DEFINER=`algorea`@`%`*/ /*!50003 TRIGGER `after_delete_groups` AFTER DELETE ON `groups` FOR EACH ROW BEGIN DELETE FROM groups_propagate where ID = OLD.ID ; END; */
-- +migrate StatementEnd

-- +migrate Up
ALTER TABLE `groups_ancestors` DROP KEY `idGroupAncestor`;
ALTER TABLE `groups_ancestors` ADD INDEX `idGroupAncestor` (`idGroupAncestor`, `idGroupChild`);
ALTER TABLE `groups_ancestors` ADD COLUMN `sPath` VARCHAR(2048) CHARACTER SET latin1 COLLATE latin1_bin;
ALTER TABLE `history_groups_ancestors` ADD COLUMN `sPath` VARCHAR(2048) CHARACTER SET latin1 COLLATE latin1_bin;
DELETE FROM `groups_ancestors` WHERE !`bIsSelf`;
DROP TRIGGER `after_insert_groups_ancestors`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50017 DEFINER=`algorea`@`%`*/ /*!50003 TRIGGER `after_insert_groups_ancestors`
AFTER INSERT ON `groups_ancestors` FOR EACH ROW BEGIN
  INSERT INTO `history_groups_ancestors` (`ID`,`iVersion`,`idGroupAncestor`,`idGroupChild`,`bIsSelf`,`sPath`)
    VALUES (NEW.`ID`,@curVersion,NEW.`idGroupAncestor`,NEW.`idGroupChild`,NEW.`bIsSelf`,NEW.`sPath`);
END; */
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_ancestors`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50017 DEFINER=`algorea`@`%`*/ /*!50003 TRIGGER `before_update_groups_ancestors`
BEFORE UPDATE ON `groups_ancestors` FOR EACH ROW BEGIN
  IF NEW.iVersion <> OLD.iVersion THEN
    SET @curVersion = NEW.iVersion; ELSE SELECT (UNIX_TIMESTAMP() * 10) INTO @curVersion;
  END IF;
  IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idGroupAncestor` <=> NEW.`idGroupAncestor` AND
          OLD.`idGroupChild` <=> NEW.`idGroupChild` AND OLD.`bIsSelf` <=> NEW.`bIsSelf` AND
          OLD.`sPath` <=> NEW.`sPath`) THEN
    SET NEW.iVersion = @curVersion;
    UPDATE `history_groups_ancestors` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;
    INSERT INTO `history_groups_ancestors` (`ID`,`iVersion`,`idGroupAncestor`,`idGroupChild`,`bIsSelf`,`sPath`)
      VALUES (NEW.`ID`,@curVersion,NEW.`idGroupAncestor`,NEW.`idGroupChild`,NEW.`bIsSelf`,NEW.`sPath`);
  END IF;
END; */
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_ancestors`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50017 DEFINER=`algorea`@`%`*/ /*!50003 TRIGGER `before_delete_groups_ancestors`
BEFORE DELETE ON `groups_ancestors` FOR EACH ROW BEGIN
  SELECT (UNIX_TIMESTAMP() * 10) INTO @curVersion;
  UPDATE `history_groups_ancestors` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;
  INSERT INTO `history_groups_ancestors` (
    `ID`,`iVersion`,`idGroupAncestor`,`idGroupChild`,`bIsSelf`, `sPath`, `bDeleted`
  ) VALUES (OLD.`ID`,@curVersion,OLD.`idGroupAncestor`,OLD.`idGroupChild`,OLD.`bIsSelf`,OLD.`sPath`,1);
END; */
-- +migrate StatementEnd
UPDATE `groups_ancestors` SET `sPath` = CONCAT('/', `idGroupAncestor`, '/');
-- +migrate StatementBegin
CREATE PROCEDURE tmp_populate_groups_ancestors_path()
BEGIN
  DECLARE pattern VARCHAR(2048) DEFAULT '/';
  REPEAT
    SET pattern = CONCAT(pattern, '%/');
    INSERT INTO `groups_ancestors` (`idGroupAncestor`, `idGroupChild`, `bIsSelf`, `sPath`) (
      SELECT `groups_ancestors`.`idGroupAncestor`, `groups_groups`.`idGroupChild`, 0,
             CONCAT(`groups_ancestors`.`sPath`, `groups_groups`.`idGroupChild`, '/')
      FROM `groups_ancestors`
             JOIN `groups_groups` ON `groups_groups`.`idGroupParent` = `groups_ancestors`.`idGroupChild`
      WHERE `groups_ancestors`.`sPath` LIKE pattern
    );
  UNTIL ROW_COUNT() = 0
  END REPEAT;
END;
-- +migrate StatementEnd
CALL tmp_populate_groups_ancestors_path;
DROP PROCEDURE tmp_populate_groups_ancestors_path;

-- +migrate Down
DELETE `groups_ancestors`
  FROM `groups_ancestors`
  JOIN `groups_ancestors` AS `duplicates` USING(`idGroupAncestor`, `idGroupChild`)
  WHERE `groups_ancestors`.`ID` > `duplicates`.`ID`;

ALTER TABLE `groups_ancestors` DROP COLUMN `sPath`;
ALTER TABLE `history_groups_ancestors` DROP COLUMN `sPath`;
DROP TRIGGER `after_insert_groups_ancestors`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50017 DEFINER=`algorea`@`%`*/ /*!50003 TRIGGER `after_insert_groups_ancestors` AFTER INSERT ON `groups_ancestors` FOR EACH ROW BEGIN INSERT INTO `history_groups_ancestors` (`ID`,`iVersion`,`idGroupAncestor`,`idGroupChild`,`bIsSelf`) VALUES (NEW.`ID`,@curVersion,NEW.`idGroupAncestor`,NEW.`idGroupChild`,NEW.`bIsSelf`); END; */
-- +migrate StatementEnd
DROP TRIGGER `before_update_groups_ancestors`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50017 DEFINER=`algorea`@`%`*/ /*!50003 TRIGGER `before_update_groups_ancestors` BEFORE UPDATE ON `groups_ancestors` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT (UNIX_TIMESTAMP() * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idGroupAncestor` <=> NEW.`idGroupAncestor` AND OLD.`idGroupChild` <=> NEW.`idGroupChild` AND OLD.`bIsSelf` <=> NEW.`bIsSelf`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_groups_ancestors` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_groups_ancestors` (`ID`,`iVersion`,`idGroupAncestor`,`idGroupChild`,`bIsSelf`)       VALUES (NEW.`ID`,@curVersion,NEW.`idGroupAncestor`,NEW.`idGroupChild`,NEW.`bIsSelf`) ; END IF; END; */
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_ancestors`;
-- +migrate StatementBegin
/*!50003 CREATE*/ /*!50017 DEFINER=`algorea`@`%`*/ /*!50003 TRIGGER `before_delete_groups_ancestors` BEFORE DELETE ON `groups_ancestors` FOR EACH ROW BEGIN SELECT (UNIX_TIMESTAMP() * 10) INTO @curVersion; UPDATE `history_groups_ancestors` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_groups_ancestors` (`ID`,`iVersion`,`idGroupAncestor`,`idGroupChild`,`bIsSelf`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idGroupAncestor`,OLD.`idGroupChild`,OLD.`bIsSelf`, 1); END; */
-- +migrate StatementEnd

-- +migrate Up
DROP TRIGGER `before_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_groups`
BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN
  IF NEW.iVersion <> OLD.iVersion THEN
    SET @curVersion = NEW.iVersion;
  ELSE
    SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
  END IF;
  IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idGroupParent` <=> NEW.`idGroupParent` AND
          OLD.`idGroupChild` <=> NEW.`idGroupChild` AND OLD.`iChildOrder` <=> NEW.`iChildOrder`AND
          OLD.`sType` <=> NEW.`sType` AND OLD.`sRole` <=> NEW.`sRole` AND OLD.`sStatusDate` <=> NEW.`sStatusDate` AND
          OLD.`idUserInviting` <=> NEW.`idUserInviting`) THEN
    SET NEW.iVersion = @curVersion;
    UPDATE `history_groups_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;
    INSERT INTO `history_groups_groups` (
      `ID`,`iVersion`,`idGroupParent`,`idGroupChild`,`iChildOrder`,`sType`,`sRole`,`sStatusDate`,`idUserInviting`
    ) VALUES (
      NEW.`ID`,@curVersion,NEW.`idGroupParent`,NEW.`idGroupChild`,NEW.`iChildOrder`,NEW.`sType`,NEW.`sRole`,
      NEW.`sStatusDate`,NEW.`idUserInviting`
    );
  END IF;
  IF (OLD.idGroupChild != NEW.idGroupChild OR OLD.idGroupParent != NEW.idGroupParent OR OLD.sType != NEW.sType) THEN
    INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idGroupChild, 'todo')
      ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo';
    INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) (
      SELECT `groups_ancestors`.`idGroupChild`, 'todo'
        FROM `groups_ancestors`
        WHERE `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupChild`
    ) ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo';
    DELETE `groups_ancestors` FROM `groups_ancestors`
      WHERE `groups_ancestors`.`idGroupChild` = OLD.`idGroupChild` AND
            `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupParent`;
    DELETE `bridges` FROM `groups_ancestors` `child_descendants`
      JOIN `groups_ancestors` `parent_ancestors`
      JOIN `groups_ancestors` `bridges`
        ON (`bridges`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor` AND
            `bridges`.`idGroupChild` = `child_descendants`.`idGroupChild`)
      WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent` AND
            `child_descendants`.`idGroupAncestor` = OLD.`idGroupChild`;
    DELETE `child_ancestors` FROM `groups_ancestors` `child_ancestors`
      JOIN `groups_ancestors` `parent_ancestors`
        ON (`child_ancestors`.`idGroupChild` = OLD.`idGroupChild` AND
            `child_ancestors`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor`)
      WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent`;
    DELETE `parent_ancestors` FROM `groups_ancestors` `parent_ancestors`
      JOIN  `groups_ancestors` `child_ancestors`
        ON (`parent_ancestors`.`idGroupAncestor` = OLD.`idGroupParent` AND
            `child_ancestors`.`idGroupChild` = `parent_ancestors`.`idGroupChild`)
      WHERE `child_ancestors`.`idGroupAncestor` = OLD.`idGroupChild`;
  END IF;
  IF (OLD.idGroupChild != NEW.idGroupChild OR OLD.idGroupParent != NEW.idGroupParent OR OLD.sType != NEW.sType) THEN
    INSERT IGNORE INTO `groups_propagate` (ID, sAncestorsComputationState) VALUES (NEW.idGroupChild, 'todo')
      ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo';
  END IF;
END;
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_groups`
BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
  SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion;
  UPDATE `history_groups_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;
  INSERT INTO `history_groups_groups` (
    `ID`,`iVersion`,`idGroupParent`,`idGroupChild`,`iChildOrder`,`sType`,`sRole`,`sStatusDate`,`idUserInviting`,`bDeleted`
  ) VALUES (
    OLD.`ID`,@curVersion,OLD.`idGroupParent`,OLD.`idGroupChild`,OLD.`iChildOrder`,OLD.`sType`,OLD.`sRole`,
    OLD.`sStatusDate`,OLD.`idUserInviting`, 1
  );
  INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idGroupChild, 'todo')
    ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo';
  INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) (
    SELECT `groups_ancestors`.`idGroupChild`, 'todo'
      FROM `groups_ancestors`
      WHERE `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupChild`
  ) ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo';
  DELETE `groups_ancestors` FROM `groups_ancestors`
    WHERE `groups_ancestors`.`idGroupChild` = OLD.`idGroupChild` AND
          `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupParent`;
  DELETE `bridges`
    FROM `groups_ancestors` `child_descendants`
    JOIN `groups_ancestors` `parent_ancestors`
    JOIN `groups_ancestors` `bridges`
      ON (`bridges`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor` AND
          `bridges`.`idGroupChild` = `child_descendants`.`idGroupChild`)
    WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent` AND
          `child_descendants`.`idGroupAncestor` = OLD.`idGroupChild`;
  DELETE `child_ancestors`
    FROM `groups_ancestors` `child_ancestors`
    JOIN  `groups_ancestors` `parent_ancestors`
      ON (`child_ancestors`.`idGroupChild` = OLD.`idGroupChild` AND
          `child_ancestors`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor`)
    WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent`;
  DELETE `parent_ancestors`
    FROM `groups_ancestors` `parent_ancestors`
    JOIN  `groups_ancestors` `child_ancestors`
      ON (`parent_ancestors`.`idGroupAncestor` = OLD.`idGroupParent` AND
          `child_ancestors`.`idGroupChild` = `parent_ancestors`.`idGroupChild`)
    WHERE `child_ancestors`.`idGroupAncestor` = OLD.`idGroupChild`;
END
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER `before_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN IF NEW.iVersion <> OLD.iVersion THEN SET @curVersion = NEW.iVersion; ELSE SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; END IF; IF NOT (OLD.`ID` = NEW.`ID` AND OLD.`idGroupParent` <=> NEW.`idGroupParent` AND OLD.`idGroupChild` <=> NEW.`idGroupChild` AND OLD.`iChildOrder` <=> NEW.`iChildOrder` AND OLD.`sType` <=> NEW.`sType` AND OLD.`sRole` <=> NEW.`sRole` AND OLD.`sStatusDate` <=> NEW.`sStatusDate` AND OLD.`idUserInviting` <=> NEW.`idUserInviting`) THEN   SET NEW.iVersion = @curVersion;   UPDATE `history_groups_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL;   INSERT INTO `history_groups_groups` (`ID`,`iVersion`,`idGroupParent`,`idGroupChild`,`iChildOrder`,`sType`,`sRole`,`sStatusDate`,`idUserInviting`)       VALUES (NEW.`ID`,@curVersion,NEW.`idGroupParent`,NEW.`idGroupChild`,NEW.`iChildOrder`,NEW.`sType`,NEW.`sRole`,NEW.`sStatusDate`,NEW.`idUserInviting`) ; END IF; IF (OLD.idGroupChild != NEW.idGroupChild OR OLD.idGroupParent != NEW.idGroupParent OR OLD.sType != NEW.sType) THEN INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idGroupChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idGroupParent, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) (SELECT `groups_ancestors`.`idGroupChild`, 'todo' FROM `groups_ancestors` WHERE `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupChild`) ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; DELETE `groups_ancestors` from `groups_ancestors` WHERE `groups_ancestors`.`idGroupChild` = OLD.`idGroupChild` and `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupParent`;DELETE `bridges` FROM `groups_ancestors` `child_descendants` JOIN `groups_ancestors` `parent_ancestors` JOIN `groups_ancestors` `bridges` ON (`bridges`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor` AND `bridges`.`idGroupChild` = `child_descendants`.`idGroupChild`) WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent` AND `child_descendants`.`idGroupAncestor` = OLD.`idGroupChild`; DELETE `child_ancestors` FROM `groups_ancestors` `child_ancestors` JOIN  `groups_ancestors` `parent_ancestors` ON (`child_ancestors`.`idGroupChild` = OLD.`idGroupChild` AND `child_ancestors`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor`) WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent`; DELETE `parent_ancestors` FROM `groups_ancestors` `parent_ancestors` JOIN  `groups_ancestors` `child_ancestors` ON (`parent_ancestors`.`idGroupAncestor` = OLD.`idGroupParent` AND `child_ancestors`.`idGroupChild` = `parent_ancestors`.`idGroupChild`) WHERE `child_ancestors`.`idGroupAncestor` = OLD.`idGroupChild`  ; END IF; IF (OLD.idGroupChild != NEW.idGroupChild OR OLD.idGroupParent != NEW.idGroupParent OR OLD.sType != NEW.sType) THEN INSERT IGNORE INTO `groups_propagate` (ID, sAncestorsComputationState) VALUES (NEW.idGroupChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'  ; END IF; END;
-- +migrate StatementEnd
DROP TRIGGER `before_delete_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN SELECT ROUND(UNIX_TIMESTAMP(CURTIME(2)) * 10) INTO @curVersion; UPDATE `history_groups_groups` SET `iNextVersion` = @curVersion WHERE `ID` = OLD.`ID` AND `iNextVersion` IS NULL; INSERT INTO `history_groups_groups` (`ID`,`iVersion`,`idGroupParent`,`idGroupChild`,`iChildOrder`,`sType`,`sRole`,`sStatusDate`,`idUserInviting`, `bDeleted`) VALUES (OLD.`ID`,@curVersion,OLD.`idGroupParent`,OLD.`idGroupChild`,OLD.`iChildOrder`,OLD.`sType`,OLD.`sRole`,OLD.`sStatusDate`,OLD.`idUserInviting`, 1); INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idGroupChild, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) VALUES (OLD.idGroupParent, 'todo') ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; INSERT IGNORE INTO `groups_propagate` (`ID`, `sAncestorsComputationState`) (SELECT `groups_ancestors`.`idGroupChild`, 'todo' FROM `groups_ancestors` WHERE `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupChild`) ON DUPLICATE KEY UPDATE `sAncestorsComputationState` = 'todo'; DELETE `groups_ancestors` from `groups_ancestors` WHERE `groups_ancestors`.`idGroupChild` = OLD.`idGroupChild` and `groups_ancestors`.`idGroupAncestor` = OLD.`idGroupParent`;DELETE `bridges` FROM `groups_ancestors` `child_descendants` JOIN `groups_ancestors` `parent_ancestors` JOIN `groups_ancestors` `bridges` ON (`bridges`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor` AND `bridges`.`idGroupChild` = `child_descendants`.`idGroupChild`) WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent` AND `child_descendants`.`idGroupAncestor` = OLD.`idGroupChild`; DELETE `child_ancestors` FROM `groups_ancestors` `child_ancestors` JOIN  `groups_ancestors` `parent_ancestors` ON (`child_ancestors`.`idGroupChild` = OLD.`idGroupChild` AND `child_ancestors`.`idGroupAncestor` = `parent_ancestors`.`idGroupAncestor`) WHERE `parent_ancestors`.`idGroupChild` = OLD.`idGroupParent`; DELETE `parent_ancestors` FROM `groups_ancestors` `parent_ancestors` JOIN  `groups_ancestors` `child_ancestors` ON (`parent_ancestors`.`idGroupAncestor` = OLD.`idGroupParent` AND `child_ancestors`.`idGroupChild` = `parent_ancestors`.`idGroupChild`) WHERE `child_ancestors`.`idGroupAncestor` = OLD.`idGroupChild` ; END
-- +migrate StatementEnd

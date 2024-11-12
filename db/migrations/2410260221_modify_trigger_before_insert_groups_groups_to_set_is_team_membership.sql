-- +migrate Up
DROP TRIGGER IF EXISTS `before_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_groups` BEFORE INSERT ON `groups_groups` FOR EACH ROW BEGIN
  SET NEW.is_team_membership = (SELECT type = 'Team' FROM `groups` WHERE id = NEW.parent_group_id FOR SHARE);
  IF NOT NEW.is_team_membership THEN
    INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
  END IF;
END
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER IF EXISTS `before_insert_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_groups` BEFORE INSERT ON `groups_groups` FOR EACH ROW BEGIN INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo') ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo' ; END
-- +migrate StatementEnd

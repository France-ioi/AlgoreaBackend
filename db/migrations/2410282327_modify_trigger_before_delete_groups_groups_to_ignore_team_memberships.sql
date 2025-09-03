-- +migrate Up
DROP TRIGGER `before_delete_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
  IF NOT OLD.`is_team_membership` THEN
    INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
    ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
  END IF;
END
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER `before_delete_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_delete_groups_groups` BEFORE DELETE ON `groups_groups` FOR EACH ROW BEGIN
    INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
    ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
END
-- +migrate StatementEnd

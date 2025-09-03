-- +migrate Up
DROP TRIGGER IF EXISTS `after_insert_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN
  INSERT INTO `groups_ancestors` (`ancestor_group_id`, `child_group_id`) VALUES (NEW.`id`, NEW.`id`);
END;
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER IF EXISTS `after_insert_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.`id`, 'todo') ; END
-- +migrate StatementEnd

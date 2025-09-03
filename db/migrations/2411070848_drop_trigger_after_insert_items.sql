-- +migrate Up
DROP TRIGGER `after_insert_items`;

-- +migrate Down
DROP TRIGGER IF EXISTS `after_insert_items`;
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_items` AFTER INSERT ON `items` FOR EACH ROW BEGIN INSERT IGNORE INTO `items_propagate` (`id`, `ancestors_computation_state`) VALUES (NEW.`id`, 'todo') ; END
-- +migrate StatementEnd

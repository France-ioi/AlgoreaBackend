-- +goose Up
DROP TRIGGER `after_insert_groups`;
-- +goose StatementBegin
CREATE TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN
  INSERT INTO `groups_ancestors` (`ancestor_group_id`, `child_group_id`, `child_group_type`) VALUES (NEW.`id`, NEW.`id`, NEW.`type`);
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER `after_insert_groups`;
-- +goose StatementBegin
CREATE TRIGGER `after_insert_groups` AFTER INSERT ON `groups` FOR EACH ROW BEGIN
  INSERT INTO `groups_ancestors` (`ancestor_group_id`, `child_group_id`) VALUES (NEW.`id`, NEW.`id`);
END;
-- +goose StatementEnd

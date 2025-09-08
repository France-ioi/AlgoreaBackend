-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER `after_update_groups` AFTER UPDATE ON `groups` FOR EACH ROW BEGIN
  IF NEW.type <> OLD.type THEN
    UPDATE `groups_groups` SET child_group_type = NEW.type WHERE child_group_id = NEW.id;
    UPDATE `groups_ancestors` SET child_group_type = NEW.type WHERE child_group_id = NEW.id;
  END IF;
END
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER `after_update_groups`;

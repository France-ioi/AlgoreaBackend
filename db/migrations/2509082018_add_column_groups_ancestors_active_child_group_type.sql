-- +goose Up
CREATE OR REPLACE VIEW groups_ancestors_active AS SELECT * FROM groups_ancestors WHERE NOW() < expires_at;

-- +goose Down
CREATE OR REPLACE VIEW `groups_ancestors_active` AS
  SELECT
    `groups_ancestors`.`ancestor_group_id` AS `ancestor_group_id`,
    `groups_ancestors`.`child_group_id` AS `child_group_id`,
    `groups_ancestors`.`is_self` AS `is_self`,
    `groups_ancestors`.`expires_at` AS `expires_at`
  FROM `groups_ancestors`
  WHERE (NOW() < `groups_ancestors`.`expires_at`);

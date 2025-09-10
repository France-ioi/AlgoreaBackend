-- +goose Up
CREATE OR REPLACE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;

-- +goose Down
CREATE OR REPLACE VIEW `groups_groups_active` AS
  SELECT `groups_groups`.`parent_group_id` AS `parent_group_id`,
         `groups_groups`.`child_group_id` AS `child_group_id`,
         `groups_groups`.`expires_at` AS `expires_at`,
         `groups_groups`.`is_team_membership` AS `is_team_membership`,
         `groups_groups`.`personal_info_view_approved_at` AS `personal_info_view_approved_at`,
         `groups_groups`.`personal_info_view_approved` AS `personal_info_view_approved`,
         `groups_groups`.`lock_membership_approved_at` AS `lock_membership_approved_at`,
         `groups_groups`.`lock_membership_approved` AS `lock_membership_approved`,
         `groups_groups`.`watch_approved_at` AS `watch_approved_at`,
         `groups_groups`.`watch_approved` AS `watch_approved`
  FROM `groups_groups` WHERE (NOW() < `groups_groups`.`expires_at`);

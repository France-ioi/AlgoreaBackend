-- +migrate Up
UPDATE `groups_groups` SET `is_team_membership` = 1 WHERE `parent_group_id` IN (SELECT `id` FROM `groups` WHERE `type` = 'Team');

-- +migrate Down

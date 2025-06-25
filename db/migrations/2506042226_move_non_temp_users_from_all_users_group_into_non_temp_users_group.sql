-- +migrate Up
SELECT `id` INTO @all_users_group FROM `groups` WHERE `type`='Base' AND `text_id`='AllUsers';
SELECT `id` INTO @non_temp_users_group FROM `groups` WHERE `type`='Base' AND `text_id`='NonTempUsers';

INSERT IGNORE INTO `groups_groups` (parent_group_id, child_group_id)
  SELECT @non_temp_users_group, users.group_id
  FROM `users` WHERE NOT temp_user;

DELETE `groups_groups`
FROM `groups_groups`
JOIN `users` ON `users`.`group_id`=`groups_groups`.`child_group_id` AND NOT `users`.`temp_user`
WHERE `groups_groups`.`parent_group_id`=@all_users_group;

-- +migrate Down
SELECT `id` INTO @all_users_group FROM `groups` WHERE `type`='Base' AND `text_id`='AllUsers';
SELECT `id` INTO @non_temp_users_group FROM `groups` WHERE `type`='Base' AND `text_id`='NonTempUsers';

INSERT IGNORE INTO `groups_groups` (parent_group_id, child_group_id)
SELECT @all_users_group, users.group_id
FROM `users` WHERE NOT temp_user;

DELETE `groups_groups`
FROM `groups_groups`
JOIN `users` ON `users`.`group_id`=`groups_groups`.`child_group_id` AND NOT `users`.`temp_user`
WHERE `groups_groups`.`parent_group_id`=@non_temp_users_group;

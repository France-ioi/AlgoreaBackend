-- +migrate Up
INSERT INTO `groups_groups` (parent_group_id, child_group_id)
  SELECT id, 4 FROM `groups` WHERE `type`='Base' AND `text_id`='AllUsers';

-- +migrate Down
DELETE FROM `groups_groups`
WHERE child_group_id=4 AND parent_group_id =
    (SELECT id FROM `groups` WHERE `type`='Base' AND `text_id`='AllUsers' LIMIT 1);

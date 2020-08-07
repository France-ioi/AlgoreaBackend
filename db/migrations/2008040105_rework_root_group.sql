-- +migrate Up
SELECT `id` INTO @root_id FROM `groups` WHERE `type`='Base' AND `text_id`='Root';
SELECT `id` INTO @root_self_id FROM `groups` WHERE `type`='Base' AND `text_id`='RootSelf';

INSERT INTO `permissions_granted` (`group_id`, `item_id`, `source_group_id`, `origin`, `can_view`, `latest_update_at`)
  SELECT @root_self_id, `item_id`, @root_self_id, 'group_membership', `can_view_value`, `latest_update_at`
  FROM `permissions_granted` AS `root`
  WHERE `root`.`group_id` = @root_id
ON DUPLICATE KEY UPDATE
  `can_view` = GREATEST(`permissions_granted`.`can_view_value`, VALUES(`can_view_value`)),
  `latest_update_at` = GREATEST(`permissions_granted`.`latest_update_at`, VALUES(latest_update_at));

DELETE FROM `groups` WHERE `type`='Base' AND `text_id`='Root';
UPDATE `groups` SET `name`='AllUsers', `text_id`='AllUsers', `description`='AllUsers' WHERE `type`='Base' AND `text_id`='RootSelf';
UPDATE `groups` SET `name`='TempUsers', `text_id`='TempUsers' WHERE `type`='Base' AND `text_id`='RootTemp';

-- +migrate Down
SET @id = FLOOR(RAND(11) * 1000000000) + FLOOR(RAND(12) * 1000000000) * 1000000000;
INSERT INTO `groups` (`id`, `name`, `type`, `text_id`, `description`) VALUES (@id, 'Root', 'Base', 'Root', 'Root');
INSERT INTO `groups_groups` (parent_group_id, child_group_id)
    SELECT @id, id FROM `groups` WHERE `type`='Base' AND `text_id`='RootSelf';

UPDATE `groups` SET `name`='RootSelf', `text_id`='RootSelf', `description`='RootSelf' WHERE `type`='Base' AND `text_id`='AllUsers';
UPDATE `groups` SET `name`='RootTemp', `text_id`='RootTemp' WHERE `type`='Base' AND `text_id`='TempUsers';

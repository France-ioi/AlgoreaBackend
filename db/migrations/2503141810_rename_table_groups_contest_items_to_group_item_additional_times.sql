-- +migrate Up
ALTER TABLE groups_contest_items
  MODIFY `additional_time` time NOT NULL DEFAULT '00:00:00' COMMENT 'Time that was attributed (can be negative) to this group for this time-limited item',
  DROP FOREIGN KEY `fk_groups_contest_items_group_id_groups_id`,
  DROP FOREIGN KEY `fk_groups_contest_items_item_id_items_id`,
  DROP INDEX `fk_groups_contest_items_item_id_items_id`,
  ADD CONSTRAINT `fk_group_item_additional_times_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `fk_group_item_additional_times_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE,
  COMMENT 'Additional times of groups on time-limited items',
  RENAME TO group_item_additional_times;

-- +migrate Down
ALTER TABLE group_item_additional_times
  MODIFY `additional_time` time NOT NULL DEFAULT '00:00:00' COMMENT 'Time that was attributed (can be negative) to this group for this contest',
  DROP FOREIGN KEY `fk_group_item_additional_times_group_id_groups_id`,
  DROP FOREIGN KEY `fk_group_item_additional_times_item_id_items_id`,
  DROP INDEX `fk_group_item_additional_times_item_id_items_id`,
  ADD CONSTRAINT `fk_groups_contest_items_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `fk_groups_contest_items_item_id_items_id` FOREIGN KEY (`item_id`) REFERENCES `items` (`id`) ON DELETE CASCADE,
  COMMENT 'Group constraints on contest participations',
  RENAME TO groups_contest_items;

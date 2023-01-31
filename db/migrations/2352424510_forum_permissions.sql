-- +migrate Up
ALTER TABLE `permissions_granted`
  ADD COLUMN `can_request_help_to`
    BIGINT NULL DEFAULT NULL
    COMMENT 'Whether the group can create a forum thread accessible to the pointed group. NULL = no rights to create.'
    AFTER `can_edit_value`;

ALTER TABLE `permissions_granted`
  ADD CONSTRAINT `fk_can_request_help_to_groups_id` FOREIGN KEY (`can_request_help_to`) REFERENCES `groups`(`id`)
    ON DELETE SET NULL;

ALTER TABLE `items_items`
  ADD COLUMN `request_help_propagation`
    TINYINT NOT NULL DEFAULT 0
    COMMENT 'Whether can_request_help_to propagates'
    AFTER `edit_propagation`;

-- +migrate Down
ALTER TABLE `permissions_granted`
  DROP FOREIGN KEY `fk_can_request_help_to_groups_id`;

ALTER TABLE `permissions_granted`
  DROP `can_request_help_to`;

ALTER TABLE `items_items`
  DROP `request_help_propagation`;

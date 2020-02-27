
-- +migrate Up
ALTER TABLE `groups_items_propagate` COMMENT 'Used by the access rights propagation algorithm to keep track of the status of the propagation.';
ALTER TABLE `groups_propagate` COMMENT 'Used by the algorithm that updates the groups_ancestors table, and keeps track of what groups still need to have their relationship with their descendants / ancestors propagated.';
ALTER TABLE `items_propagate` COMMENT 'Used by the algorithm that updates the items_ancestors table, and keeps track of what items still need to have their relationship with their descendants / ancestors propagated.';

-- +migrate Down
ALTER TABLE `groups_items_propagate` COMMENT '';
ALTER TABLE `groups_propagate` COMMENT '';
ALTER TABLE `items_propagate` COMMENT '';

-- +migrate Up
-- +migrate StatementBegin
CREATE TRIGGER `after_insert_permissions_generated` AFTER INSERT ON `permissions_generated` FOR EACH ROW BEGIN
    IF NEW.`can_view_generated` != 'none' THEN
        UPDATE `attempts`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `attempts`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `attempts`.`group_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`
        SET `result_propagation_state` = 'changed';
    END IF;
END
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE TRIGGER `after_update_permissions_generated` AFTER UPDATE ON `permissions_generated` FOR EACH ROW BEGIN
    IF (OLD.`can_view_generated` = 'none' OR OLD.`item_id` != NEW.`item_id` OR OLD.`group_id` != NEW.`group_id`) AND
       NEW.can_view_generated != 'none' THEN
        UPDATE `attempts`
            JOIN `items_ancestors` ON `items_ancestors`.`child_item_id` = `attempts`.`item_id` AND
                                      `items_ancestors`.`ancestor_item_id` = NEW.`item_id`
            JOIN `groups_ancestors_active` ON `groups_ancestors_active`.`child_group_id` = `attempts`.`group_id` AND
                                              `groups_ancestors_active`.`ancestor_group_id` = NEW.`group_id`
        SET `result_propagation_state` = 'changed';
    END IF;
END
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER `after_insert_permissions_generated`;
DROP TRIGGER `after_update_permissions_generated`;

-- +migrate Up
DROP TRIGGER `before_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF OLD.`parent_group_id` != NEW.`parent_group_id` OR OLD.`child_group_id` != NEW.`child_group_id` OR OLD.`is_team_membership` != NEW.`is_team_membership` THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Unable to change immutable columns of groups_groups (parent_group_id/child_group_id/is_team_membership)';
    END IF;
END
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER `before_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF OLD.`parent_group_id` != NEW.`parent_group_id` OR OLD.`child_group_id` != NEW.`child_group_id` THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Unable to change immutable groups_groups.parent_group_id and/or groups_groups.child_group_id';
    END IF;
END
-- +migrate StatementEnd

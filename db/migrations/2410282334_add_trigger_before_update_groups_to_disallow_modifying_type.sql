-- +migrate Up
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups` BEFORE UPDATE ON `groups` FOR EACH ROW BEGIN
  IF OLD.`type` != NEW.`type` AND (OLD.`type` IN ('User', 'Team') OR NEW.`type` IN ('User', 'Team')) THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Unable to change groups.type from/to User/Team';
  END IF;
END
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER `before_update_groups`;

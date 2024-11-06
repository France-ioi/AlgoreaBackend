-- +migrate Up
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups` BEFORE UPDATE ON `groups` FOR EACH ROW BEGIN
  IF OLD.`type` != NEW.`type` THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Unable to change immutable groups.type';
  END IF;
END
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER `before_update_groups`;

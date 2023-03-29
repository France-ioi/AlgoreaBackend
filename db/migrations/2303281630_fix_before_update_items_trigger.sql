-- +migrate Up

DROP TRIGGER `before_update_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items` BEFORE UPDATE ON `items` FOR EACH ROW
BEGIN
  IF NOT OLD.url <=> NEW.url THEN
    IF NEW.url IS NOT NULL THEN
      SET @platformID = (SELECT platforms.id FROM platforms WHERE NEW.url REGEXP platforms.regexp ORDER BY platforms.priority DESC LIMIT 1);

      SET NEW.platform_id=@platformID;
    ELSE
      SET NEW.platform_id = NULL;
    END IF;
  END IF;
END;
-- +migrate StatementEnd


-- +migrate Down

DROP TRIGGER `before_update_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_items` BEFORE UPDATE ON `items` FOR EACH ROW
BEGIN
  IF NOT OLD.url <=> NEW.url THEN
    IF NEW.url IS NOT NULL THEN
      SELECT platforms.id INTO @platformID FROM platforms
      WHERE NEW.url REGEXP platforms.regexp
      ORDER BY platforms.priority DESC LIMIT 1;

      SET NEW.platform_id=@platformID;
    ELSE
      SET NEW.platform_id = NULL;
    END IF;
  END IF;
END;
-- +migrate StatementEnd

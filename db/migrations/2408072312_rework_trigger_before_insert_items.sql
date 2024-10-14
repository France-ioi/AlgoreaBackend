-- +migrate Up
DROP TRIGGER IF EXISTS `before_insert_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items` BEFORE INSERT ON `items` FOR EACH ROW BEGIN
  IF (NEW.id IS NULL OR NEW.id = 0) THEN
    SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000;
  END IF;

  IF NEW.url IS NOT NULL THEN
    SET NEW.platform_id = (SELECT platforms.id FROM platforms WHERE NEW.url REGEXP platforms.regexp ORDER BY platforms.priority DESC LIMIT 1);
  ELSE
    SET NEW.platform_id = NULL;
  END IF;
END
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER IF EXISTS `before_insert_items`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_items` BEFORE INSERT ON `items` FOR EACH ROW BEGIN
  IF (NEW.id IS NULL OR NEW.id = 0) THEN
    SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000;
  END IF;

  IF NEW.url IS NOT NULL THEN
    SELECT platforms.id INTO @platformID FROM platforms
    WHERE NEW.url REGEXP platforms.regexp
    ORDER BY platforms.priority DESC LIMIT 1;

    SET NEW.platform_id=@platformID;
  ELSE
    SET NEW.platform_id = NULL;
  END IF;
END
-- +migrate StatementEnd

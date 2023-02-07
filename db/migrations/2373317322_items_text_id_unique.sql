-- +migrate Up

# Test data:
# SET FOREIGN_KEY_CHECKS=0;
# INSERT INTO `items` (`text_id`, `default_language_tag`) VALUES
# ('test', 'fr'),
# (NULL, 'fr'),
# ('test', 'fr'),
# (NULL, 'fr'),
# ('test', 'fr'),
# ('test', 'fr'),
# ('test3', 'fr'),
# ('plop', 'fr');

# We first create a field text_id_unique with the same characteristics as text_id, with a UNIQUE constraint
# Then, we update text_id_unique for each row, and use a TRIGGER to add a number if the value already exists
# Finally, we rename text_id_unique into text_id_unique and REMOVE the TRIGGER

LOCK TABLES `items` WRITE;

UPDATE `items` SET `text_id` = NULL WHERE `text_id` = '';

ALTER TABLE `items`
  ADD COLUMN `text_id_unique` VARCHAR(200) DEFAULT NULL COMMENT 'Unique string identifying the item, independently of where it is hosted'
  AFTER `text_id`;
ALTER TABLE `items`
  ADD UNIQUE `unique_text_id_unique`(`text_id_unique`);

DROP TRIGGER IF EXISTS itemTextIdUniqueUpdate;
DELIMITER |
CREATE TRIGGER itemTextIdUniqueUpdate BEFORE UPDATE ON `items`
  FOR EACH ROW BEGIN
    SET @counter = 1;
    SET NEW.`text_id_unique` = OLD.`text_id`;
    WHILE exists (SELECT 1 FROM `items` WHERE `items`.`text_id_unique` = NEW.`text_id_unique`) DO
      SET NEW.`text_id_unique` = CONCAT(OLD.`text_id`, @counter);
      SET @counter = @counter + 1;
    END WHILE;
  END;
|
DELIMITER ;

UPDATE `items` SET `items`.`text_id_unique` = `items`.`text_id`;

DROP TRIGGER itemTextIdUniqueUpdate;

ALTER TABLE `items`
  DROP COLUMN `text_id`;
ALTER TABLE `items`
  RENAME COLUMN `text_id_unique` TO `text_id`;

UNLOCK TABLES;

-- +migrate Down


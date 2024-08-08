-- +migrate Up
UPDATE `items`
SET `platform_id` = IF(
  `items`.`url` IS NULL,
  NULL,
  (SELECT `platforms`.`id`
   FROM `platforms`
   WHERE `items`.`url` REGEXP `platforms`.`regexp`
   ORDER BY `platforms`.`priority` DESC
   LIMIT 1)
);

-- +migrate Down


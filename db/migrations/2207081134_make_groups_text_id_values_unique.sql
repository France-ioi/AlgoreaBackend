-- +migrate Up
UPDATE `groups` JOIN (
  SELECT `text_id`,
         (SELECT `id`
          FROM `groups` AS `g`
          WHERE `g`.`text_id` = `groups`.`text_id`
          ORDER BY `created_at` DESC
          LIMIT 1) AS `keep_id`
  FROM `groups`
  GROUP BY `text_id`
  HAVING COUNT(*) > 1
     AND `text_id` != ''
) AS `duplicated` USING (`text_id`)
SET `text_id` = CONCAT(`text_id`, '_', `id`)
WHERE `id` != `keep_id`;

ALTER TABLE `groups`
  ADD UNIQUE INDEX `text_id` (`text_id`);

-- +migrate Down
ALTER TABLE `groups` DROP INDEX `text_id`;

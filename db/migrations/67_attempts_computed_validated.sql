-- +migrate Up

-- fix wrong computation of validated_at for chapters where it was not supposed to be set (cannot be reverted in the down-migration)
UPDATE `groups_attempts` SET `validated_at` = NULL WHERE NOT `validated` = 0;

ALTER TABLE `groups_attempts`
  -- change comment
  MODIFY COLUMN `validated_at` datetime DEFAULT NULL COMMENT 'When the item was validated within this attempt (validation criteria depends on item)',

  -- make it generated (from validated_at) and change comment
  -- there are 394 attempts left which are validated with validated_at=NULL, these one will lose their validation
  DROP COLUMN `validated`,
  ADD COLUMN `validated` tinyint(1) AS (`validated_at` IS NOT NULL) NOT NULL COMMENT 'See `validated_at`' AFTER `children_validated`;


-- +migrate Down

ALTER TABLE `groups_attempts`
  MODIFY COLUMN `validated_at` datetime DEFAULT NULL COMMENT 'When the item was validated, within this attempt.',
  DROP COLUMN `validated`,
  ADD COLUMN `validated` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether this item is validated, within this attempt (different items have different criteria for validation)' AFTER `children_validated`;

UPDATE `groups_attempts` SET `validated` = 1 WHERE `validated_at` IS NOT NULL;

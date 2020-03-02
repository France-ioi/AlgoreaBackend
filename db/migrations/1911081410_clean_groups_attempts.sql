-- +migrate Up

ALTER TABLE `groups_attempts`
  DROP COLUMN `precision`,
  DROP COLUMN `autonomy`,
  DROP COLUMN `ranked`,
  DROP COLUMN `corrections_read`,
  DROP COLUMN `thread_started_at`,
  DROP COLUMN `all_lang_prog`,

  RENAME COLUMN `submissions_attempts` TO `submissions`,

  -- rename and change comment
  CHANGE COLUMN `key_obtained` `has_unlocked_items` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the score of this attempt allows unlocking other items (score >= items.score_min_unlock)';


-- +migrate Down

ALTER TABLE `groups_attempts`
  ADD COLUMN `ranked` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether this attempt is official for this item (as opposed to an extra attempt after an exam has ended, for example)' AFTER `latest_hint_at`,
  ADD COLUMN `corrections_read` int(11) NOT NULL DEFAULT '0' COMMENT 'Number of solutions that the group read among the descendants of this item, for this attempt.' AFTER `hints_cached`,
  ADD COLUMN `precision` int(11) NOT NULL DEFAULT '0' COMMENT 'Precision (based on a formula to be defined) of the user, when working on this item and its descendants.' AFTER `corrections_read`,
  ADD COLUMN `autonomy` int(11) NOT NULL DEFAULT '0' COMMENT 'Autonomy (based on a formula to be defined) of the user, when working on this item and its descendants (how much help / hints he used)' AFTER `precision`,
  ADD COLUMN `thread_started_at` datetime DEFAULT NULL COMMENT 'When the discussion thread was started by this group on the forum' AFTER `latest_activity_at`,
  ADD COLUMN `all_lang_prog` varchar(200) DEFAULT NULL COMMENT 'List of programming languages used' AFTER `latest_hint_at`,

  RENAME COLUMN `submissions` TO `submissions_attempts`,

  CHANGE COLUMN `has_unlocked_items` `key_obtained` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the user obtained the key on this item (changed to 1 if the user gets a score >= items.score_min_unlock, will grant access to new items from items.unlocked_item_ids). This information is propagated to users_items.';

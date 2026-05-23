-- +goose Up
ALTER TABLE `items`
  ADD COLUMN `display_settings` JSON NOT NULL DEFAULT (JSON_OBJECT())
    COMMENT 'JSON object containing display/UI-only settings consumed by the frontend (e.g. children_layout, prompt_to_join_group_by_code). The backend does not interpret these settings; it stores and returns them as-is.'
    AFTER `options`;

-- Backfill `display_settings` from the legacy columns. We follow the
-- "omit defaults" convention: a key is only present when its value differs from
-- the column default (`children_layout = 'List'`, `prompt_to_join_group_by_code = 0`).
-- `IFNULL(children_layout, 'List')` treats a NULL legacy value as the default.
--
-- `prompt_to_join_group_by_code` is a TINYINT(1), and CAST(<tinyint> AS JSON)
-- yields a JSON *number* (e.g. 1) — but we want the JSON *boolean* `true`, since
-- that's what new clients write through the API and what the response builder
-- type-asserts on. CAST(IF(col, 'true', 'false') AS JSON) is the documented way
-- to produce a JSON boolean from a numeric source in MySQL 8.
UPDATE `items`
SET `display_settings` = JSON_REMOVE(
  JSON_OBJECT(
    'children_layout',              IFNULL(`children_layout`, 'List'),
    'prompt_to_join_group_by_code', CAST(IF(`prompt_to_join_group_by_code`, 'true', 'false') AS JSON)
  ),
  -- `'$.__none__'` is a no-op path: JSON_REMOVE silently skips nonexistent paths.
  CASE WHEN IFNULL(`children_layout`, 'List') = 'List'
       THEN '$.children_layout' ELSE '$.__none__' END,
  CASE WHEN `prompt_to_join_group_by_code` = 0
       THEN '$.prompt_to_join_group_by_code' ELSE '$.__none__' END
);

ALTER TABLE `items`
  DROP COLUMN `repository_path`,
  DROP COLUMN `title_bar_visible`,
  DROP COLUMN `display_details_in_parent`,
  DROP COLUMN `full_screen`,
  DROP COLUMN `fixed_ranks`,
  DROP COLUMN `show_user_infos`,
  DROP COLUMN `children_layout`,
  DROP COLUMN `prompt_to_join_group_by_code`;

-- +goose Down
ALTER TABLE `items`
  ADD COLUMN `repository_path` text,
  ADD COLUMN `title_bar_visible` tinyint unsigned NOT NULL DEFAULT '1'
    COMMENT 'Whether the title bar should be visible initially when this item is loaded',
  ADD COLUMN `display_details_in_parent` tinyint unsigned NOT NULL DEFAULT '0'
    COMMENT 'If true, display a large icon, the subtitle, and more within the parent chapter',
  ADD COLUMN `full_screen` enum('forceYes','forceNo','default') NOT NULL DEFAULT 'default'
    COMMENT 'Whether the item should be loaded in full screen mode (without the navigation panel and most of the top header). By default, tasks are displayed in full screen, but not chapters.',
  ADD COLUMN `fixed_ranks` tinyint(1) NOT NULL DEFAULT '0'
    COMMENT 'If true, prevents users from changing the order of the children by drag&drop and auto-calculation of the order of children. Allows for manual setting of the order, for instance in cases where we want to have multiple items with the same order (check items_items.child_order).',
  ADD COLUMN `show_user_infos` tinyint(1) NOT NULL DEFAULT '0'
    COMMENT 'Always show user infos in title bar of all descendants. Allows the teacher to see who is working on what (e.g., during an exam).',
  ADD COLUMN `children_layout` enum('List','Grid','Hide') DEFAULT 'List'
    COMMENT 'How the children list are displayed (for chapters and skills)',
  ADD COLUMN `prompt_to_join_group_by_code` tinyint(1) NOT NULL DEFAULT '0'
    COMMENT 'Whether the UI should display a form for joining a group by code on the item page';

-- Re-derive the two columns that we know how to extract from `display_settings`.
-- `->>` returns the string "null" when the key is absent (the JSON_EXTRACT'd
-- value is JSON null), hence the NULLIF guard before COALESCE.
UPDATE `items`
SET
  `children_layout` = COALESCE(
    NULLIF(`display_settings`->>'$.children_layout', 'null'),
    'List'
  ),
  `prompt_to_join_group_by_code` = COALESCE(
    JSON_EXTRACT(`display_settings`, '$.prompt_to_join_group_by_code') = TRUE,
    0
  );

ALTER TABLE `items` DROP COLUMN `display_settings`;

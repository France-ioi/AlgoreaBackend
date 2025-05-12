-- +migrate Up
ALTER TABLE `attempts` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `badges`
  MODIFY `name` text,
  MODIFY `code` text NOT NULL,
  DEFAULT CHARACTER SET utf8mb4;
ALTER TABLE `error_log`
  MODIFY `url` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_as_ci NOT NULL,
  MODIFY `browser` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_as_ci NOT NULL,
  MODIFY `details` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_as_ci NOT NULL,
  DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_as_ci;
ALTER TABLE `filters` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `gorp_migrations` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `gradings` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `group_managers` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `group_membership_changes` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `group_pending_requests` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `groups` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `groups`
  MODIFY `description` text
    COMMENT 'Purpose of this group. Will be visible by its members. or by the public if the group is public.';
ALTER TABLE `groups_ancestors` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `groups_groups` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `groups_propagate` CONVERT TO CHARACTER SET utf8mb4;

ALTER TABLE `permissions_generated` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `permissions_granted` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `permissions_propagate` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `permissions_propagate_sync` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `platforms` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `platforms`
  MODIFY `regexp` text
    COMMENT 'Regexp matching the urls, to automatically detect content from this platform. It is the only way to specify which items are from which platform. Recomputation of items.platform_id is triggered when changed.';
ALTER TABLE `results` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `results`
  MODIFY `hints_requested` mediumtext COMMENT 'JSON array of the hints that have been requested for this attempt';
ALTER TABLE `results_propagate` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `results_propagate_sync` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `results_recompute_for_items` CONVERT TO CHARACTER SET utf8mb4;

ALTER TABLE `users` CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE `users`
  MODIFY `login` varchar(100) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT 'login provided by the auth platform',
  MODIFY `student_id` text COMMENT 'A student id provided by the school, provided by auth platform',
  MODIFY `address` mediumtext COMMENT 'Address, provided by auth platform',
  MODIFY `zipcode` longtext COMMENT 'Zip code, provided by auth platform',
  MODIFY `city` longtext COMMENT 'City, provided by auth platform',
  MODIFY `land_line_number` longtext COMMENT 'Phone number, provided by auth platform',
  MODIFY `cell_phone_number` longtext COMMENT 'Mobile phone number, provided by auth platform',
  MODIFY `free_text` mediumtext COMMENT 'Text provided by the user, to be displayed on his public profile';

-- +migrate Down
ALTER TABLE `users` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `users`
  MODIFY `login` varchar(100) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT 'login provided by the auth platform';

ALTER TABLE `results_recompute_for_items` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `results_propagate_sync` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `results_propagate` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `results` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `platforms` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `permissions_propagate_sync` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `permissions_propagate` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `permissions_granted` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `permissions_generated` CONVERT TO CHARACTER SET utf8mb3;

ALTER TABLE `groups_propagate` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `groups_groups` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `groups_ancestors` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `groups` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `group_pending_requests` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `group_membership_changes` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `group_managers` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `gradings` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `gorp_migrations` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `filters` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `error_log` CONVERT TO CHARACTER SET utf8mb3 COLLATE utf8_unicode_ci;
ALTER TABLE `badges` CONVERT TO CHARACTER SET utf8mb3;
ALTER TABLE `attempts` CONVERT TO CHARACTER SET utf8mb3;

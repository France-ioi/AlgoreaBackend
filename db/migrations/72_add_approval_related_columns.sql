-- +migrate Up
ALTER TABLE `groups`
    ADD COLUMN `require_personal_info_access_approval` ENUM('none', 'view', 'edit') NOT NULL DEFAULT 'none'
        COMMENT 'If not ''none'', requires (for joining) members to approve that managers may be able to view or edit their personal information'
            AFTER `lock_user_deletion_until`,
    ADD COLUMN `require_lock_membership_approval_until` DATETIME DEFAULT NULL
        COMMENT 'If not null or in the future, requires (for joining) members to approve that they will not be able to leave the group without approval until the given date'
            AFTER `require_personal_info_access_approval`,
    ADD COLUMN `require_watch_approval` TINYINT(1) NOT NULL DEFAULT 0
        COMMENT 'Whether it requires (for joining) members to approve that managers may be able to watch their results and answers'
            AFTER `require_lock_membership_approval_until`;

ALTER TABLE `groups_groups`
    ADD COLUMN `personal_info_view_approved_at` DATETIME DEFAULT NULL AFTER `expires_at`,
    ADD COLUMN `personal_info_view_approved` TINYINT(1)
        AS (`personal_info_view_approved_at` IS NOT NULL) NOT NULL
        COMMENT 'personal_info_view_approved_at as boolean' AFTER `personal_info_view_approved_at`,
    ADD COLUMN `lock_membership_approved_at` DATETIME DEFAULT NULL AFTER `personal_info_view_approved`,
    ADD COLUMN `lock_membership_approved` TINYINT(1)
        AS (`lock_membership_approved_at` IS NOT NULL) NOT NULL
        COMMENT 'lock_membership_approved_at as boolean' AFTER `lock_membership_approved_at`,
    ADD COLUMN `watch_approved_at` DATETIME DEFAULT NULL AFTER `lock_membership_approved`,
    ADD COLUMN `watch_approved` TINYINT(1)
        AS (`watch_approved_at` IS NOT NULL) NOT NULL COMMENT 'watch_approved_at as boolean' AFTER `watch_approved_at`;

ALTER TABLE `group_pending_requests`
    ADD COLUMN `personal_info_view_approved` TINYINT(1) NOT NULL DEFAULT 0
        COMMENT 'for join requests' AFTER `at`,
    ADD COLUMN `lock_membership_approved` TINYINT(1) NOT NULL DEFAULT 0
        COMMENT 'for join requests' AFTER `personal_info_view_approved`,
    ADD COLUMN `watch_approved` TINYINT(1) NOT NULL DEFAULT 0
        COMMENT 'for join requests' AFTER `lock_membership_approved`;

ALTER TABLE `group_managers`
    ADD COLUMN `can_edit_personal_info` TINYINT(1) NOT NULL DEFAULT 0
        COMMENT 'Can change member’s personal info, for those who have agreed (not visible to managers, only for specific uses)'
        AFTER `can_watch_members`;

UPDATE `groups` SET `require_lock_membership_approval_until` = `lock_user_deletion_until`;
ALTER TABLE `groups` DROP COLUMN `lock_user_deletion_until`;

UPDATE `groups_groups`
JOIN `groups` AS `parent` ON `parent`.`id` = `groups_groups`.`parent_group_id` AND
                             `parent`.`require_lock_membership_approval_until` IS NOT NULL
JOIN `groups` AS `child` ON `child`.`id` = `groups_groups`.`child_group_id` AND
                            `child`.`type` = 'UserSelf'
SET `groups_groups`.`lock_membership_approved_at` = NOW();

DROP VIEW IF EXISTS groups_groups_active;
CREATE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;

-- +migrate Down
ALTER TABLE `groups`
    DROP COLUMN `require_personal_info_access_approval`,
    ADD COLUMN `lock_user_deletion_until` date DEFAULT NULL
        COMMENT 'Prevent users from this group to delete their own user themselves until this date'
        AFTER `send_emails`;
UPDATE `groups` SET `lock_user_deletion_until` = CAST(`require_lock_membership_approval_until` AS DATE);

ALTER TABLE `groups`
    DROP COLUMN `require_lock_membership_approval_until`,
    DROP COLUMN `require_watch_approval`;

ALTER TABLE `groups_groups`
    DROP COLUMN `personal_info_view_approved_at`,
    DROP COLUMN `personal_info_view_approved`,
    DROP COLUMN `lock_membership_approved_at`,
    DROP COLUMN `lock_membership_approved`,
    DROP COLUMN `watch_approved_at`,
    DROP COLUMN `watch_approved`;

ALTER TABLE `group_pending_requests`
    DROP COLUMN `personal_info_view_approved`,
    DROP COLUMN `lock_membership_approved`,
    DROP COLUMN `watch_approved`;

ALTER TABLE `group_managers`
    DROP COLUMN `can_edit_personal_info`;

DROP VIEW IF EXISTS groups_groups_active;
CREATE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;

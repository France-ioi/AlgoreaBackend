-- +migrate Up
ALTER TABLE `groups`
    ADD COLUMN `require_personal_info_access_approval` ENUM('none', 'view', 'edit') NOT NULL DEFAULT 'none'
        AFTER `lock_user_deletion_until`,
    ADD COLUMN `require_lock_membership_approval_until` DATETIME DEFAULT NULL
        AFTER `require_personal_info_access_approval`,
    ADD COLUMN `require_watch_approval` TINYINT(1) NOT NULL DEFAULT 0
        AFTER `require_lock_membership_approval_until`;

ALTER TABLE `groups_groups`
    ADD COLUMN `personal_info_access_approved_at` DATETIME DEFAULT NULL AFTER `expires_at`,
    ADD COLUMN `personal_info_access_approved` TINYINT(1)
        AS (IFNULL(`personal_info_access_approved_at`, 0)) NOT NULL
        COMMENT 'personal_info_access_approved_at as boolean' AFTER `personal_info_access_approved_at`,
    ADD COLUMN `lock_membership_approved_at` DATETIME DEFAULT NULL AFTER `personal_info_access_approved`,
    ADD COLUMN `lock_membership_approved` TINYINT(1)
        AS (IFNULL(`lock_membership_approved_at`, 0)) NOT NULL
        COMMENT 'lock_membership_approved_at as boolean' AFTER `lock_membership_approved_at`,
    ADD COLUMN `watch_approved_at` DATETIME DEFAULT NULL AFTER `lock_membership_approved`,
    ADD COLUMN `watch_approved` TINYINT(1)
        AS (IFNULL(`watch_approved_at`, 0)) NOT NULL COMMENT 'watch_approved_at as boolean' AFTER `watch_approved_at`;

ALTER TABLE `group_pending_requests`
    ADD COLUMN `personal_info_access_approved_at` DATETIME DEFAULT NULL
        COMMENT 'for join requests' AFTER `at`,
    ADD COLUMN `lock_membership_approved_at` DATETIME DEFAULT NULL
        COMMENT 'for join requests' AFTER `personal_info_access_approved_at`,
    ADD COLUMN `watch_approved_at` DATETIME DEFAULT NULL
        COMMENT 'for join requests' AFTER `lock_membership_approved_at`;

ALTER TABLE `group_managers`
    ADD COLUMN `can_edit_personal_info` TINYINT(1) NOT NULL DEFAULT 0
        COMMENT 'Can change member’s personal info, for those who have agreed (not visible to managers, only for specific uses)'
        AFTER `can_watch_members`;

-- +migrate Down
ALTER TABLE `groups`
    DROP COLUMN `require_personal_info_access_approval`,
    DROP COLUMN `require_lock_membership_approval_until`,
    DROP COLUMN `require_watch_approval`;

ALTER TABLE `groups_groups`
    DROP COLUMN `personal_info_access_approved_at`,
    DROP COLUMN `personal_info_access_approved`,
    DROP COLUMN `lock_membership_approved_at`,
    DROP COLUMN `lock_membership_approved`,
    DROP COLUMN `watch_approved_at`,
    DROP COLUMN `watch_approved`;

ALTER TABLE `group_pending_requests`
    DROP COLUMN `personal_info_access_approved_at`,
    DROP COLUMN `lock_membership_approved_at`,
    DROP COLUMN `watch_approved_at`;

ALTER TABLE `group_managers`
    DROP COLUMN `can_edit_personal_info`;

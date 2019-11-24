-- +migrate Up
CREATE TABLE `group_membership_changes` (
    `group_id` BIGINT(20) NOT NULL,
    `member_id` BIGINT(20) NOT NULL,
    `at` DATETIME NOT NULL DEFAULT NOW() COMMENT 'Time of the action',
    `action` ENUM('invitation_created', 'invitation_withdrawn', 'invitation_refused', 'invitation_accepted',
                  'join_request_created', 'join_request_withdrawn', 'join_request_refused', 'join_request_accepted',
                  'left', 'removed', 'joined_by_code', 'added_directly', 'expired'),
    `initiator_id` BIGINT(20) DEFAULT NULL COMMENT 'The user who initiated the action (if any), typically the group owner/manager or the member himself',
    PRIMARY KEY (`group_id`, `member_id`, `at`),
    INDEX `group_id_member_id` (`group_id`, `member_id`),
    INDEX `action` (`action`),
    INDEX `group_id_member_id_at_desc` (`group_id`, `member_id`, `at` DESC),
    CONSTRAINT `fk_group_membership_changes_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_group_membership_changes_member_id_groups_id` FOREIGN KEY (`member_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_group_membership_changes_initiator_id_users_group_id` FOREIGN KEY (`initiator_id`) REFERENCES `users`(`group_id`) ON DELETE SET NULL
) COMMENT 'Stores the history of group membership changes' ENGINE=InnoDB DEFAULT CHARSET=utf8;

DELETE `groups_groups` FROM `groups_groups`
    LEFT JOIN `groups` ON `groups`.`id` = `groups_groups`.`parent_group_id`
    WHERE `groups`.`id` IS NULL;

DELETE `groups_groups` FROM `groups_groups`
    LEFT JOIN `groups` ON `groups`.`id` = `groups_groups`.`child_group_id`
    WHERE `groups`.`id` IS NULL;

DELETE FROM `groups_groups` WHERE `type` = 'removed' AND `type_changed_at` IS NULL;

INSERT INTO `group_membership_changes`
    SELECT `parent_group_id`, `child_group_id`, `type_changed_at`,
           CASE `type`
               WHEN 'invitationSent' THEN 'invitation_created'
               WHEN 'invitationRefused' THEN 'invitation_refused'
               WHEN 'invitationAccepted' THEN 'invitation_accepted'
               WHEN 'requestSent' THEN 'join_request_created'
               WHEN 'requestRefused' THEN 'join_request_refused'
               WHEN 'requestAccepted' THEN 'join_request_accepted'
               WHEN 'joinedByCode' THEN 'joined_by_code'
               ELSE `type`
           END,
           IF(`type` IN ('requestSent', 'invitationAccepted', 'invitationRefused', 'left', 'joinedByCode'),
               `child_group_id`, `inviting_user_id`)
    FROM `groups_groups`
    WHERE `type` != 'direct';

DELETE FROM `groups_groups` WHERE `type` IN ('invitationRefused', 'requestRefused', 'removed', 'left');

ALTER TABLE `groups_groups`
    MODIFY COLUMN `type`
        ENUM('invitationSent','requestSent','invitationAccepted','requestAccepted','direct','joinedByCode') NOT NULL
            DEFAULT 'direct',
    DROP FOREIGN KEY `fk_groups_groups_inviting_user_id_users_group_id`,
    DROP COLUMN `inviting_user_id`;

DROP VIEW `groups_groups_active`;
CREATE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;

-- +migrate Down
ALTER TABLE `groups_groups`
    MODIFY COLUMN `type`
        ENUM('invitationSent','requestSent','invitationAccepted','requestAccepted','invitationRefused','requestRefused',
            'removed','left','direct','joinedByCode') NOT NULL DEFAULT 'direct',
    ADD COLUMN `inviting_user_id` BIGINT(20) DEFAULT NULL
        COMMENT 'User (one of the admins of the parent group) who initiated the invitation or accepted the request'
        AFTER `role`,
    ADD CONSTRAINT `fk_groups_groups_inviting_user_id_users_group_id` FOREIGN KEY (`inviting_user_id`)
        REFERENCES `users` (`group_id`) ON DELETE SET NULL;

UPDATE `groups_groups`
    SET `inviting_user_id` = (
        SELECT `initiator_id`
        FROM `group_membership_changes`
        WHERE `group_membership_changes`.`group_id` = `groups_groups`.`parent_group_id`
          AND `group_membership_changes`.`member_id` = `groups_groups`.`child_group_id`
          AND `group_membership_changes`.`action` = 'invitation_created'
        ORDER BY at DESC
        LIMIT 1
    )
    WHERE `type` = 'invitationSent';

INSERT INTO `groups_groups` (`parent_group_id`, `child_group_id`, `type_changed_at`, `type`, `inviting_user_id`)
    WITH `last_actions` AS (
        SELECT `group_id`, `member_id`, MAX(`at`) AS `at` FROM `group_membership_changes` GROUP BY `group_id`, `member_id`
    )
    SELECT `group_id`, `member_id`, `at`,
           CASE `action`
               WHEN 'invitation_refused' THEN 'invitationRefused'
               WHEN 'join_request_refused' THEN 'requestRefused'
               ELSE `action`
           END,
           IF(`action` IN ('invitation_refused', 'left'), NULL, `initiator_id`)
    FROM `group_membership_changes`
    LEFT JOIN `last_actions` USING (`group_id`, `member_id`, `at`)
    WHERE `action` IN ('invitation_refused', 'join_request_refused', 'removed', 'left');

DROP TABLE `group_membership_changes`;

DROP VIEW `groups_groups_active`;
CREATE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;

-- +migrate Up
CREATE TABLE `group_pending_requests` (
    `group_id` BIGINT(20) NOT NULL,
    `member_id` BIGINT(20) NOT NULL,
    `type` ENUM('invitation', 'join_request'),
    `at` DATETIME NOT NULL DEFAULT NOW(),
    PRIMARY KEY (`group_id`, `member_id`),
    INDEX `group_id_member_id_at_desc` (`group_id`, `member_id`, `at` DESC),
    CONSTRAINT `fk_group_pending_requests_group_id_groups_id` FOREIGN KEY (`group_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_group_pending_requests_member_id_groups_id` FOREIGN KEY (`member_id`) REFERENCES `groups`(`id`) ON DELETE CASCADE
) COMMENT 'Requests that require an action from a user (group owner/manager or member)' ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO `group_pending_requests`
    SELECT `parent_group_id`, `child_group_id`,
           IF(`type` = 'requestSent', 'join_request', 'invitation'),
           `type_changed_at`
    FROM `groups_groups`
    WHERE `type` IN ('requestSent', 'invitationSent');

DELETE FROM `groups_groups` WHERE `type` IN ('requestSent', 'invitationSent');

ALTER TABLE `groups_groups`
    DROP INDEX `parent_type`,
    DROP COLUMN `type`,
    DROP COLUMN `type_changed_at`;

DROP TRIGGER `before_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF (OLD.child_group_id != NEW.child_group_id OR OLD.parent_group_id != NEW.parent_group_id OR NOT OLD.expires_at <=> NEW.expires_at) THEN
        INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) (
            SELECT `groups_ancestors`.`child_group_id`, 'todo'
            FROM `groups_ancestors`
            WHERE `groups_ancestors`.`ancestor_group_id` = OLD.`child_group_id`
        ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        DELETE `bridges` FROM `groups_ancestors` `child_descendants`
                                  JOIN `groups_ancestors` `parent_ancestors`
                                  JOIN `groups_ancestors` `bridges`
                                       ON (`bridges`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id` AND
                                           `bridges`.`child_group_id` = `child_descendants`.`child_group_id`)
        WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id` AND
                `child_descendants`.`ancestor_group_id` = OLD.`child_group_id`;

        INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    END IF;
END
-- +migrate StatementEnd

DROP VIEW `groups_groups_active`;
CREATE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;

-- +migrate Down
ALTER TABLE `groups_groups`
    ADD COLUMN `type`
        ENUM('invitationSent','requestSent','invitationAccepted','requestAccepted','invitationRefused','requestRefused',
            'removed','left','direct','joinedByCode') NOT NULL DEFAULT 'direct' AFTER `child_order`,
    ADD COLUMN `type_changed_at` datetime DEFAULT NULL COMMENT 'When was the type last Changed.' AFTER `role`,
    ADD INDEX `parent_type` (`parent_group_id`, `type`);

DROP TRIGGER `before_update_groups_groups`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_groups` BEFORE UPDATE ON `groups_groups` FOR EACH ROW BEGIN
    IF (OLD.child_group_id != NEW.child_group_id OR OLD.parent_group_id != NEW.parent_group_id OR OLD.type != NEW.type OR NOT OLD.expires_at <=> NEW.expires_at) THEN
        INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) VALUES (OLD.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        INSERT IGNORE INTO `groups_propagate` (`id`, `ancestors_computation_state`) (
            SELECT `groups_ancestors`.`child_group_id`, 'todo'
            FROM `groups_ancestors`
            WHERE `groups_ancestors`.`ancestor_group_id` = OLD.`child_group_id`
        ) ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
        DELETE `bridges` FROM `groups_ancestors` `child_descendants`
                                  JOIN `groups_ancestors` `parent_ancestors`
                                  JOIN `groups_ancestors` `bridges`
                                       ON (`bridges`.`ancestor_group_id` = `parent_ancestors`.`ancestor_group_id` AND
                                           `bridges`.`child_group_id` = `child_descendants`.`child_group_id`)
        WHERE `parent_ancestors`.`child_group_id` = OLD.`parent_group_id` AND
                `child_descendants`.`ancestor_group_id` = OLD.`child_group_id`;

        INSERT IGNORE INTO `groups_propagate` (id, ancestors_computation_state) VALUES (NEW.child_group_id, 'todo')
        ON DUPLICATE KEY UPDATE `ancestors_computation_state` = 'todo';
    END IF;
END
-- +migrate StatementEnd

UPDATE `groups_groups`
SET `type` = IFNULL(
        (SELECT CASE `group_membership_changes`.`action`
                    WHEN 'invitation_created' THEN 'invitationSent'
                    WHEN 'invitation_refused' THEN 'invitationRefused'
                    WHEN 'invitation_accepted' THEN 'invitationAccepted'
                    WHEN 'join_request_created' THEN 'requestSent'
                    WHEN 'join_request_refused' THEN 'requestRefused'
                    WHEN 'join_request_accepted' THEN 'requestAccepted'
                    WHEN 'joined_by_code' THEN 'joinedByCode'
                    WHEN 'added_directly' THEN 'direct'
                    ELSE `group_membership_changes`.`action`
                    END
         FROM `group_membership_changes`
         WHERE `group_membership_changes`.`group_id` = `groups_groups`.`parent_group_id`
           AND `group_membership_changes`.`member_id` = `groups_groups`.`child_group_id`
         ORDER BY `at` DESC
         LIMIT 1), `type`);

UPDATE `groups_groups`
SET `type_changed_at` = (
    SELECT `at`
    FROM `group_membership_changes`
    WHERE `group_membership_changes`.`group_id` = `groups_groups`.`parent_group_id`
      AND `group_membership_changes`.`member_id` = `groups_groups`.`child_group_id`
      AND CASE `groups_groups`.`type`
        WHEN 'invitationSent' THEN 'invitation_created'
        WHEN 'invitationRefused' THEN 'invitation_refused'
        WHEN 'invitationAccepted' THEN 'invitation_accepted'
        WHEN 'requestSent' THEN 'join_request_created'
        WHEN 'requestRefused' THEN 'join_request_refused'
        WHEN 'requestAccepted' THEN 'join_request_accepted'
        WHEN 'joinedByCode' THEN 'joined_by_code'
        WHEN 'direct' THEN 'added_directly'
        ELSE `groups_groups`.`type`
      END = `group_membership_changes`.`action`
    ORDER BY `at` DESC
    LIMIT 1
);

INSERT INTO `groups_groups` (`parent_group_id`, `child_group_id`, `type_changed_at`, `type`)
    WITH `last_actions` AS (
        SELECT `group_id`, `member_id`, MAX(`at`) AS `at` FROM `group_membership_changes` GROUP BY `group_id`, `member_id`
    )
    SELECT `group_id`, `member_id`, `last_actions`.`at`,
           IF(`group_pending_requests`.`type` = 'invitation', 'invitationSent', 'requestSent')
    FROM `group_pending_requests`
    LEFT JOIN `last_actions` USING (`group_id`, `member_id`)
    WHERE `group_pending_requests`.`type` IN ('join_request', 'invitation');

DROP TABLE `group_pending_requests`;

DROP VIEW groups_groups_active;
CREATE VIEW groups_groups_active AS SELECT * FROM groups_groups WHERE NOW() < expires_at;

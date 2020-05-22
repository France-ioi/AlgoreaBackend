-- +migrate Up
ALTER TABLE `groups`
    RENAME COLUMN `activity_id` TO `root_activity_id`,
    DROP FOREIGN KEY `fk_groups_activity_id_items_id`,
    DROP KEY `fk_groups_activity_id_items_id`,
    ADD CONSTRAINT `fk_groups_root_activity_id_items_id`
        FOREIGN KEY (`root_activity_id`) REFERENCES `items`(`id`) ON DELETE SET NULL,
    ADD COLUMN `root_skill_id` BIGINT(20) DEFAULT NULL COMMENT 'Root skill associated with this group'
        AFTER `root_activity_id`,
    ADD CONSTRAINT `fk_groups_root_skill_id_items_id`
        FOREIGN KEY (`root_skill_id`) REFERENCES `items`(`id`) ON DELETE SET NULL;

-- +migrate Down
ALTER TABLE `groups`
    RENAME COLUMN `root_activity_id` TO `activity_id`,
    DROP FOREIGN KEY `fk_groups_root_activity_id_items_id`,
    DROP KEY `fk_groups_root_activity_id_items_id`,
    ADD CONSTRAINT `fk_groups_activity_id_items_id`
        FOREIGN KEY (`activity_id`) REFERENCES `items`(`id`) ON DELETE SET NULL,
    DROP FOREIGN KEY `fk_groups_root_skill_id_items_id`,
    DROP KEY `fk_groups_root_skill_id_items_id`,
    DROP COLUMN `root_skill_id`;

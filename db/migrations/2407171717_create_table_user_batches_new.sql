-- +migrate Up

CREATE TABLE IF NOT EXISTS `user_batches_v2` (
    `group_prefix` VARCHAR(13) NOT NULL COMMENT 'Authorized (first) part of the full login prefix',
    `custom_prefix` VARCHAR(14) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL
        COMMENT 'Second part of the full login prefix, given by the user that created the batch',
    `size` MEDIUMINT UNSIGNED NOT NULL COMMENT 'Number of users created in this batch',
    `creator_id` BIGINT(20) DEFAULT NULL,
    `created_at` DATETIME NOT NULL DEFAULT NOW(),
    PRIMARY KEY (`group_prefix`, `custom_prefix`),
    CONSTRAINT `ck_user_batches_v2_custom_prefix` CHECK (REGEXP_LIKE(`custom_prefix`, '^[a-z0-9-]+$')),
    CONSTRAINT `fk_user_batches_v2_group_prefix_user_batch_prefixes_group_pref`
        FOREIGN KEY (`group_prefix`) REFERENCES `user_batch_prefixes`(`group_prefix`) ON DELETE RESTRICT,
    CONSTRAINT `fk_user_batches_v2_creator_id_users_group_id`
        FOREIGN KEY (`creator_id`) REFERENCES `users`(`group_id`) ON DELETE SET NULL
) ENGINE=InnoDB CHARSET=utf8
    COMMENT='Batches of users that were created (replaces user_batches which has been broken by a MySQL update)';

-- +migrate Down
/* we pretend the table was always named `user_batches_v2`, so we don't need to drop it */
;

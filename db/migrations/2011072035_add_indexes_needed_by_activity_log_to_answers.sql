-- +migrate Up
ALTER TABLE `answers` ADD INDEX `type_created_at_desc_item_id_participant_id_attempt_id_desc`
    (`type`,`created_at` DESC,`item_id`,`participant_id`,`attempt_id` DESC);

-- +migrate Down
ALTER TABLE `answers` DROP INDEX `type_created_at_desc_item_id_participant_id_attempt_id_desc`;

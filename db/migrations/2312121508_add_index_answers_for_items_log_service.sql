-- +migrate Up
ALTER TABLE `answers`
ADD INDEX `type_participant_id_item_id_created_at_desc_attempt_id_desc`
(`type`, `participant_id`, `item_id`, `created_at` DESC, `attempt_id` DESC);

-- +migrate Down
ALTER TABLE `answers`
DROP INDEX `type_participant_id_item_id_created_at_desc_attempt_id_desc`;

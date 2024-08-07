-- +migrate Up
ALTER TABLE `answers`
  ADD INDEX `created_at_d_item_id_participant_id_attempt_id_d_type_d_id_autho`
    (`created_at` DESC, `item_id`, `participant_id`, `attempt_id` DESC, `type` DESC, `id`, `author_id`);

-- +migrate Down
ALTER TABLE `answers`
  DROP INDEX `created_at_d_item_id_participant_id_attempt_id_d_type_d_id_autho`;

-- +goose Up
ALTER TABLE `answers`
  DROP INDEX `created_at_d_item_id_participant_id_attempt_id_d_type_d_id_autho`,
  ADD COLUMN `activity_type_int` TINYINT GENERATED ALWAYS AS (
    (CASE `type` WHEN 'Submission' THEN 2 WHEN 'Saved' THEN 4 WHEN 'Current' THEN 5 END)
  ) VIRTUAL,
  ADD INDEX `created_at_d_item_id_participant_id_attempt_id_d_atype_d_id_a_t`
    (`created_at` DESC, `item_id`, `participant_id`, `attempt_id` DESC, `activity_type_int` DESC, `id`, `author_id`, `type`);

-- +goose Down
ALTER TABLE `answers`
  DROP INDEX `created_at_d_item_id_participant_id_attempt_id_d_atype_d_id_a_t`,
  DROP COLUMN `activity_type_int`,
  ADD INDEX `created_at_d_item_id_participant_id_attempt_id_d_type_d_id_autho`
    (`created_at` DESC, `item_id`, `participant_id`, `attempt_id` DESC, `type` DESC, `id`, `author_id`);

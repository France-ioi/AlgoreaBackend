-- +goose Up
ALTER TABLE `results`
  ADD INDEX `item_id_participant_id_started_at_desc_attempt_id_desc`
    (`item_id`, `participant_id`, `started_at` DESC, `attempt_id` DESC),
  ADD INDEX `item_id_participant_id_validated_at_desc_attempt_id_desc`
    (`item_id`, `participant_id`, `validated_at` DESC, `attempt_id` DESC);

-- +goose Down
ALTER TABLE `results`
  DROP INDEX `item_id_participant_id_started_at_desc_attempt_id_desc`,
  DROP INDEX `item_id_participant_id_validated_at_desc_attempt_id_desc`;

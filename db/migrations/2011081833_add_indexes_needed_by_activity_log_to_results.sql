-- +migrate Up
ALTER TABLE `results`
    ADD INDEX `started_at_desc_item_id_participant_id_attempt_id_desc` (`started_at` DESC,`item_id`,`participant_id`,`attempt_id` DESC),
    ADD INDEX `validated_at_desc_item_id_participant_id_attempt_id_desc` (`validated_at` DESC,`item_id`,`participant_id`,`attempt_id` DESC);

-- +migrate Down
ALTER TABLE `results`
    DROP INDEX `started_at_desc_item_id_participant_id_attempt_id_desc`,
    DROP INDEX `validated_at_desc_item_id_participant_id_attempt_id_desc`;

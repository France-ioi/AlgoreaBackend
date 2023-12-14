-- +migrate Up
ALTER TABLE `answers`
  ADD INDEX `c_at_desc_item_id_part_id_attempt_id_desc_type_desc_answers`
  (`created_at` DESC,`item_id`,`participant_id`,`attempt_id` DESC, `type` DESC, id)
;

-- +migrate Down
ALTER TABLE `answers`
  DROP INDEX `c_at_desc_item_id_part_id_attempt_id_desc_type_desc_answers`;

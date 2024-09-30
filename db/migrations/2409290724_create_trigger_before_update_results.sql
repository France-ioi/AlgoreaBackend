-- +migrate Up
DROP TRIGGER IF EXISTS `before_update_results`;
-- +migrate StatementBegin
CREATE TRIGGER `before_update_results`
  BEFORE UPDATE
  ON `results`
  FOR EACH ROW
BEGIN
  IF NEW.recomputing_state = 'recomputing' THEN
    SET NEW.recomputing_state = IF(
      NEW.latest_activity_at <=> OLD.latest_activity_at AND
      NEW.tasks_tried <=> OLD.tasks_tried AND
      NEW.tasks_with_help <=> OLD.tasks_with_help AND
      NEW.validated_at <=> OLD.validated_at AND
      NEW.score_computed <=> OLD.score_computed AND
      -- We always consider results with the default latest_activity_at as changed
      -- because they look like a newly inserted result for a chapter/skill.
      -- It makes sure that the newly inserted result is propagated.
      NEW.latest_activity_at <> '1000-01-01 00:00:00',
      'unchanged',
      'modified');
  END IF;
END;
-- +migrate StatementEnd

-- +migrate Down
DROP TRIGGER IF EXISTS `before_update_results`;

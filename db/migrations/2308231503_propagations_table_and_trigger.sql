-- +migrate Up

CREATE TABLE `propagations` (
  `propagation_id` BIGINT NOT NULL,
  `propagate` TINYINT DEFAULT 0 COMMENT 'MUST be set to 0. Is set in 1 in a trigger.',
  `scheduled_counter` BIGINT DEFAULT 0 COMMENT 'Set in Trigger only, used to make sure propagation is not triggered more than once before it is executed.',
  PRIMARY KEY (`propagation_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Used to know if a propagation has to be run. Has trigger to call propagation via AWS Lambda.';

-- +migrate StatementBegin
CREATE TRIGGER `before_update_propagation` BEFORE UPDATE ON `propagations` FOR EACH ROW
BEGIN
  IF OLD.propagate = 0 THEN
    # The following line is replaced by the config value of database.aws_aurora_propagation_trigger.
    -- %%aws_aurora_propagation_trigger%%

    -- Is incremented only when a propagation is scheduled and no propagation is scheduled yet.
    -- Used to test that we don't call the Lambda Trigger when not needed.
    SET NEW.scheduled_counter = OLD.scheduled_counter + 1;

    SET NEW.propagate = 1;
  END IF;
END;
-- +migrate StatementEnd


-- +migrate Down

DROP TRIGGER IF EXISTS `before_insert_propagation`;

DROP TABLE `propagations`;

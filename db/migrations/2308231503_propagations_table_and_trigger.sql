-- +migrate Up

CREATE TABLE `propagations` (
  `propagation_id` BIGINT NOT NULL,
  `propagate` TINYINT DEFAULT 0 COMMENT 'MUST be set to 0. Is set in 1 in a trigger.',
  PRIMARY KEY (`propagation_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Used to know if a propagation has to be run. Has trigger to call propagation via AWS Lambda.';

-- +migrate StatementBegin
CREATE TRIGGER `before_update_propagation` BEFORE UPDATE ON `propagations` FOR EACH ROW
BEGIN
  # This must be set on AWS Aurora.
  -- lambda_async('arn:aws:lambda:REGION:ACCOUNT_ID:function:Propagate', '{}');
  SET NEW.propagate = 1;
END;
-- +migrate StatementEnd


-- +migrate Down

DROP TRIGGER IF EXISTS `before_insert_propagation`;

DROP TABLE `propagations`;

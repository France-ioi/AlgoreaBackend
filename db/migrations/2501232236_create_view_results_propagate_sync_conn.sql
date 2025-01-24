-- +migrate Up
CREATE VIEW `results_propagate_sync_conn` AS
  SELECT * FROM `results_propagate_sync` WHERE `connection_id` = CONNECTION_ID();

-- +migrate Down
DROP VIEW `results_propagate_sync_conn`;

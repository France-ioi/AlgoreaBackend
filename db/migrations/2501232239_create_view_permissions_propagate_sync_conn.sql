-- +migrate Up
CREATE VIEW `permissions_propagate_sync_conn` AS
  SELECT * FROM `permissions_propagate_sync` WHERE `connection_id` = CONNECTION_ID();

-- +migrate Down
DROP VIEW `permissions_propagate_sync_conn`;

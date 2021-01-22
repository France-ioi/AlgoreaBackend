-- +migrate Up
UPDATE `results` SET`result_propagation_state` = 'to_be_propagated';

-- +migrate Down

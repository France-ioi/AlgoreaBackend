-- +migrate Up
ALTER TABLE `results`
  ADD COLUMN `recomputing_state` ENUM('recomputing', 'modified', 'unchanged') NOT NULL DEFAULT 'unchanged'
    COMMENT 'State of the result, used during recomputing' AFTER `help_requested`;

-- +migrate Down
ALTER TABLE `results` DROP COLUMN `recomputing_state`;

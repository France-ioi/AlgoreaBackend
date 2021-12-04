-- +migrate Up
ALTER TABLE `results`
  ADD COLUMN `help_requested` TINYINT(1) NOT NULL DEFAULT 0
    COMMENT 'Whether the participant has requested help for the item in this attempt'
    AFTER `latest_hint_at`;

-- +migrate Down
ALTER TABLE `results` DROP COLUMN `help_requested`;

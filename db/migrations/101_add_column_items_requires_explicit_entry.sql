-- +migrate Up
ALTER TABLE `items`
    ADD COLUMN `requires_explicit_entry` TINYINT(1) NOT NULL DEFAULT 0
        COMMENT 'Whether this item requires an explicit entry to be started (create an attempt)'
        AFTER `default_language_tag`;

UPDATE `items` SET `requires_explicit_entry` = IFNULL(`entry_participant_type` = 'Team' OR `duration` IS NOT NULL, 0);

-- +migrate Down
ALTER TABLE `items`
    DROP COLUMN `requires_explicit_entry`;

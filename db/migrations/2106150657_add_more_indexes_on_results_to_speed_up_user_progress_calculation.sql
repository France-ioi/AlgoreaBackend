-- +migrate Up
ALTER TABLE `results`
    ADD COLUMN `started` tinyint(1) GENERATED ALWAYS AS ((`started_at` IS NOT NULL)) VIRTUAL NOT NULL
        COMMENT 'Auto-generated from `started_at`' AFTER `tasks_tried`,
    ADD KEY `participant_id_item_id_latest_activity_at_desc` (`participant_id`,`item_id`,`latest_activity_at` DESC),
    ADD KEY `participant_id_item_id_started_started_at` (`participant_id`,`item_id`,`started`,`started_at`),
    ADD KEY `participant_id_item_id_validated_validated_at` (`participant_id`,`item_id`,`validated`,`validated_at`);

-- +migrate Down
ALTER TABLE `results`
    DROP KEY `participant_id_item_id_validated_validated_at`,
    DROP KEY `participant_id_item_id_started_started_at`,
    DROP KEY `participant_id_item_id_latest_activity_at_desc`,
    DROP COLUMN `started`;

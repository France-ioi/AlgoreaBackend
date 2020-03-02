-- +migrate Up
ALTER TABLE `groups_attempts`
    DROP INDEX `group_item_minus_score_best_answer_date_id`,
    DROP COLUMN `minus_score`,
    ADD INDEX `group_item_score_desc_best_answer_at` (`group_id`,`item_id`,`score` DESC,`best_answer_at`);

DROP TRIGGER IF EXISTS `before_insert_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_attempts` BEFORE INSERT ON `groups_attempts` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF; END
-- +migrate StatementEnd
DROP TRIGGER IF EXISTS `before_update_groups_attempts`;

-- +migrate Down
ALTER TABLE `groups_attempts`
    DROP INDEX `group_item_score_desc_best_answer_at`,
    ADD COLUMN `minus_score` float DEFAULT NULL;

UPDATE `groups_attempts` SET minus_score = -score;

ALTER TABLE `groups_attempts`
    ADD INDEX `group_item_minus_score_best_answer_date_id` (`group_id`,`item_id`,`minus_score`,`best_answer_at`);

DROP TRIGGER IF EXISTS `before_insert_groups_attempts`;
-- +migrate StatementBegin
CREATE TRIGGER `before_insert_groups_attempts` BEFORE INSERT ON `groups_attempts` FOR EACH ROW BEGIN IF (NEW.id IS NULL OR NEW.id = 0) THEN SET NEW.id = FLOOR(RAND() * 1000000000) + FLOOR(RAND() * 1000000000) * 1000000000; END IF ; SET NEW.minus_score = -NEW.score; END
-- +migrate StatementEnd
-- +migrate StatementBegin
CREATE TRIGGER `before_update_groups_attempts` BEFORE UPDATE ON `groups_attempts` FOR EACH ROW BEGIN IF NOT (OLD.`score` <=> NEW.`score`) THEN SET NEW.minus_score = -NEW.score; END IF; END
-- +migrate StatementEnd

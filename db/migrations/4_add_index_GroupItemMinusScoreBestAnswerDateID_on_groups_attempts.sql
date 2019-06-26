-- +migrate Up
ALTER TABLE `groups_attempts` ADD INDEX GroupItemMinusScoreBestAnswerDateID (idGroup, idItem, iMinusScore, sBestAnswerDate);


-- +migrate Down
ALTER TABLE `groups_attempts` DROP INDEX GroupItemMinusScoreBestAnswerDateID;


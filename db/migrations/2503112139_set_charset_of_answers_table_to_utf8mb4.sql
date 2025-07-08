-- +migrate Up
ALTER TABLE `answers`
  MODIFY `type` ENUM('Submission','Saved','Current') NOT NULL
    COMMENT '''Submission'' for answers submitted for grading, ''Saved'' for manual backups of answers, ''Current'' for automatic snapshots of the latest answers (unique for a user on an attempt)',
  MODIFY `state` mediumtext COMMENT 'Saved state (sent by the task platform)',
  MODIFY `answer` mediumtext COMMENT 'Saved answer (sent by the task platform)',
  DEFAULT CHARACTER SET utf8mb4;

-- +migrate Down
ALTER TABLE `answers` CONVERT TO CHARACTER SET utf8mb3;

-- +migrate Up
SET foreign_key_checks = 0;
ALTER TABLE `user_batch_prefixes`
  CHARACTER SET utf8mb4,
  MODIFY `group_prefix` varchar(13) NOT NULL COLLATE utf8mb4_bin
    COMMENT 'Prefix used in front of all batches';
ALTER TABLE `user_batches_v2`
  CHARACTER SET utf8mb4,
  MODIFY `group_prefix` varchar(13) NOT NULL COLLATE utf8mb4_bin
    COMMENT 'Authorized (first) part of the full login prefix',
  MODIFY `custom_prefix` varchar(14) NOT NULL COLLATE utf8mb4_bin
    COMMENT 'Second part of the full login prefix, given by the user that created the batch';
SET foreign_key_checks = 1;

-- +migrate Down
SET foreign_key_checks = 0;
ALTER TABLE `user_batch_prefixes`
  CHARACTER SET utf8mb3,
  MODIFY `group_prefix` varchar(13) NOT NULL
    COMMENT 'Prefix used in front of all batches';
ALTER TABLE `user_batches_v2`
  CHARACTER SET utf8mb3,
  MODIFY `group_prefix` varchar(13) NOT NULL
    COMMENT 'Authorized (first) part of the full login prefix',
  MODIFY `custom_prefix` varchar(14) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL
    COMMENT 'Second part of the full login prefix, given by the user that created the batch';
SET foreign_key_checks = 1;

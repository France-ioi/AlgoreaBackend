-- +goose Up
ALTER TABLE `items`
  MODIFY COLUMN `entry_min_admitted_members_ratio` enum('All','Half','One','None') NOT NULL DEFAULT 'All'
    COMMENT 'The ratio of members in the team (a user alone being considered as a team of one) who needs the “can_enter” permission so that the group can enter';
UPDATE `items` SET `entry_min_admitted_members_ratio` = 'All';

-- +goose Down
-- Data change is not reverted: original per-item values are lost.
ALTER TABLE `items`
  MODIFY COLUMN `entry_min_admitted_members_ratio` enum('All','Half','One','None') NOT NULL DEFAULT 'None'
    COMMENT 'The ratio of members in the team (a user alone being considered as a team of one) who needs the “can_enter” permission so that the group can enter';

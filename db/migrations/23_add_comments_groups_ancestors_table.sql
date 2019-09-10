-- +migrate Up
ALTER TABLE `groups_ancestors`
  COMMENT 'This table stores all child-ancestor relationships for groups (a group is its own ancestor). It is a cache table that can be recomputed based on the content of groups_groups.',
  MODIFY COLUMN `bIsSelf` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether idGroupAncestor = idGroupChild.';

-- +migrate Down
ALTER TABLE `groups_ancestors`
  COMMENT '',
  MODIFY COLUMN `bIsSelf` tinyint(1) NOT NULL DEFAULT '0';

-- +migrate Up
ALTER TABLE `groups_ancestors`
    COMMENT 'All ancestor relationships for groups, given that a group is its own ancestor and team ancestors are not propagated to their members. It is a cache table that can be recomputed based on the content of groups_groups.';

-- +migrate Down
ALTER TABLE `groups_ancestors`
    COMMENT 'This table stores all child-ancestor relationships for groups (a group is its own ancestor). It is a cache table that can be recomputed based on the content of groups_groups.';

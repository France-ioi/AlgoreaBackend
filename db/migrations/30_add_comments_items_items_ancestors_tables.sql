-- +migrate Up
ALTER TABLE `items_ancestors`
  COMMENT 'All child-ancestor relationships (a item is not its own ancestor). Cache table that can be recomputed based on the content of groups_groups.';
ALTER TABLE `items_items`
  COMMENT 'Parent-child (N-N) relationship between items (acyclic graph)',
  MODIFY COLUMN `iChildOrder` int(11) NOT NULL COMMENT 'Position, relative to its siblings, when displaying all the children of the parent. If multiple items have the same iChildOrder, they will be sorted in a random way, specific to each user (a user will always see the items in the same order).',
  MODIFY COLUMN `sCategory` enum('Undefined','Discovery','Application','Validation','Challenge') NOT NULL DEFAULT 'Undefined' COMMENT 'Tag that indicates the role of this item, from the point of view of the parent item''s validation criteria. Also gives indication to the user of the role of the item.',
  MODIFY COLUMN `bAlwaysVisible` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Whether the title of this child should always be visible within the parent (gray), even if the user did not unlock the access.',
  MODIFY COLUMN `bAccessRestricted` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'Whether this item is locked by default within the parent. If false, anyone who has access to the parent will also have access to the child.',
  MODIFY COLUMN `iDifficulty` int(11) NOT NULL COMMENT 'Indication of the difficulty of this item relative to its siblings.';

-- +migrate Down
ALTER TABLE `groups_ancestors`
  COMMENT '';
ALTER TABLE `items_items`
  COMMENT '',
  MODIFY COLUMN `iChildOrder` int(11) NOT NULL,
  MODIFY COLUMN `sCategory` enum('Undefined','Discovery','Application','Validation','Challenge') NOT NULL DEFAULT 'Undefined',
  MODIFY COLUMN `bAlwaysVisible` tinyint(1) NOT NULL DEFAULT '0',
  MODIFY COLUMN `bAccessRestricted` tinyint(1) NOT NULL DEFAULT '1',
  MODIFY COLUMN `iDifficulty` int(11) NOT NULL;


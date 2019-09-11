-- +migrate Up
ALTER TABLE `groups_login_prefixes`
  COMMENT 'Used to keep track of prefixes of logins that were generated in batch and attached to this group. Only the prefix is stored, not the actual usernames.';

-- +migrate Down
ALTER TABLE `groups_login_prefixes`
  COMMENT '';

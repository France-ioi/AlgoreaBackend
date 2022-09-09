-- +migrate Up
ALTER TABLE `users`
  ADD COLUMN `latest_profile_sync_at` DATETIME DEFAULT NULL
    COMMENT 'Last time when the profile was synced with the login module'
    AFTER `latest_activity_at`;

-- +migrate Down
ALTER TABLE `users` DROP COLUMN `latest_profile_sync_at`;

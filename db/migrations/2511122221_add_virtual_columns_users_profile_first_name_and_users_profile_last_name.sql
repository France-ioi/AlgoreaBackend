-- +goose Up
ALTER TABLE `users`
  ADD COLUMN `profile_first_name` TEXT GENERATED ALWAYS AS (
    IF(JSON_TYPE(profile->'$.first_name') = 'NULL', NULL, profile->>'$.first_name')
  ) VIRTUAL,
  ADD COLUMN `profile_last_name` TEXT GENERATED ALWAYS AS (
    IF(JSON_TYPE(profile->'$.last_name') = 'NULL', NULL, profile->>'$.last_name')
  ) VIRTUAL;

-- +goose Down
ALTER TABLE `users`
  DROP COLUMN `profile_first_name`,
  DROP COLUMN `profile_last_name`;

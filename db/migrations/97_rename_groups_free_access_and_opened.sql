-- +migrate Up
ALTER TABLE `groups`
    CHANGE COLUMN `opened` `is_open` TINYINT(1) NOT NULL DEFAULT '0'
        COMMENT 'Whether it appears to users as open to new members, i.e. the users can join using the code or create a join request',
    CHANGE COLUMN `free_access` `is_public` TINYINT(1) NOT NULL DEFAULT '0'
        COMMENT 'Whether it is visible to all users (through search) and open to join requests';


-- +migrate Down
ALTER TABLE `groups`
    CHANGE COLUMN `is_open` `opened` TINYINT(1) NOT NULL DEFAULT '0'
        COMMENT 'Can users still join this group or request to join it?',
    CHANGE COLUMN `is_public` `free_access` TINYINT(1) NOT NULL DEFAULT '0'
        COMMENT 'Can users search for this group and ask to join it?';

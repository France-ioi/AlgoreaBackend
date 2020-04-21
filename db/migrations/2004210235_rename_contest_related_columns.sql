-- +migrate Up
ALTER TABLE `items`
    CHANGE COLUMN `contest_participants_group_id` `participants_group_id` BIGINT(20) DEFAULT NULL
        COMMENT 'Group to which all the entered participants (users or teams) belong. Must not be null for an explicit-entry item.',
    CHANGE COLUMN `contest_max_team_size` `entry_max_team_size` INT(11) NOT NULL DEFAULT '0'
        COMMENT 'The maximum number of members a team can have to enter',
    RENAME COLUMN `contest_entering_condition` TO `entry_min_allowed_members`;

ALTER TABLE `groups`
    CHANGE COLUMN `open_contest` `open_activity_when_joining` TINYINT(1) NOT NULL DEFAULT '0'
        COMMENT 'Whether the activity should be started for participants as soon as they join the group';

-- +migrate Down
ALTER TABLE `items`
    CHANGE COLUMN `participants_group_id` `contest_participants_group_id` BIGINT(20) DEFAULT NULL
        COMMENT 'Group to which all the contest participants (users or teams) belong. Must not be null for a contest.',
    CHANGE COLUMN `entry_max_team_size` `contest_max_team_size` INT(11) NOT NULL DEFAULT '0'
        COMMENT 'The maximum number of members a team can have to enter the contest',
    RENAME COLUMN `entry_min_allowed_members` TO `contest_entering_condition`;


ALTER TABLE `groups`
    CHANGE COLUMN `open_activity_when_joining` `open_contest` TINYINT(1) NOT NULL DEFAULT '0'
        COMMENT 'If true and the group is associated through activity_id with an item that is a contest, the contest should be started for this user as soon as he joins the group';

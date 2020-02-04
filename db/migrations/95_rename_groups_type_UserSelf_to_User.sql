-- +migrate Up
ALTER TABLE `groups`
    MODIFY COLUMN `type` ENUM('Class','Team','Club','Friends','Other','UserSelf','User','Base','ContestParticipants') NOT NULL;

UPDATE `groups` SET `type` = 'User' WHERE `type` = 'UserSelf';

ALTER TABLE `groups`
    MODIFY COLUMN `type` ENUM('Class','Team','Club','Friends','Other','User','Base','ContestParticipants') NOT NULL;

-- +migrate Down
ALTER TABLE `groups`
    MODIFY COLUMN `type` ENUM('Class','Team','Club','Friends','Other','UserSelf','User','Base','ContestParticipants') NOT NULL;

UPDATE `groups` SET `type` = 'UserSelf' WHERE `type` = 'User';

ALTER TABLE `groups`
    MODIFY COLUMN `type` ENUM('Class','Team','Club','Friends','Other','UserSelf','Base','ContestParticipants') NOT NULL;

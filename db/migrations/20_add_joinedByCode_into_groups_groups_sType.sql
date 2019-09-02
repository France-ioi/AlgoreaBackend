-- +migrate Up
ALTER TABLE `groups_groups` MODIFY COLUMN `sType` enum('invitationSent','requestSent','invitationAccepted','requestAccepted','invitationRefused','requestRefused','removed','left','direct','joinedByCode') NOT NULL DEFAULT 'direct';

-- +migrate Down
UPDATE `groups_groups` SET `sType` = 'requestAccepted' WHERE `sType` = 'joinedByCode';
ALTER TABLE `groups_groups` MODIFY COLUMN `sType` enum('invitationSent','requestSent','invitationAccepted','requestAccepted','invitationRefused','requestRefused','removed','left','direct') NOT NULL DEFAULT 'direct';

-- +migrate Up
ALTER TABLE `history_groups_groups` MODIFY COLUMN `sType` enum('invitationSent','requestSent','invitationAccepted','requestAccepted','invitationRefused','requestRefused','removed','left','direct','joinedByCode') NOT NULL DEFAULT 'direct';

-- +migrate Down
UPDATE `history_groups_groups` SET `sType` = 'requestAccepted' WHERE `sType` = 'joinedByCode';
ALTER TABLE `history_groups_groups` MODIFY COLUMN `sType` enum('invitationSent','requestSent','invitationAccepted','requestAccepted','invitationRefused','requestRefused','removed','left','direct') NOT NULL DEFAULT 'direct';

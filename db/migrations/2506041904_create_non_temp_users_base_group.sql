-- +migrate Up
INSERT INTO `groups` (`id`, `name`, `type`, `text_id`, `description`) VALUE
  (4, 'NonTempUsers', 'Base', 'NonTempUsers', 'non-temporary users');

-- +migrate Down
DELETE FROM `groups` WHERE `id`=4 AND `type`='Base' AND `text_id`='NonTempUsers';

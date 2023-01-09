-- +migrate Up
UPDATE `items` SET `type`='Task' WHERE `type`='Course';
ALTER TABLE `items` CHANGE `type` `type` ENUM ('Chapter', 'Task', 'Skill');

-- +migrate Down
ALTER TABLE `items` CHANGE `type` `type` ENUM ('Chapter', 'Task', 'Course', 'Skill');

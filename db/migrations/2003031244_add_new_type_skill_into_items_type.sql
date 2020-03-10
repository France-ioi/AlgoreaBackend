-- +migrate Up
ALTER TABLE `items`
    MODIFY COLUMN `type` ENUM('Chapter','Task','Course','Skill') NOT NULL;

-- +migrate Down
DELETE FROM `items` WHERE `type` = 'Skill';
ALTER TABLE `items`
    MODIFY COLUMN `type` ENUM('Chapter','Task','Course') NOT NULL;

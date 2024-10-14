-- +migrate Up
ALTER TABLE `items` MODIFY `type` ENUM('Chapter','Task','Skill') NOT NULL;

-- +migrate Down
ALTER TABLE `items` MODIFY `type` ENUM('Chapter','Task','Skill') DEFAULT NULL;


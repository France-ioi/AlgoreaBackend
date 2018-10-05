-- +migrate Up
CREATE TABLE groups (id int);

-- +migrate Down
DROP TABLE groups;
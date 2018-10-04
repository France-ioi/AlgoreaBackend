-- +migrate Up
CREATE TABLE people (id int);

-- +migrate Down
DROP TABLE people;
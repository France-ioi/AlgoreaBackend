-- +migrate Up
CREATE TABLE `stopwords`(value VARCHAR(30)) ENGINE = INNODB
  COMMENT 'Stopwords for the fulltext search. The table is empty on purpose. It is used to prevent the fulltext search from using the default stopwords list.
All the MySQL fulltext indexes would have to be recreated with innodb_ft_server_stopword_table pointing to this table if we wanted to add some new stopwords.
Also the stopwords would have to be filtered out from searched strings in the application code.';

-- +migrate Down
DROP TABLE `stopwords`;

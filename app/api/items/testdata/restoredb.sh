#!/usr/bin/env bash

pushd ../../../
cat db/schema/20181024.sql | docker exec -i $(docker-compose ps -q db) /usr/bin/mysql -u root --password=a_root_db_password algorea_db
popd

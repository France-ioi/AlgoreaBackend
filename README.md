# Algorea Backend

## Requirements

This project requires Go >=1.11 with modules enabled (GO111MODULE=on). However, we recommend to develop into your $GOPATH (old style) as the linter does not support modules for now.

## Running the app

Compile the app:
```
make
```

You can then run the app: (call `./bin/AlgoreaBackend` to print the list of available commands)
```
./bin/AlgoreaBackend <cmd> <opt>
```
For instance, you can launch the web server using `./bin/AlgoreaBackend serve`.

## Running the setup

The application needs a database (MySQL) to run and requires it for a major part of its tests.

To make testing and development easier, a `docker-compose` file declares a database using the default configuration. Launch `docker-compose up` to run tests without any configuration efforts.

## Configuration

The app configuration stands in the `conf/config.yml` file. The file `conf/config.sample.yml` is a sample configuration to start from, it is configured to work with the `docker-compose` configuration for local development. All configuration parameter can be also defined using environment variables (with an higher priority), see `.circleci/config.yml` for examples.

## Seeding the database

An empty dump (schema without data) can be loaded using the `./bin/AlgoreaBackend db-restore` followed by `./bin/AlgoreaBackend db-migrate`.

## Testing

Run all tests (unit and bdd):
```
make test
```
Only unit:
```
make test-unit
```
Only bdd (cucumber using `godog`), using the database connection:
```
make test-bdd
```
or if you want only to run bdd tests with a specific tag:
```
make ARGS="--tags=wip" test-bdd
```

## Style

A `.editorconfig` file defines the basic editor style configuration to use. Check the "editorconfig" support for your favorite editor if it is not installed by default.

For the Go coding styles, we use the standard linters (many). You can install and run them with:
```
make lint
```

## Software Walkthrough

### Routing a request

* The web app is defined in `app.go` which loads all the middlewares and routes. The routing part consists in mounting the API on `/` and giving a context to it (i.e., the app database connection)
* The API routing (`app/api/api.go`) does the same for mounting all group of services.
* A service group (e.g., `app/api/groups/groups.go`.) mounts all its services and pass the context again.
* Each service has its dedicated file (e.g., `app/api/groups/get-all.go`). We try to separate the actual HTTP request parsing and response generation from the actual business logic and the call to the database.

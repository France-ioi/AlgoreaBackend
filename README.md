# Algorea Backend

## Running the app (for development)

Compile the app:
```
make
```

You can then run the app: (call `./bin/AlgoreaBackend` to print the list of available commands)
```
./bin/AlgoreaBackend <cmd> <opt>
```
For instance, you can launch the web server using `./bin/AlgoreaBackend serve`.

## Running the setup (as API consumer)

The easiest way to run the backend for consumer it is to run it in a container with its database. To do that:

* clone this repository (or download the `docker-compose.yml` file and replace `build: .` by `image: franceioi/algoreabackend:latest` to use the public image)
* Seed the database:

  ```docker-compose run backend /bin/sh -c "sleep 1; ALGOREA_DATABASE__USER="root"; ALGOREA_DATABASE__PASSWD="a_root_db_password"; AlgoreaBackend db-restore && AlgoreaBackend db-migrate && AlgoreaBackend install"```
* Launch the docker compose setup (db+backend): `docker-compose up`
* Visit `http://127.0.0.1:8080/status` with your browser, you should get a success status message.

If needed, you can connect on the MySQL CLI using:
```
docker exec -it algoreabackend_db_1 mysql -h localhost -u algorea -pa_db_password  --protocol=TCP algorea_db
```

## Running the setup (as dev)

The application needs a database (MySQL) to run and requires it for a major part of its tests.

It is required to set
  * `innodb_lock_wait_timeout`=5,
  * `innodb_ft_min_token_size`=1
  * `max-allowed-packet`=10485760

and provide at least 2Gb of memory to the MySQL server.

To make testing and development easier, a `docker-compose` file declares a database using the default configuration. Launch `docker-compose up` to run tests without any configuration efforts.

## Configuration

The app configuration stands in the `conf/config.yml` file. The file `conf/config.sample.yml` is a sample configuration to start from, it is configured to work with the `docker-compose` configuration for local development. All configuration parameter can be also defined using environment variables (with an higher priority), see `.circleci/config.yml` for examples.

Environment-specific configurations can be defined using `conf/config.ENV.yml` files when ENV can be "prod", "dev" or "test.

### Configuration of `test` environment

The `test` environment is used for running the tests.
For the `test` environment, we don't fall back to the default configuration file, so you need to provide a `conf/config.test.yml` file.
This is to avoid running tests on a production database by mistake and erasing data.

## Creating the keys

```
openssl genrsa --out private_key.pem 4096
openssl rsa -in private_key.pem -pubout -out public_key.pem
```


## Seeding the database

An empty dump (schema without data) can be loaded using the
```
./bin/AlgoreaBackend db-restore
```
followed by
```
./bin/AlgoreaBackend db-migrate
```
Probably you may also want to run
```
./bin/AlgoreaBackend install
```
to insert the data required by the config.

Also, after changing the DB data manually, you probably want to run
```
./bin/AlgoreaBackend db-recompute
```
in order to recompute DB caches.

## Testing

### make test
To execute all tests (unit and bdd) with race detection and collect the test code coverage you can run:
```
make test
```

This mode is the slowest one, it doesn't use the cache, and it always runs all tests. It is useful to run before pushing code to the repository.

### make test-dev

To get test results faster during development, you may want to run all tests without race detection and without collecting the test code coverage to get advantage of Golang per-package caching:
```
make test-dev
```

### make test-unit

You may want to run only unit tests that are marked with the "unit" tag:
```
make test-unit
```
`make test-unit` doesn't cache test results, all the matching tests will be run. For unit tests not marked as "unit", use `make test` or `make test-dev` instead.

### make test-bdd

It is possible to run only Gherkin tests defined in *.feature files:
```
make test-bdd
```
or if you want only to run bdd tests with a specific tag, in a specific directory. Specifying the directory is mandatory when using tags.

```
DIRECTORY=./app/api/answers/ TAGS=wip make test-bdd
```
To add a tag to a test, just precede it by @wip on the line above it in the *.feature file. This is useful to only execute appropriate tests.

`make test-bdd` doesn't cache test results, all the matching tests will be run.

### Tests filtering
For all `make test*` it is possible to filter with a certain directory and the name of the test function you want to run:
```
make DIRECTORY=./app/database FILTER=TestItemStore_TriggerBeforeInsert_SetsPlatformID test
```
Note that `FILTER` is not currently supported for `make test-bdd`.


## Install the git hooks

Copy `githooks/pre-commit` to `.git/hooks/pre-commit`. You may want to adapt the content in case you have personal hooks.

## Style

A `.editorconfig` file defines the basic editor style configuration to use. Check the "editorconfig" support for your favorite editor if it is not installed by default.

For the Go coding styles, we use the standard linters (many). You can install and run them with:
```
make lint
```

## Code formatting

We use `gofumpt` to format the code. It is stricter than `gofmt`.


## Generating the API documentation

The API documentation is an OpenAPI spec file generated automatically by [Go-Swagger](https://github.com/go-swagger/go-swagger) and [swagger2openapi](https://www.npmjs.com/package/swagger2openapi).
Each time the code is pushed in the "master" branch, the CI generates the spec file and deploys it to the documentation server.

To perform the spec generation locally, install our patched version of Go-Swagger from source (requires Go 1.21+):
```
go install github.com/France-ioi/go-swagger/cmd/swagger@00200fa
```

and swagger2openapi:
```
npm install -g swagger2openapi
```

Also, you need to install [redocly-cli](https://redocly.com/redocly-cli) to serve the documentation locally:
```
npm install -g @redocly/cli@1.25.15
```

After everything is installed, you can generate the specification file from code and validate it:
```
make swagger-generate
```

To view the documentation in a browser, start serving it:
```
make swagger-serve
```
and open http://127.0.0.1:8080 in your favorite browser.

## Create a release

In order to create a release:
- decide of a new version number (using semver)
- update the changelog (add a new section, with the date of today and listing the fix and new features)
- commit this change as a commit "Release vx.y.z"
- tag the current commit "vx.y.z" (`git tag -a -m "Release vx.y.z" vx.y.z`)
- push everything (`git push origin master; git push origin vx.y.z`)
- the rest (github release, doc generation and deployment) is done by the CI

## Software Walkthrough

### Routing a request

* The web app is defined in `app.go` which loads all the middlewares and routes. The routing part consists in mounting the API on `/` and giving a context to it (i.e., the app database connection)
* The API routing (`app/api/api.go`) does the same for mounting all group of services.
* A service group (e.g., `app/api/groups/groups.go`.) mounts all its services and pass the context again.
* Each service has its dedicated file (e.g., `app/api/groups/get-all.go`). We try to separate the actual HTTP request parsing and response generation from the actual business logic and the call to the database.

## How to profile a service
1. Start the server in 'dev' environment
```
./bin/AlgoreaBackend serve dev
```
```
2019/07/10 00:15:39 Loading environment: test
INFO 2019/07/10 00:15:39 Starting application: environment = dev
INFO 2019/07/10 00:15:39 Loading environment: dev
INFO 2019/07/10 00:15:39 Configuring server...
INFO 2019/07/10 00:15:39 Starting server...
INFO 2019/07/10 00:15:39 Listening on :8080
```

2. Start making many requests to the service you want to profile and wait for 10 seconds:
```
ab -k -c 1 -n 10000 -H "Authorization: Bearer 1" "http://127.0.0.1:8080/groups/5/team-descendants"
```
('-c 1' means 'concurrency = 1', you can try other values as well)

3. Get the profile:
```
go tool pprof http://127.0.0.1:8080/debug/pprof/profile?seconds=10
```
Type 'web' as a pprof command to see the call graph with durations.

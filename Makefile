GOCMD=env GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLIST=$(GOCMD) list
BIN_NAME=AlgoreaBackend

TEST_REPORT_DIR=test-results

LOCAL_BIN_DIR=./bin

ifndef BIN_DIR # to allow BIN_DIR to be given as args (see CI)
	FIRSTGOPATH=$(shell echo $(GOPATH) | cut -d: -f1 -)
	BIN_DIR=$(FIRSTGOPATH)/bin
endif
BIN_PATH=$(LOCAL_BIN_DIR)/$(BIN_NAME)
GODOG=$(BIN_DIR)/godog
GOLANGCILINT=$(LOCAL_BIN_DIR)/golangci-lint
MYSQL_CONNECTOR_JAVA=$(LOCAL_BIN_DIR)/mysql-connector-java-8.jar
SCHEMASPY=$(LOCAL_BIN_DIR)/schemaspy-6.0.0.jar

# extract AWS_PROFILE if given
ifdef AWS_PROFILE
	AWS_PARAMS=--profile $(AWS_PROFILE)
endif

# use the NOECHO env var to disable the echo (printing the executed commands) on make test
ifeq ("$(NOECHO)","1")
	Q := @
	vecho = @true
else
	Q :=
	vecho = @echo
endif

# Check that given variables are set and all have non-empty values,
# die with an error otherwise.
#
# Params:
#   1. Variable name(s) to test.
#   2. (optional) Error message to print.
check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))

all: build
build: $(BIN_PATH)
$(BIN_PATH): .FORCE # let go decide what need to be rebuilt
	$(GOBUILD) -o $(BIN_PATH) -v -tags=prod
gen-keys:
	openssl genpkey -algorithm RSA -out private_key.pem 2>/dev/null | openssl genrsa -out private_key.pem 1024
	openssl rsa -pubout -in private_key.pem -out public_key.pem
db-restore: $(BIN_PATH)
	$(BIN_PATH) db-restore
db-migrate: $(BIN_PATH)
	$(BIN_PATH) db-migrate
db-migrate-undo: $(BIN_PATH)
	$(BIN_PATH) db-migrate-undo
db-recompute: $(BIN_PATH)
	$(BIN_PATH) db-recompute

test: $(TEST_REPORT_DIR)
	$(Q)# the tests using the db do not currently support parallism
	$(Q)$(GOTEST) -gcflags=all=-l -race -coverprofile=$(TEST_REPORT_DIR)/coverage.txt -covermode=atomic -v ./app/... -p 1 -parallel 1
test-unit:
	$(GOTEST) -gcflags=all=-l -race -cover -v ./app/... -tags=unit
test-bdd: $(GODOG)
	# to pass args: make ARGS="--tags=wip" test-bdd
	$(GODOG) --format=progress $(ARGS) .
lint: $(GOLANGCILINT)
	$(GOLANGCILINT) run --deadline 10m0s

dbdoc: $(MYSQL_CONNECTOR_JAVA) $(SCHEMASPY)
	$(call check_defined, DBNAME)
	$(call check_defined, DBHOST)
	$(call check_defined, DBUSER)
	$(call check_defined, DBPASS)
	java -jar $(SCHEMASPY) -t mysql -dp $(MYSQL_CONNECTOR_JAVA) -db $(DBNAME) -host $(DBHOST) -port 3306 -u $(DBUSER) -p $(DBPASS) -o db/doc -s $(DBNAME) -noimplied -nopages
clean:
	$(GOCLEAN)
	$(GOCLEAN) -testcache
	rm -rf $(LOCAL_BIN_DIR)/*

lambda-build:
	GOOS=linux $(GOBUILD) -o $(BIN_PATH)-linux
lambda-archive: lambda-build
	zip -j $(LOCAL_BIN_DIR)/lambda.zip $(BIN_PATH)-linux
lambda-upload: lambda-archive
	# pass AWS profile with AWS_PROFILE: make AWS_PROFILE="myprofile" lambda-upload
	aws lambda update-function-code --function-name AlgoreaBackend --zip-file fileb://$(LOCAL_BIN_DIR)/lambda.zip $(AWS_PARAMS)

$(TEST_REPORT_DIR):
	mkdir -p $(TEST_REPORT_DIR)
$(GODOG):
	$(GOGET) -u github.com/cucumber/godog/cmd/godog@v0.9.0
$(GOLANGCILINT):
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(LOCAL_BIN_DIR) v1.18.0
$(MYSQL_CONNECTOR_JAVA):
	curl -sfL https://dev.mysql.com/get/Downloads/Connector-J/mysql-connector-java-8.0.16.tar.gz | tar -xzf - mysql-connector-java-8.0.16/mysql-connector-java-8.0.16.jar
	mv mysql-connector-java-8.0.16/mysql-connector-java-8.0.16.jar $(MYSQL_CONNECTOR_JAVA)
	rm -rf mysql-connector-java-8.0.16
$(SCHEMASPY):
	curl -sfL -o $(SCHEMASPY) https://github.com/schemaspy/schemaspy/releases/download/v6.0.0/schemaspy-6.0.0.jar

.FORCE: # force the rule using it to always re-run
.PHONY: all build test test-unit test-bdd lint clean lambda-build lambda-archive lambda-upload db-restore db-migrate gen-keys dbdoc

GOCMD=go
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

all: build
build: $(BIN_PATH)
$(BIN_PATH): .FORCE # let go decide what need to be rebuilt
	$(GOBUILD) -o $(BIN_PATH) -v -race
gen-keys:
	openssl genpkey -algorithm RSA -out private_key.pem 2>/dev/null | openssl genrsa -out private_key.pem 1024
	openssl rsa -pubout -in private_key.pem -out public_key.pem
db-restore: $(BIN_PATH)
	$(BIN_PATH) db-restore
db-migrate: $(BIN_PATH)
	$(BIN_PATH) db-migrate

test: $(TEST_REPORT_DIR)
	$(Q)# the tests using the db do not currently support parallism
	$(Q)$(GOTEST) -race -coverprofile=$(TEST_REPORT_DIR)/coverage.txt -covermode=atomic -v ./app/... -p 1 -parallel 1
test-unit:
	$(GOTEST) -race -cover -v ./app/... -tags=unit
test-bdd: $(GODOG)
	# to pass args: make ARGS="--tags=wip" test-bdd
	$(GODOG) --format=progress $(ARGS) .
lint: $(GOLANGCILINT)
	$(GOLANGCILINT) run --deadline 5m0s

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
	$(GOGET) -u github.com/DATA-DOG/godog/cmd/godog
$(GOLANGCILINT):
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(LOCAL_BIN_DIR) v1.15.0

.FORCE: # force the rule using it to always re-run
.PHONY: all build test test-unit test-bdd lint clean lambda-build lambda-archive lambda-upload db-restore db-migrate gen-keys

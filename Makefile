GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLIST=$(GOCMD) list
BINARY_NAME=AlgoreaBackend

TEST_REPORT_DIR=test-results

LOCAL_BIN_DIR=./bin

ifndef BIN_DIR # to allow BIN_DIR to be given as args (see CI)
	FIRSTGOPATH=$(shell echo $(GOPATH) | cut -d: -f1 -)
	BIN_DIR=$(FIRSTGOPATH)/bin
endif
GODOG=$(BIN_DIR)/godog
GOMETALINTER=$(LOCAL_BIN_DIR)/gometalinter

# extract AWS_PROFILE if given
ifdef AWS_PROFILE
	AWS_PARAMS=--profile $(AWS_PROFILE)
endif

# use the NOTVERBOSE env var to disable verbosity on make test
ifneq ("$(NOT_VERBOSE)","1")
	Q :=
	vecho = @echo
else
	Q := @
	vecho = @true
endif

.PHONY: all build test lint clean deps print-deps

all: build
build:
	$(GOBUILD) -o $(LOCAL_BIN_DIR)/$(BINARY_NAME) -v -race
test: $(TEST_REPORT_DIR)
	$(Q)# the tests using the db do not currently support parallism
	$(Q)$(GOTEST) -race -coverprofile=$(TEST_REPORT_DIR)/coverage.txt -covermode=atomic -v ./app/... -p 1 -parallel 1
test-unit:
	TESTS_NODB=1 $(GOTEST) -race -cover -v ./app/...
test-bdd: $(GODOG)
	# to pass args: make ARGS="--tags=wip" test-bdd
	$(GODOG) --format=progress $(ARGS) .
lint: $(GOMETALINTER)
	PATH=$(LOCAL_BIN_DIR):$(PATH) GO111MODULE=off $(GOMETALINTER) ./... --deadline=90s
clean:
	$(GOCLEAN)
	$(GOCLEAN) -testcache
	rm -rf $(LOCAL_BIN_DIR)/*
deps:
	GO111MODULE=off $(GOGET) -t ./...
print-deps:
	$(GOLIST) -f {{.Deps}} && $(GOLIST) -f {{.TestImports}} ./...
lambda-build:
	GOOS=linux $(GOBUILD) -o $(LOCAL_BIN_DIR)/$(BINARY_NAME)-linux
lambda-archive: lambda-build
	zip -j $(LOCAL_BIN_DIR)/lambda.zip $(LOCAL_BIN_DIR)/$(BINARY_NAME)-linux
lambda-upload: lambda-archive
	# pass AWS profile with AWS_PROFILE: make AWS_PROFILE="myprofile" lambda-upload
	aws lambda update-function-code --function-name AlgoreaBackend --zip-file fileb://$(LOCAL_BIN_DIR)/lambda.zip $(AWS_PARAMS)
$(TEST_REPORT_DIR):
	mkdir -p $(TEST_REPORT_DIR)
$(GODOG):
	$(GOGET) -u github.com/DATA-DOG/godog/cmd/godog
$(GOMETALINTER):
	curl -L https://git.io/vp6lP | sh /dev/stdin -b $(LOCAL_BIN_DIR)

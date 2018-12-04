GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLIST=$(GOCMD) list
BINARY_NAME=AlgoreaBackend

TEST_REPORT_DIR=test-results

ifndef BIN_DIR # to allow BIN_DIR to be given as args (see CI)
	FIRSTGOPATH=$(shell echo $(GOPATH) | cut -d: -f1 -)
	BIN_DIR=$(FIRSTGOPATH)/bin
endif
GODOG=$(BIN_DIR)/godog
GO_JUNIT_REPORT=$(BIN_DIR)/go-junit-report
GOMETALINTER=./bin/gometalinter

.PHONY: all build test lint clean deps print-deps

all: build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v
test-unit:
	$(GOTEST) -cover -v ./app/...
test-bdd: $(GODOG)
	# to pass args: make ARGS="--tags=wip" test-bdd
	$(GODOG) --format=progress $(ARGS)
test-unit-report: $(GO_JUNIT_REPORT)
	mkdir -p $(TEST_REPORT_DIR)/go-test
	$(GOTEST) -cover -v ./app/... 2>&1 | $(GO_JUNIT_REPORT) > $(TEST_REPORT_DIR)/go-test/junit.xml
test-bdd-report: $(GODOG)
	mkdir -p $(TEST_REPORT_DIR)/cucumber
	$(GODOG) --format=junit > $(TEST_REPORT_DIR)/cucumber/junit.xml
test: test-unit test-bdd
lint: $(GOMETALINTER)
	PATH=./bin:$(PATH) $(GOMETALINTER) ./... --deadline=90s
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
deps: $(GODOG) $(GO_JUNIT_REPORT)
	$(GOGET) -t ./...
print-deps:
	$(GOLIST) -f {{.Deps}} && $(GOLIST) -f {{.TestImports}} ./...
$(GODOG):
	$(GOGET) -u github.com/DATA-DOG/godog/cmd/godog
$(GO_JUNIT_REPORT):
	$(GOGET) -u github.com/jstemmer/go-junit-report
$(GOMETALINTER):
	curl -L https://git.io/vp6lP | sh

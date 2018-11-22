MAINGOPATH=$(shell echo $(GOPATH) | cut -d: -f1 -)
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLIST=$(GOCMD) list
BIN_DIR=$(MAINGOPATH)/bin
BINARY_NAME=AlgoreaBackend

TEST_REPORT_DIR=test-results

GODOG=$(BIN_DIR)/godog
GO_JUNIT_REPORT=$(BIN_DIR)/go-junit-report
GOMETALINTER=$(BIN_DIR)/gometalinter

.PHONY: all build test lint clean deps print-deps

all: build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v
test-unit:
	$(GOTEST) -v ./tests/unit/...
test-bdd: $(GODOG)
	# to pass args: make ARGS="--tags=wip" test-bdd
	(cd tests/bdd && $(GODOG) --format=progress $(ARGS))
test-unit-report: $(GO_JUNIT_REPORT)
	mkdir -p $(TEST_REPORT_DIR)/go-test
	$(GOTEST) -v ./tests/unit/... 2>&1 | $(GO_JUNIT_REPORT) > $(TEST_REPORT_DIR)/go-test/junit.xml
test-bdd-report: $(GODOG)
	mkdir -p $(TEST_REPORT_DIR)/cucumber
	(cd tests/bdd && $(GODOG) --format=junit) > $(TEST_REPORT_DIR)/cucumber/junit.xml
test: test-unit test-bdd
lint: $(GOMETALINTER)
	$(GOMETALINTER) ./... --deadline=90s
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
deps: $(GODOG) $(GO_JUNIT_REPORT)
	$(GOGET) -t ./...
print-deps:
	$(GOLIST) -f {{.Deps}}
$(GODOG):
	$(GOGET) -u github.com/DATA-DOG/godog/cmd/godog
$(GO_JUNIT_REPORT):
	$(GOGET) -u github.com/jstemmer/go-junit-report
$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install &> /dev/null

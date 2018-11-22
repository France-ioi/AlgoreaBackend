MAINGOPATH=$(shell echo $(GOPATH) | cut -d: -f1 -)
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLIST=$(GOCMD) list
BIN_DIR=$(MAINGOPATH)/bin
BINARY_NAME=AlgoreaBackend

GODOG=$(BIN_DIR)/godog

.PHONY: all build test lint clean deps print-deps

all: build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v
test-unit:
	$(GOTEST) -v ./tests/unit/...
test-bdd: $(GODOG)
	# to pass args: make ARGS="--tags=wip" test-bdd
	(cd tests/bdd && $(GODOG) --format=progress $(ARGS))
test: test-unit test-bdd
lint:
	gometalinter ./... --vendor
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
deps: $(GODOG)
	$(GOGET) -t ./...
print-deps:
	$(GOLIST) -f {{.Deps}}
$(GODOG):
	$(GOGET) -u github.com/DATA-DOG/godog/cmd/godog
$(BIN_DIR)/gometalinter:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install &> /dev/null

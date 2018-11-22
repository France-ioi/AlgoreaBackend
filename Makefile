MAINGOPATH=$(shell echo $(GOPATH) | cut -d: -f1 -)
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLIST=$(GOCMD) list
BIN_DIR=$(MAINGOPATH)/bin
GOMETALINTER=$(BIN_DIR)/gometalinter
BINARY_NAME=AlgoreaBackend

.PHONY: all build test lint clean deps print-deps

all: build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v
test-unit:
	$(GOTEST) -v ./tests/unit/...
test-bdd:
	# to pass args: make ARGS="--godog.tags=wip" test-bdd
	$(GOTEST) -v ./tests/bdd/... $(ARGS)
test: test-unit test-bdd
lint:
	gometalinter ./... --vendor
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
deps:
	$(GOGET) -t ./...
print-deps:
	$(GOLIST) -f {{.Deps}}
$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install &> /dev/null

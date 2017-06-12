LDFLAGS += -X "github.com/kyleterry/sufr/config.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
LDFLAGS += -X "github.com/kyleterry/sufr/config.BuildGitHash=$(shell git rev-parse HEAD)"
PREFIX?=/usr/local
INSTALL_BIN=$(PREFIX)/bin/
BIN_OUT=sufr

all: build

build: vendor-get generate
	go build -o $(BIN_OUT) -v -ldflags '$(LDFLAGS)'

clean:
	-rm $(BIN_OUT)

cross-compile:
	go get github.com/mitchellh/gox
	gox -ldflags '$(LDFLAGS)'

generate:
	go generate

install:
	@cp $(BIN_OUT) $(INSTALL_BIN)sufr

vendor-get:
	go get -u -v github.com/jteeuwen/go-bindata/...

test:
	go test -v $(shell go list ./... | grep -v vendor)

.PHONY: all clean build cross-compile generate install vendor-get

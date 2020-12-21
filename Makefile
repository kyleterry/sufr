LDFLAGS += -X "github.com/kyleterry/sufr/config.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
LDFLAGS += -X "github.com/kyleterry/sufr/config.BuildGitHash=$(shell git rev-parse HEAD)"
LDFLAGS += -X "github.com/kyleterry/sufr/config.Version=$(shell git describe --tags)"
PREFIX?=/usr/local
INSTALL_BIN=$(PREFIX)/bin/
BIN_OUT=bin/sufr

all: build

build: generate
	go build -o $(BIN_OUT) -v -ldflags '$(LDFLAGS)' ./cmd/sufr

clean:
	-rm $(BIN_OUT)

cross-compile:
	go get github.com/mitchellh/gox
	gox -ldflags '$(LDFLAGS)'

generate:
	go generate ./...

install:
	@cp $(BIN_OUT) $(INSTALL_BIN)sufr

test:
	go test -v ./...

.PHONY: all clean build cross-compile generate install

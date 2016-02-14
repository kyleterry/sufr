LDFLAGS += -X "github.com/kyleterry/sufr/config.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
LDFLAGS += -X "github.com/kyleterry/sufr/config.BuildGitHash=$(shell git rev-parse HEAD)"
PREFIX?=/usr/local
INSTALL_BIN=$(PREFIX)/bin/
BIN_OUT=sufr

all: build

build: vendor-get generate
	go build -o $(BIN_OUT) -v -ldflags '$(LDFLAGS)'

generate:
	go generate

install:
	@cp $(BIN_OUT) $(INSTALL_BIN)sufr

vendor-get:
	go get github.com/jteeuwen/go-bindata
	go get github.com/mjibson/esc

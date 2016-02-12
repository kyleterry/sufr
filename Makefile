LDFLAGS += -X "github.com/kyleterry/sufr/config.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
LDFLAGS += -X "github.com/kyleterry/sufr/config.BuildGitHash=$(shell git rev-parse HEAD)"
all: build

build: generate
	go build -v -ldflags '$(LDFLAGS)'

generate:
	go generate

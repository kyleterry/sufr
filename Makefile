all: build

generate:
	go generate

build: generate
	go build

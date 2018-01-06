BINARY=furnace-gcp

.DEFAULT_GOAL := build

.PHONY: clean build test linux


build:
	go build -ldflags="-s -w" -i -o ${BINARY}

osx:
	go build -i -o ${BINARY}-osx

test:
	go test ./...

get-deps:
	dep ensure

clean:
	if [ -f ${BINARY} ]; then rm ${BINARY}; fi

linux:
	env GOOS=linux GOARCH=arm go build -o ${BINARY}-linux

windows:
	env GOOS=windows GOARCH=386 go build -o ${BINARY}-windows.exe

all: osx linux windows

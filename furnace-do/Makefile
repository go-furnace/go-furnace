BINARY=furnace-do

.DEFAULT_GOAL := build

.PHONY: clean build test linux


build:
	go build -ldflags="-s -w" -i -o ../cmd/${BINARY}

osx:
	go build -ldflags="-s -w" -i -o ../cmd/${BINARY}-osx

test:
	go test ./...

get-deps:
	vgo build

linux:
	env GOOS=linux GOARCH=arm go build -o ../cmd/${BINARY}-linux

windows:
	env GOOS=windows GOARCH=386 go build -o ../cmd/${BINARY}-windows.exe

clean:
	go clean -i

all: osx linux windows

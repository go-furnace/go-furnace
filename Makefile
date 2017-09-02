BINARY=furnace

.DEFAULT_GOAL := build

.PHONY: clean build test linux

get-deps:
 	go get github.com/aws/aws-sdk-go
 	go get github.com/Yitsushi/go-commander
 	go get github.com/fatih/color


build:
	go build -i -o ${BINARY}

osx:
	go build -i -o ${BINARY}-osx

test:
	go test ./...

install:
	go install

clean:
	if [ -f ${BINARY} ]; then rm ${BINARY}; fi

linux:
	env GOOS=linux GOARCH=arm go build -o ${BINARY}-linux

windows:
	env GOOS=windows GOARCH=386 go build -o ${BINARY}-windows.exe

all: osx linux windows

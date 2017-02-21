BINARY=furnace

.DEFAULT_GOAL := build

build:
	go build -o ${BINARY}

test:
	go test -v ./...

get-deps:
	go get github.com/aws/aws-sdk-go
	go get github.com/Yitsushi/go-commander
	go get github.com/fatih/color

clean:
	if [ -f ${BINARY} ]; then rm ${BINARY}; fi

linux:
	env GOOS=linux GOARCH=arm go build -o ${BINARY}

.PHONY: clean build test linux

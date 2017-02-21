BINARY=furnace

.DEFAULT_GOAL := build

linux:
	env GOOS=linux GOARCH=arm go build -o ${BINARY}

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

.PHONY: clean build
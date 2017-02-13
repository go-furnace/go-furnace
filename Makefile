BINARY=furnace

.DEFAULT_GOAL := build

linux:
	env GOOS=linux GOARCH=arm go build -o ${BINARY}

build:
	go build -o ${BINARY}

clean:
	if [ -f ${BINARY} ]; then rm ${BINARY}; fi

.PHONY: clean build
.DEFAULT_GOAL := build-all

build-all:
	make -C aws && make -C gcp

test:
	go test ./...

get-deps-all:
	make dep ensure -C aws && make dep ensure -C gcp

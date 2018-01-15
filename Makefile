.DEFAULT_GOAL := build-all

build-all:
	make -C aws && make -C gcp

test:
	go test ./...

get-deps-all:
	make get-deps -C aws && make get-deps -C gcp


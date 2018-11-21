.DEFAULT_GOAL := build-all

build-all:
	make -C furnace-aws && make -C furnace-gcp

test:
	go test ./...

get-deps-all:
	go get ./...

install-all:
	go install ./...

clean-all:
	make clean -C furnace-aws && make clean -C furnace-gcp

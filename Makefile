.DEFAULT_GOAL := build-all

build-all:
	make -C furnace-aws && make -C furnace-gcp

test:
	go test ./...

get-deps-all:
	make get-deps -C furnace-aws && make get-deps -C furnace-gcp && make get-deps -C furnace-do

install-all:
	go install ./...

clean-all:
	make clean -C furnace-aws && make clean -C furnace-gcp

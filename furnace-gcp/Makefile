NAME=furnace-gcp

# Set the build dir, where built cross-compiled binaries will be output
BUILDDIR := ../cmd

# List the GOOS and GOARCH to build
GOOSARCHES = darwin/amd64 darwin/386 freebsd/amd64 freebsd/386 linux/arm linux/arm64 linux/amd64 linux/386 solaris/amd64 windows/amd64 windows/386
GO_LDFLAGS_STATIC=-ldflags "-w $(CTIMEVAR) -extldflags -static"

.DEFAULT_GOAL := build

.PHONY: build
build:
	go build -ldflags="-s -w" -i -o ../cmd/${NAME}

.PHONY: test
test:
	go test ./...

.PHONY: get-deps
get-deps:
	dep ensure

.PHONY: clean
clean:
	go clean -i

define buildpretty
mkdir -p $(BUILDDIR)/$(1)/$(2);
GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 go build \
	 -o $(BUILDDIR)/$(1)/$(2)/$(NAME) \
	 -a -tags "$(BUILDTAGS)" \
	 -installsuffix netgo ${GO_LDFLAGS_STATIC} .;
shasum -a 256 $(BUILDDIR)/$(1)/$(2)/$(NAME) > $(BUILDDIR)/$(1)/$(2)/$(NAME).sha256;
endef

.PHONY: cross
cross: *.go ## Builds the cross-compiled binaries, creating a clean directory structure (eg. GOOS/GOARCH/binary)
	@echo "+ $@"
	$(foreach GOOSARCH,$(GOOSARCHES), $(call buildpretty,$(subst /,,$(dir $(GOOSARCH))),$(notdir $(GOOSARCH))))

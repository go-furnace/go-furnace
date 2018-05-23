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
	vgo get -u

.PHONY: clean
clean:
	go clean -i

define buildrelease
GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 go build \
	 -o $(BUILDDIR)/$(NAME)-$(1)-$(2) \
	 -a -tags "$(BUILDTAGS)" \
	 -installsuffix netgo ${GO_LDFLAGS_STATIC} .;
shasum -a 256 $(BUILDDIR)/$(NAME)-$(1)-$(2) > $(BUILDDIR)/$(NAME)-$(1)-$(2).sha256;
endef

.PHONY: release
release: *.go ## Builds the cross-compiled binaries, naming them in such a way for release (eg. binary-GOOS-GOARCH)
	@echo "+ $@"
	$(foreach GOOSARCH,$(GOOSARCHES), $(call buildrelease,$(subst /,,$(dir $(GOOSARCH))),$(notdir $(GOOSARCH))))

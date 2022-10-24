# +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
# +++ Check bash installed

DEPTEST=$(shell command -v bash 2> /dev/null)
ifeq ($(DEPTEST),)
$(error "Install bash to make it work")
endif


# +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++


BUILDDIR                 := $(CURDIR)/build


# It's necessary to set this because some environments don't link sh -> bash.
SHELL                             = /usr/bin/env bash

GOPATH                            = $(shell go env GOPATH)
ARCH                              = $(shell uname -p)

GIT_COMMIT                        = $(shell git rev-parse HEAD)
GIT_SHA                           = $(shell git rev-parse --short HEAD)
GIT_TAG                           = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY                         = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")


# --------------------------------------------------------------------------------
# --------------------------------------------------------------------------------
# --------------------------------------------------------------------------------

.PHONY: all
all: build

.PHONY: build
build: build-nsjail build-toastainer build-test

.PHONY: build-toastainer
build-toastainer: $(BUILDDIR)
	@cd $(CURDIR)/cmd/toastainer && go build && mv -f toastainer $(BUILDDIR)

.PHONY: build-test
build-test: $(BUILDDIR) build-toastainer
	@cd $(CURDIR)/cmd/toastest && go build && mv -f toastest $(BUILDDIR)

# A static build may be necessary if you compile with a GLIB_C version higher than the one on the target system
.PHONY: build-toastainer-static
build-toastainer-static: $(BUILDDIR)
	@cd $(CURDIR)/cmd/toastainer && go build -ldflags '-extldflags "-fno-PIC -static"' -buildmode pie -tags 'osusergo netgo static_build' && mv -f toastainer $(BUILDDIR)

# A static build may be necessary if you compile with a GLIB_C version higher than the one on the target system
.PHONY: build-test-static
build-test-static: $(BUILDDIR) build-toastainer-static
	@cd $(CURDIR)/cmd/toastest && go build -ldflags '-extldflags "-fno-PIC -static"' -buildmode pie -tags 'osusergo netgo static_build' && mv -f toastest $(BUILDDIR)

.PHONY: build-nsjail
build-nsjail: $(BUILDDIR)
	@$(CURDIR)/makescripts/nsjail.sh $(BUILDDIR) $(CURDIR)

.PHONY: gen-config-example
gen-config-example: $(BUILDDIR) build-toastainer
	@$(BUILDDIR)/toastainer configexpl -p $(BUILDDIR)/config_example.json

.PHONY: setup-network
setup-network:
	@sudo $(CURDIR)/makescripts/setupnet.sh

$(BUILDDIR):
	@mkdir -p $(BUILDDIR)
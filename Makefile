# +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
# +++ Check bash installed

DEPTEST=$(shell command -v bash 2> /dev/null)
ifeq ($(DEPTEST),)
$(error "Install bash to make it work")
endif

CONFIG_FILE                       ?= Makefile.config

# +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
# +++ Config

ifeq ($(wildcard $(CONFIG_FILE)),)
$(error config file $(CONFIG_FILE) not found.)
endif

include $(CONFIG_FILE)


# +++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++


BUILDDIR                 := $(CURDIR)/build


# It's necessary to set this because some environments don't link sh -> bash.
SHELL                             = /usr/bin/env bash

GOPATH                            = $(shell go env GOPATH)
GOBIN                             = $(shell which go)
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
build: $(BUILDDIR)/nsjail $(BUILDDIR)/toastfront build-toastainer build-test

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

.PHONY: build-toastfront
build-toastfront:
	@$(CURDIR)/makescripts/setuptoastfront.sh $(BUILDDIR)

.PHONY: serve-web
serve-web: $(BUILDDIR)/toastfront
	@cd $(CURDIR)/test/servedashboard && sudo $(GOBIN) test -v -timeout 99999s

.PHONY: build-web
build-web: $(BUILDDIR)/toastfront
	@cd $(CURDIR)/web && TOASTAINER_API_HOST=$(TOASTAINER_API_HOST) ../build/toastfront build && rm -rf $(BUILDDIR)/web && mv build $(BUILDDIR)/web

.PHONY: gen-config-example
gen-config-example: $(BUILDDIR) build-toastainer
	@$(BUILDDIR)/toastainer configexpl -p $(BUILDDIR)/config_example.json

.PHONY: setup-network
setup-network:
	@sudo $(CURDIR)/makescripts/setupnet.sh

$(BUILDDIR):
	@mkdir -p $(BUILDDIR)

$(BUILDDIR)/toastfront:
	@$(CURDIR)/makescripts/setuptoastfront.sh $(BUILDDIR)

$(BUILDDIR)/nsjail:
	@$(CURDIR)/makescripts/nsjail.sh $(BUILDDIR) $(CURDIR)
PKG := github.com/xuxinlai2002/creda-celo-balance

GOBUILD := go build -v
GOINSTALL := go install -v
GOTEST := go test

COMMIT := $(shell git describe --tags --dirty)
COMMIT_HASH := $(shell git rev-parse HEAD)
GOVERSION := $(shell go version | awk '{print $$3}')
DEV_TAGS := $(if ${tags},$(DEV_TAGS) ${tags},$(DEV_TAGS))

# We only return the part inside the double quote here to avoid escape issues
# when calling the external release script. The second parameter can be used to
# add additional ldflags if needed (currently only used for the release).
make_ldflags = $(2) -X $(PKG)/build.Commit=$(COMMIT) \
	-X $(PKG)/build.CommitHash=$(COMMIT_HASH) \
	-X $(PKG)/build.GoVersion=$(GOVERSION) \
	-X $(PKG)/build.RawTags=$(shell echo $(1) | sed -e 's/ /,/g')

DEV_LDFLAGS := -ldflags "$(call make_ldflags, $(DEV_TAGS))"

all: build

build:
	@$(call print, "Building...")
	$(GOBUILD) -tags="$(DEV_TAGS)" -o main $(DEV_LDFLAGS) $(PKG)

.PHONY: all build
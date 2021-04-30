#!/bin/env make


PROGRAM = nyan

ARCH = linux_386 linux_amd64 linux_arm linux_arm64
ARCH += darwin_amd64 darwin_arm64
ARCH += windows_386.exe windows_amd64.exe

TARGETS := $(foreach arch,$(ARCH),$(addprefix $(PROGRAM)_,$(arch)))

all: $(TARGETS)


get_os   = $(word 2,$(subst _, ,$(word 1,$(subst ., ,$@))))
get_arch = $(word 3,$(subst _, ,$(word 1,$(subst ., ,$@))))

$(TARGETS): %:
	@mkdir -p build
	CGO_ENABLED=0 GOOS=$(get_os) GOARCH=$(get_arch) go build -ldflags="-s -w" -o build/$@

GOOS             ?= linux
GOARCH           ?= amd64
CGO_ENABLED      ?= 1
OUTPUT_DIR	     ?= ./target
OUTPUT_NAME      ?= budich-cli
OUTPUT_EXT       ?=
OUTPUT_VERSION   ?= $(shell go test -v -count=1 -test.run='^TestShowCurrentVersion$$' ./metadata | grep METADATA_CURRENT_VERSION | cut -d' ' -f2)
OUTPUT_FULL_PATH  = $(OUTPUT_DIR)/bin/$(OUTPUT_NAME)_$(OUTPUT_VERSION)_$(GOOS)_$(GOARCH)$(OUTPUT_EXT)
GPG_KEY          ?= 0x59BFB401A134CAE1

ifdef GIT_COMMIT_HASH
	LDFLAGS += -X io.github.binatory/budich-cli/metadata.versionBuild=$(GIT_COMMIT_HASH)
endif

.PHONY: clean

all:

test:
	go test -count=1 -cover ./...

test-it:
	go test -count=1 -cover -tags=it ./...

.check_version:
	test $(OUTPUT_VERSION) # require $$OUTPUT_VERSION

build: .check_version $(OUTPUT_FULL_PATH) $(OUTPUT_FULL_PATH).asc

$(OUTPUT_FULL_PATH): $(OUTPUT_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build -ldflags "$(LDFLAGS)" -o "$@" .

$(OUTPUT_FULL_PATH).asc: GPG := gpg --batch --yes --pinentry-mode loopback
$(OUTPUT_FULL_PATH).asc: $(OUTPUT_FULL_PATH)
	echo "$$PK" | $(GPG) --import
	echo "$$PP" | $(GPG) --passphrase-fd 0 -u "$(GPG_KEY)" --detach-sign --armor -o "$@" $<

build-linux-amd64: install_alsa_lib
	$(MAKE) build GOOS=linux GOARCH=amd64 CGO_ENABLED=1

build-macos-amd64:
	$(MAKE) build GOOS=darwin GOARCH=amd64 CGO_ENABLED=1

build-windows-amd64:
	$(MAKE) build GOOS=windows GOARCH=amd64 CGO_ENABLED=0 OUTPUT_EXT=.exe

ALSA_VERSION := 1.2.6.1
install_alsa_lib: $(OUTPUT_DIR)/alsa-lib-$(ALSA_VERSION)_installed

$(OUTPUT_DIR)/alsa-lib-$(ALSA_VERSION): $(OUTPUT_DIR)
	wget -qO- "http://www.alsa-project.org/files/pub/lib/alsa-lib-$(ALSA_VERSION).tar.bz2" | tar -C "$(OUTPUT_DIR)" -xjvf-
	touch "$@"

$(OUTPUT_DIR)/alsa-lib-$(ALSA_VERSION)_installed: $(OUTPUT_DIR)/alsa-lib-$(ALSA_VERSION)
	cd "$(OUTPUT_DIR)/alsa-lib-$(ALSA_VERSION)"; \
		./configure && make install
	touch "$@"

$(OUTPUT_DIR): .check_version
	@mkdir -p "$(OUTPUT_DIR)"

clean:
	rm -rfv "$(OUTPUT_DIR)"
	go clean -testcache

LDFLAGS += -s -w
ifdef GIT_COMMIT_HASH
	LDFLAGS += -X io.github.binatory/budich-cli/metadata.versionBuild=$(GIT_COMMIT_HASH)
endif

OUTPUT_DIR              ?= ./target
OUTPUT_BIN_DIR           = $(OUTPUT_DIR)/bin
OUTPUT_NAME             ?= budich-cli
OUTPUT_VERSION           = $(shell go test -ldflags "$(LDFLAGS)" -v -count=1 -test.run='^TestShowCurrentVersion$$' ./metadata | grep METADATA_CURRENT_VERSION | cut -d' ' -f2)
OUTPUT_FULL_NAME_PREFIX  = $(OUTPUT_NAME)_$(OUTPUT_VERSION)_
OUTPUT_FULL_NAME         = $(OUTPUT_FULL_NAME_PREFIX)$(GOOS)_$(GOARCH)$(OUTPUT_EXT)
OUTPUT_FULL_PATH         = $(OUTPUT_BIN_DIR)/$(OUTPUT_FULL_NAME)
OUTPUT_FULL_PATH_PREFIX  = $(OUTPUT_BIN_DIR)/$(OUTPUT_FULL_NAME_PREFIX)
GPG_SIGN                ?= 0
GPG_KEY                 ?= 0x59BFB401A134CAE1
GNUPGHOME               ?= ./.gnupg

.PHONY: clean

all:

clean:
	rm -rfv "$(OUTPUT_DIR)"
	go clean -testcache

test:
	go test -ldflags "$(LDFLAGS)" -count=1 -cover ./...

test-it:
	go test -ldflags "$(LDFLAGS)" -count=1 -cover -tags=it ./...

GPG := gpg --batch --yes --pinentry-mode loopback
$(OUTPUT_BIN_DIR)/%:
	$(eval suffix  = $(subst $(OUTPUT_FULL_NAME_PREFIX),,$*))
	$(eval params = $(subst _, ,$(suffix)))
	$(eval os     = $(word 1,$(params)))
	$(eval arch   = $(basename $(word 2,$(params))))
	$(eval cgo    = $(shell test "$(os)" = "windows" && echo 0 || echo 1))
	GOOS=$(os) GOARCH=$(arch) CGO_ENABLED=$(cgo) go build -ldflags "$(LDFLAGS)" -o "$@" .
	upx --fast "$@"

ifeq ($(GPG_SIGN),1)
ifdef GPG_PRIVATE_KEY
	echo "$$GPG_PRIVATE_KEY" | $(GPG) --import
endif
	echo "$$GPG_PASSPHRASE" | $(GPG) --passphrase-fd 0 -u "$(GPG_KEY)" --detach-sign --armor "$@"
endif

build-darwin-amd64: $(OUTPUT_FULL_PATH_PREFIX)darwin_amd64
build-darwin-arm64: $(OUTPUT_FULL_PATH_PREFIX)darwin_arm64
build-windows-amd64: $(OUTPUT_FULL_PATH_PREFIX)windows_amd64.exe
build-linux-amd64: install_alsa_lib $(OUTPUT_FULL_PATH_PREFIX)linux_amd64

$(OUTPUT_DIR):
	@mkdir -p "$(OUTPUT_DIR)"

ALSA_VERSION := 1.2.6.1
install_alsa_lib: $(OUTPUT_DIR)/alsa-lib-$(ALSA_VERSION)_installed

$(OUTPUT_DIR)/alsa-lib-$(ALSA_VERSION): $(OUTPUT_DIR)
	wget -qO- "http://www.alsa-project.org/files/pub/lib/alsa-lib-$(ALSA_VERSION).tar.bz2" | tar -C "$(OUTPUT_DIR)" -xjvf-
	touch "$@"

$(OUTPUT_DIR)/alsa-lib-$(ALSA_VERSION)_installed: $(OUTPUT_DIR)/alsa-lib-$(ALSA_VERSION)
	cd "$(OUTPUT_DIR)/alsa-lib-$(ALSA_VERSION)"; \
		./configure && make install
	touch "$@"

APP_NAME := ailops
BUILD_DIR := dist
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Default target
.PHONY: all
all: build

# Build for current platform (native)
.PHONY: build
build:
	mkdir -p $(BUILD_DIR)
	@if [ -n "$(GOOS)" ] && [ -n "$(GOARCH)" ]; then \
		OUT=$(BUILD_DIR)/$(APP_NAME); \
		if [ "$(GOOS)" = "windows" ]; then OUT=$$OUT.exe; fi; \
		echo "Cross-compiling for $(GOOS)/$(GOARCH)"; \
		GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $$OUT .; \
	else \
		go build -o $(BUILD_DIR)/$(APP_NAME) .; \
	fi

# Build for a specific OS/Arch (e.g. make build-cross GOOS=linux GOARCH=arm64)
.PHONY: build-cross
build-cross:
	@echo "Building for $(GOOS)/$(GOARCH)"
	@if [ -z "$(GOOS)" ] || [ -z "$(GOARCH)" ]; then \
		echo "GOOS and GOARCH must be set (e.g. make build-cross GOOS=linux GOARCH=amd64)"; exit 1; \
	fi
	mkdir -p $(BUILD_DIR)
	OUT=$(BUILD_DIR)/$(APP_NAME)-$(GOOS)-$(GOARCH); \
	if [ "$(GOOS)" = "windows" ]; then OUT=$$OUT.exe; fi; \
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $$OUT .

# Build all platforms (for local testing/packaging)
.PHONY: build-all
build-all:
	@for PLATFORM in $(PLATFORMS); do \
		GOOS=$${PLATFORM%/*}; GOARCH=$${PLATFORM#*/}; \
		OUT=$(BUILD_DIR)/$(APP_NAME)-$$GOOS-$$GOARCH; \
		if [ "$$GOOS" = "windows" ]; then OUT=$$OUT.exe; fi; \
		echo "Building $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build -o $$OUT .; \
	done

# Install locally (current platform)
.PHONY: install
install: build
	cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
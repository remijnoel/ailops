APP_NAME := ailops
BUILD_DIR := build

.PHONY: all build clean

all: build

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) .

build-darwin-arm64:
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 .

build-linux-arm64:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 .

build-linux-amd64:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .

install: build
	cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/

clean:
	rm -rf $(BUILD_DIR)
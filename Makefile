BINARY_NAME := server
MAIN_PACKAGE := ./cmd/server

.PHONY: all build run clean help

all: build

## build: Build the server binaries for Mac and Windows
build:
	@echo "Building for Mac..."
	GOOS=darwin go build -o app-mac $(MAIN_PACKAGE)
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -o app-windows.exe $(MAIN_PACKAGE)

## run: Run the server directly
run:
	@echo "Running the server..."
	go run $(MAIN_PACKAGE)

## clean: Remove build artifacts
clean:
	@echo "Cleaning up..."
	rm -f app-mac app-windows.exe $(BINARY_NAME)

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' $(MAKEFILE_LIST) | sed -e 's/## //g' | column -t -s ':'
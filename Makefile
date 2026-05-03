BINARY_NAME := server
MAIN_PACKAGE := ./cmd/server

.PHONY: all build run clean help

all: build

## build: Build the server binary
build:
	@echo "Building the server..."
	go build -o $(BINARY_NAME) $(MAIN_PACKAGE)

## run: Run the server directly
run:
	@echo "Running the server..."
	go run $(MAIN_PACKAGE)

## clean: Remove build artifacts
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' $(MAKEFILE_LIST) | sed -e 's/## //g' | column -t -s ':'
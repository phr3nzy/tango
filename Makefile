# Description: Makefile for the project

# Variables
BINARY_NAME=tango
VERSION=0.1.0
BUILD=`date +%FT%T%z`

# Build for all platforms
build-all: build-linux build-windows build-macos

# Build the project
build:
	@echo "Building the project"
	@go build -o "$(BINARY_NAME)" -ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)" ./...

# Build for Linux optimized
build-linux:
	@echo "Building the project for Linux"
	@GOOS=linux GOARCH=amd64 go build -o "$(BINARY_NAME)-linux" -ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)" ./...

# Build for Windows optimized
build-windows:
	@echo "Building the project for Windows"
	@GOOS=windows GOARCH=amd64 go build -o "$(BINARY_NAME)-windows".exe -ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)" ./...

# Build for MacOS optimized
build-macos:
	@echo "Building the project for MacOS"
	@GOOS=darwin GOARCH=amd64 go build -o "$(BINARY_NAME)-macos" -ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)" ./...

# Run the project
run:
	@echo "Running the project"
	@go run main.go

# Clean the project
clean:
	@echo "Cleaning the project"
	@go clean
	@rm -f $(BINARY_NAME)

# Test the project
test:
	@echo "Running tests"
	@go test ./...

# Test the project with coverage
test-coverage:
	@echo "Running coverage tests"
	@go test -cover ./...

# Test the project with coverage and generate HTML report
test-html-coverage:
	@echo "Running coverage tests"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out
	@rm coverage.out

# Benchmark the project
benchmark:
	@echo "Running benchmarks"
	@go test -bench ./...

# Benchmark the project with memory profiling
benchmark-mem:
	@echo "Running memory benchmarks"
	@go test -bench . -benchmem

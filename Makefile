.PHONY: build run clean test deps

# Build the application
build:
	@echo "Building notebooklm-audio-generator..."
	cd scripts && go build -o ../bin/notebooklm-audio-generator add_to_notebooklm.go

# Run the application
run: build
	@echo "Running notebooklm-audio-generator..."
	./bin/notebooklm-audio-generator

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	go vet ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  build  - Build the application"
	@echo "  run    - Build and run the application"
	@echo "  deps   - Install dependencies"
	@echo "  clean  - Remove build artifacts"
	@echo "  test   - Run tests"
	@echo "  fmt    - Format code"
	@echo "  lint   - Run linter"
	@echo "  help   - Show this help message"

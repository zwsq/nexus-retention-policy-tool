.PHONY: build run test clean install deps

# Build the application
build:
	go build -o nexus-retention-policy ./cmd

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o nexus-retention-policy-linux-amd64 ./cmd
	GOOS=darwin GOARCH=amd64 go build -o nexus-retention-policy-darwin-amd64 ./cmd
	GOOS=windows GOARCH=amd64 go build -o nexus-retention-policy-windows-amd64.exe ./cmd

# Run the application
run:
	go run ./cmd -config config.yaml

# Run in dry-run mode
dry-run:
	@echo "Running in dry-run mode..."
	@sed -i.bak 's/dry_run: false/dry_run: true/' config.yaml || true
	go run ./cmd -config config.yaml
	@mv config.yaml.bak config.yaml || true

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f nexus-retention-policy
	rm -f nexus-retention-policy-*
	rm -f deletion_log.csv

# Install the binary
install: build
	cp nexus-retention-policy /usr/local/bin/

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run ./...

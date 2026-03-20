.PHONY: proto proto-docker test test-integration build clean help deps

help:
	@echo "Board Service Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  proto              - Generate gRPC code from proto files (requires protoc)"
	@echo "  proto-docker       - Generate gRPC code using Docker (protocolbuffers/protoc image)"
	@echo "  test               - Run unit tests"
	@echo "  test-integration   - Run integration tests"
	@echo "  build              - Build the board service binary"
	@echo "  deps               - Download and verify dependencies"
	@echo "  clean              - Remove generated files and binaries"

deps:
	@echo "Downloading and verifying dependencies..."
	go mod download
	go mod verify
	@echo "Dependencies verified"

proto:
	@echo "Generating gRPC code from proto files..."
	@if command -v protoc >/dev/null 2>&1; then \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			api/proto/v1/board.proto && \
		echo "Proto code generated successfully"; \
	else \
		echo "Error: protoc not found. Try 'make proto-docker' or install protoc manually."; \
		exit 1; \
	fi

proto-docker:
	@echo "Generating gRPC code using Docker..."
	@docker run --rm \
		-v $(PWD):/workspace \
		-w /workspace \
		protocolbuffers/protoc:latest \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/v1/board.proto && \
	echo "Proto code generated successfully via Docker"

test:
	@echo "Running unit tests..."
	go test ./internal/domain/ -v

test-integration:
	@echo "Running integration tests..."
	go test ./tests/integration/ -v

build:
	@echo "Building board service binary..."
	CGO_ENABLED=0 GOOS=linux go build -o ./bin/board ./cmd/server
	@echo "Binary built successfully at ./bin/board"

build-debug:
	@echo "Building board service binary (debug mode)..."
	go build -o ./bin/board ./cmd/server
	@echo "Binary built successfully at ./bin/board"

clean:
	@echo "Cleaning up generated files and binaries..."
	@rm -rf ./bin
	@rm -f api/proto/v1/*.pb.go
	@rm -f api/proto/v1/*_grpc.pb.go
	@echo "Cleanup completed"

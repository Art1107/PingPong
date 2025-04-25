.PHONY: proto build build-cli run clean

# Configuration
PROTO_DIR=./proto
PROTO_OUT=./proto
GO_OUT=./cmd
CLI_OUT=./cmd/cli
BINARY_NAME=pingpong
CLI_BINARY_NAME=pingpong-cli

# Generate Go code from Protocol Buffers
proto:
	@echo "Generating Go code from protobuf definitions..."
	protoc --proto_path=$(PROTO_DIR) \
		--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/pingpong.proto

# Build server binary
build: proto
	@echo "Building server binary..."
	go build -o $(GO_OUT)/$(BINARY_NAME) ./cmd

# Build CLI binary
build-cli: proto
	@echo "Building CLI binary..."
	go build -o $(CLI_OUT)/$(CLI_BINARY_NAME) ./cmd/cli

# Run the server
run: build
	@echo "Starting PingPong server..."
	$(GO_OUT)/$(BINARY_NAME)

# Run a test match
test-match: build-cli
	@echo "Starting a new test match..."
	$(CLI_OUT)/$(CLI_BINARY_NAME) new-match

# Clean up compiled files
clean:
	@echo "Cleaning up..."
	rm -f $(GO_OUT)/$(BINARY_NAME)
	rm -f $(CLI_OUT)/$(CLI_BINARY_NAME)
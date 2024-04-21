BINARY_NAME=logistics_coordinator
BUILD_DIR=bin
CMD_CLIENT=cmd/logistics
CMD_SERVER=cmd/server
VERSION?=1.0.0
LDFLAGS=-ldflags "-X main.Version=${VERSION}"
SERVER_BIN=$(BUILD_DIR)/server_$(BINARY_NAME)
CLIENT_BIN=$(BUILD_DIR)/client_$(BINARY_NAME)

all: clean build test

build:
	@echo "Building Binaries to $(BUILD_DIR)/$(BINARY_NAME)"
	@cd $(CMD_SERVER) && GO111MODULE=on go build $(LDFLAGS) -o $(SERVER_BIN) .
	@cd $(CMD_CLIENT) && GO111MODULE=on go build $(LDFLAGS) -o $(CLIENT_BIN) . 

clean:
	@echo "Cleaning..."
	@go mod tidy
	@go clean
	@rm $(CMD_CLIENT)/$(CLIENT_BIN)
	@rm $(CMD_SERVER)/$(SERVER_BIN)

test:
	@echo "Running tests..."
	@go test -v ./...       

client: clean build
	@./$(CMD_CLIENT)/$(CLIENT_BIN)

server: clean build
	@SERVER_HOST="0.0.0.0" SERVER_TCP_PORT=50051 ./$(CMD_SERVER)/$(SERVER_BIN)

genbuff:
	@cd api && buf generate 
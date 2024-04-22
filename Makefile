BINARY_NAME=logistics_coordinator
BUILD_DIR=bin
CMD_CLIENT=cmd/logistics
CMD_SERVER=cmd/server
VERSION?=1.0.0
LDFLAGS=-ldflags "-X main.Version=${VERSION}"
SERVER_BIN=$(BUILD_DIR)/server_$(BINARY_NAME)
CLIENT_BIN=$(BUILD_DIR)/client_$(BINARY_NAME)

all: clean buildserver buildclient test

buildserver:
	@echo "Building Server Binaries"
	@cd $(CMD_SERVER) && GO111MODULE=on go build $(LDFLAGS) -o $(SERVER_BIN) .

buildclient:
	@echo "Building Client Binary"
	@cd $(CMD_CLIENT) && GO111MODULE=on go build $(LDFLAGS) -o $(CLIENT_BIN) . 

clean:
	@echo "Cleaning..."
	@go mod tidy
	@go clean
	@if [ -f "$(CMD_CLIENT)/$(CLIENT_BIN)" ]; then rm "$(CMD_CLIENT)/$(CLIENT_BIN)"; fi
	@if [ -f "$(CMD_SERVER)/$(SERVER_BIN)" ]; then rm "$(CMD_SERVER)/$(SERVER_BIN)"; fi

test:
	@echo "Running tests..."
	@go test -v --race ./...       

client: clean buildclient
	@./$(CMD_CLIENT)/$(CLIENT_BIN)

server: clean buildserver
	@SERVER_HOST="0.0.0.0" SERVER_TCP_PORT=50051 ./$(CMD_SERVER)/$(SERVER_BIN)

genbuff:
	@buf generate .
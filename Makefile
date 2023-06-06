# Name of the binary to build
BINARY_NAME=xd

# Go source files
SRC=$(shell find . -name "*.go" -type f)

# Build the binary for the current platform
build:
	go build -race -o $(BINARY_NAME) -ldflags="-s -w" ./cmd/xd

# Install the binary to the user's path
install: build
	install -m 755 $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

# Uninstall the binary from the user's path
uninstall:
	rm -f /usr/local/bin/$(BINARY_NAME)

# Clean the project
clean:
	go clean
	rm -f $(BINARY_NAME)

# Run the tests
test:
	go test -v ./...

# Format the source code
fmt:
	gofmt -w $(SRC)

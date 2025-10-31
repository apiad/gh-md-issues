# This is the name of our final executable
BINARY = gh-md-issues

# The 'all' target is the default. It depends on the binary.
all: $(BINARY)

# This target builds the Go binary.
# It depends on 'main.go', which is our tangled source.
$(BINARY): main.go
	go build -o $(BINARY)
	@chmod +x $(BINARY)

# This target tidies 'main.go'.
main.go:
	@go mod tidy

# A helper to install the extension locally for testing.
install: $(BINARY)
	@echo "Installing $(BINARY) as 'gh' extension..."
	@gh extension install --force .
	@echo "Done! Try 'gh md-issues'"

# A helper to clean up build artifacts.
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY) main.go go.sum
	@go clean -modcache

.PHONY: all install clean

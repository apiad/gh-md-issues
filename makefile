# This is the name of our final executable
BINARY = gh-md-issues

# The 'all' target is the default. It depends on the binary.
all: $(BINARY)

# This target builds the Go binary.
# It depends on 'main.go', which is our tangled source.
$(BINARY): main.go
	go build -o $(BINARY)
	@chmod +x $(BINARY)

# This target creates 'main.go' from our literate source.
# It depends on 'gh-md-issues.md'. If we edit the .md file,
# 'make' will know it needs to re-run this.
main.go: gh-md-issues.md
	@echo "Tangling gh-md-issues.md -> main.go"
	@illiterate gh-md-issues.md
	@echo "Tidying go modules..."
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

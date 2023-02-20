# Define the name of the output binary
BINARY_NAME=ipamtool

# Define the source files for your application
SRC_FILES=main.go helper.go driver.go

# Define the target operating systems and architectures
TARGETS=darwin/amd64 linux/amd64 windows/amd64

# Define the Go compiler and linker flags
GO=go
GOFLAGS=build
LDFLAGS=-ldflags="-s -w"

.PHONY: all clean $(TARGETS)

all: $(TARGETS)

# Clean up the build artifacts
clean:
	rm -f $(BINARY_NAME)

# Cross-compile the application for each target
$(TARGETS):
	GOOS=$(word 1,$(subst /, ,$@)) GOARCH=$(word 2,$(subst /, ,$@)) \
	$(GO) $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME)-$(word 1,$(subst /, ,$@))-$(word 2,$(subst /, ,$@)) $(SRC_FILES)


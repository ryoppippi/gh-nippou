# Build the binary
build:
    go build -v ./...

# Run tests
test:
    go test -v ./...

# Build and test
all: build test

# Clean build artefacts
clean:
    rm -f gh-nippou gh-nippou.exe

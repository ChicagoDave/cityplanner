.PHONY: build build-solver build-renderer dev test clean

# Build everything: renderer first (Go embeds its output), then solver
build: build-renderer build-solver

build-solver:
	cd solver && go build -o cityplanner ./cmd/cityplanner

build-renderer:
	cd renderer && npm run build

# Development: start Vite dev server (solver serve mode is separate)
dev-renderer:
	cd renderer && npm run dev

dev-solver:
	cd solver && go run ./cmd/cityplanner serve ../examples/default-city/

# Run all tests
test: test-solver test-renderer

test-solver:
	cd solver && go test ./...

test-renderer:
	cd renderer && npm test

# Clean build artifacts
clean:
	rm -f solver/cityplanner
	rm -rf renderer/dist

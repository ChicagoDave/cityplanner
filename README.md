# CityPlanner

A specification-driven design engine for a car-free, solar-powered charter city of 50,000 residents.

Write a declarative YAML spec. The solver validates constraints, generates spatial layouts, and produces a 3D scene graph. The renderer lets you explore the result interactively in your browser.

## Architecture

```
city.yaml  →  Go Solver  →  scene.json  →  Three.js Renderer
  (spec)      (CLI/server)   (scene graph)    (browser)
```

- **Solver** — Go binary. Parses the spec, runs a two-phase constraint solver (analytical resolution + spatial generation), outputs a JSON scene graph.
- **Renderer** — TypeScript/Three.js web app. Loads the scene graph, renders the city in 3D with camera modes, layer toggling, and cross-section views.
- **JSON Schema** — The scene graph format is the contract between solver and renderer. Both are independently buildable and testable.

See [docs/architecture/adrs/](docs/architecture/adrs/) for detailed design decisions.

## Quick Start

### Prerequisites

- Go 1.25+ ([go.dev/dl](https://go.dev/dl/))
- Node.js 22+ ([nodejs.org](https://nodejs.org/))

### Build

```bash
# Build everything
make build

# Or separately:
cd solver && go build -o cityplanner ./cmd/cityplanner
cd renderer && npm install && npm run build
```

### Run

```bash
# Validate a city spec
./solver/cityplanner validate examples/default-city/

# Run the full solver
./solver/cityplanner solve examples/default-city/

# Start the interactive dev server
./solver/cityplanner serve examples/default-city/
```

### Development

```bash
# Terminal 1: Go server (solver API on :3000)
make dev-solver

# Terminal 2: Vite dev server (renderer on :5173, proxies /api to :3000)
make dev-renderer
```

## Project Structure

```
solver/                  Go module — solver + CLI + dev server
  cmd/cityplanner/       CLI entry point (solve, validate, cost, serve)
  pkg/spec/              City spec types and YAML parsing
  pkg/analytics/         Phase 1: analytical constraint resolution
  pkg/layout/            Pod layout (Voronoi) and building placement
  pkg/routing/           Underground infrastructure routing
  pkg/scene/             Scene graph types and JSON serialization
  pkg/cost/              Cost model computation
  pkg/validation/        Structured error reporting
  internal/server/       HTTP server for serve mode

renderer/                TypeScript + Vite + Three.js
  src/main.ts            App entry, Three.js scene setup
  src/api.ts             REST client for solver API
  src/scene/             Scene graph loading
  src/camera/            Camera mode management
  src/ui/                Controls and overlays

shared/schema/           JSON Schemas (language-neutral contract)
examples/default-city/   Sample city spec and project manifest
docs/                    Vision document, technical spec, 14 ADRs
```

## Documentation

- [Vision Document](docs/specification/charter_city_vision.md) — What the city is and why
- [Technical Spec](docs/specification/charter_city_technical_spec.md) — Formulas, data model, engine spec
- [ADRs](docs/architecture/adrs/) — 14 architecture decision records covering the full pipeline

## Status

Scaffold complete. All packages compile and build. Solver logic is not yet implemented — all packages are stubs with TODOs.

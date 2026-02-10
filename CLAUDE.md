# CLAUDE.md

## Project

Charter city design engine: YAML spec → Go solver → JSON scene graph → Three.js renderer.

## Architecture

- **Go solver** (`solver/`): CLI binary with four commands — `solve`, `validate`, `cost`, `serve`. Module path: `github.com/ChicagoDave/cityplanner`.
- **TypeScript renderer** (`renderer/`): Vite + Three.js browser app. Fetches scene data from the solver's HTTP API.
- **Boundary contract**: JSON scene graph (`shared/schema/scene-graph.schema.json`). Solver writes it, renderer reads it. No shared code between Go and TypeScript.
- **Two-phase solver** (ADR-009): Phase 1 is analytical (pure math, microseconds). Phase 2 is spatial generation (pod layout, building placement, infrastructure routing).

## Build & Run

```bash
# Go — always export PATH when running go commands
export PATH=$PATH:/usr/local/go/bin

# Build solver
cd solver && go build -o cityplanner ./cmd/cityplanner

# Build renderer
cd renderer && npm install && npm run build

# Run tests
cd solver && go test ./...
cd renderer && npm test
```

## Code Conventions

### Go (solver/)
- Standard Go project layout: `cmd/` for entry points, `pkg/` for public packages, `internal/` for private packages
- Use `gopkg.in/yaml.v3` for YAML parsing
- Use `github.com/spf13/cobra` for CLI
- Use stdlib `net/http` for the server — no web framework
- Validation always returns structured `validation.Report` with `validation.Result` items
- Every solver stage returns `(result, *validation.Report)` — errors are reported, not panicked

### TypeScript (renderer/)
- Strict TypeScript, ES2022 target
- Vite for bundling, no other framework
- API types defined in `src/api.ts` match the scene graph JSON schema
- Vite dev server proxies `/api` to `localhost:3000` (Go server)

## Key Files

- `examples/default-city/city.yaml` — Reference spec with all values from the technical spec
- `shared/schema/city-spec.schema.json` — JSON Schema for the YAML spec format
- `shared/schema/scene-graph.schema.json` — JSON Schema for solver output
- `solver/pkg/spec/types.go` — Go types for the city spec
- `solver/pkg/scene/scene.go` — Go types for the scene graph

## ADRs

All architectural decisions are documented in `docs/architecture/adrs/`. Key ones:
- ADR-001: Pipeline architecture (Go binary as CLI + local dev server)
- ADR-004: Go solver + TypeScript renderer
- ADR-005: Constrained Voronoi pod layout
- ADR-008: Scene graph structure (layered hybrid with groups)
- ADR-009: Two-phase solver strategy

## Environment

- WSL2 on Windows
- Go 1.25.7 at `/usr/local/go/bin` (must be added to PATH)
- Node 22.14.0 via nvm
- `code .` does not work (WSL binfmt interop issue)

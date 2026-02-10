# ADR-001: Overall Pipeline Architecture

**Status:** Proposed
**Date:** 2026-02-09
**Deciders:** David Cornelson

---

## Context

The Charter City design tool follows a four-stage pipeline as described in the technical specification:

1. **Spec** — Declarative YAML input defining all city parameters
2. **Solver** — Constraint satisfaction engine that generates spatial data
3. **Scene Graph** — Traversable 3D spatial structure (JSON)
4. **Renderer** — Interactive 3D visualization (Three.js in browser)

The spec explicitly states the solver should be "fully decoupled from the renderer, runnable headless for batch evaluation."

ADR-004 decides the solver is written in Go and the renderer is TypeScript/Three.js. This means:

- The solver compiles to a native Go binary
- The renderer is a web application running in the browser
- They cannot share memory or function calls — they communicate via serialized data (JSON)

This creates a fundamental constraint that the current options must address: **the browser cannot read local files or spawn local processes**. The renderer needs some mechanism to obtain the scene graph from the solver. This is the core architectural question.

## Decision Drivers

- The solver must run headless for batch evaluation, CI, and automated testing — independent of any renderer
- The design loop (edit spec → solve → view → adjust → re-solve) must feel near-instant for iterative use
- The renderer is a browser app (Three.js) — it cannot access the local filesystem or spawn processes
- A full city scene graph may be 10-100 MB of JSON — transfer mechanism must handle this efficiently
- The user experience should be: one command to start working, minimal manual steps per iteration
- The solver and renderer are in different languages (Go and TypeScript) — the boundary is necessarily serialization-based
- The architecture should start simple but not preclude collaborative or remote use later

## Options

### Option A: Pure CLI + Static File Renderer

The solver is a CLI: `cityplanner solve city.yaml -o city.scene.json`. The renderer is a static web page. The user opens the scene graph file in the renderer via file picker, drag-and-drop, or a local file URL.

**How it works:**
```
1. User edits city.yaml in their editor
2. User runs: cityplanner solve my-city/
3. Solver writes city.scene.json
4. User opens renderer.html in browser
5. User clicks "Load" and selects city.scene.json via file picker
6. User sees the 3D city
7. User edits city.yaml, re-runs solver, re-loads file in browser
```

**Pros:**
- Simplest implementation — solver is just a CLI, renderer is just static HTML/JS/CSS
- No HTTP server to build or run
- Solver is headless by default — no special mode needed
- Renderer can be hosted anywhere (GitHub Pages, CDN) as a static site
- Zero dependencies beyond Go and a browser

**Cons:**
- **Clunky iteration loop.** Every re-solve requires: switch to terminal → run command → switch to browser → re-load file. This is 4 manual steps per iteration across 50+ iterations per session. The friction kills the design loop.
- No automatic re-rendering when the spec changes
- File picker has browser security restrictions (must re-select file each time in many browsers)
- No watch mode possible without a server — browser can't observe local file changes
- User must manually keep solver output and renderer input in sync
- Feels like 1990s tooling — disconnected executables with manual file passing

### Option B: Go Binary as Local Dev Server

The Go binary serves dual roles: it's the solver AND a local HTTP server that serves the renderer. One command starts everything: `cityplanner serve my-city/`. The browser opens automatically. The web UI has a "Regenerate" button that triggers a re-solve via HTTP API. File watching auto-triggers re-solves on spec changes.

**How it works:**
```
1. User runs: cityplanner serve my-city/
2. Browser opens to http://localhost:3000
3. Three.js renderer loads, fetches scene graph via GET /api/scene
4. User sees the 3D city
5. User edits city.yaml in their editor
6. File watcher detects change, re-runs solver, pushes update via WebSocket
7. Renderer updates automatically
```

The Go binary bundles the renderer's compiled JavaScript/HTML as embedded static assets (Go's `embed` package). No separate web server or Node.js process needed.

**Architecture:**
```
cityplanner binary (Go)
  ├── CLI mode:    cityplanner solve my-city/     → headless, writes JSON
  ├── CLI mode:    cityplanner validate my-city/  → headless, prints report
  └── Serve mode:  cityplanner serve my-city/     → starts local server
        ├── Static file server: serves bundled renderer HTML/JS/CSS
        ├── GET /api/scene         → returns current scene graph JSON
        ├── GET /api/cost          → returns current cost report
        ├── GET /api/validation    → returns validation results
        ├── POST /api/solve        → triggers re-solve, returns updated scene graph
        └── WebSocket /api/ws      → pushes scene updates on spec file change
```

**Pros:**
- **Best iteration experience.** One command to start. Edit YAML → see result automatically. Zero manual steps per iteration after startup.
- Solver still runs headless via `cityplanner solve` — the serve mode is additive, not replacing
- Single binary — no separate processes, no Node.js runtime, no npm install for the renderer
- Go's `embed` package bundles the compiled renderer JS/HTML into the binary — distribution is one file
- File watching + WebSocket push means the browser updates within seconds of saving the spec
- The HTTP API is a natural foundation for future collaborative/remote use
- Go is excellent at HTTP servers — this adds minimal complexity to the binary
- The renderer's TypeScript code never needs to know about files or processes — it just fetches JSON from an API

**Cons:**
- More complex than a pure CLI — the Go binary now includes an HTTP server and WebSocket handler
- The renderer's compiled JS must be built separately (TypeScript → JavaScript) and then embedded into the Go binary at build time — adds a build step
- The HTTP API must be designed and versioned (though it's simple: serve JSON, trigger solve)
- Localhost-only by default — exposing to the network requires explicit configuration
- The bundled renderer JS is frozen at Go build time — updating the renderer requires rebuilding the Go binary

### Option C: Separate Solver CLI + Separate Renderer Dev Server

The solver is a standalone Go CLI. The renderer is a standalone Node.js/Vite dev server. They coordinate via the filesystem — solver writes JSON, renderer's dev server watches for changes and hot-reloads.

**How it works:**
```
Terminal 1: cityplanner watch my-city/     → watches YAML, re-solves on change
Terminal 2: cd renderer && npm run dev     → Vite dev server on localhost:5173
```

The Vite dev server serves the renderer and watches the scene graph JSON file. When it changes, the browser hot-reloads.

**Pros:**
- Standard web development workflow — familiar to anyone who's used Vite/webpack
- Hot module replacement for renderer development (edit Three.js code, see changes instantly)
- Solver and renderer are fully independent — each has its own build, its own process
- Good for active renderer development — the Vite dev server provides source maps, fast rebuilds, etc.

**Cons:**
- **Two terminals, two processes, two toolchains.** User must start both and keep both running.
- Requires Node.js installed for the renderer (in addition to Go for the solver)
- File-watching coordination between solver and renderer is fragile — race conditions on write/read
- Distribution requires shipping both a Go binary and a Node.js project
- Not a good end-user experience — fine for development, bad for someone evaluating a city design
- The Vite dev server is a development tool, not a production delivery mechanism

### Option D: Electron/Tauri Desktop Application

The renderer is packaged as a desktop application (Electron or Tauri) with filesystem access. It spawns the solver as a child process and reads the scene graph directly.

**How it works:**
```
1. User launches the CityPlanner desktop app
2. App spawns solver subprocess for solving
3. User edits spec in built-in editor or watches external file
4. App calls solver, reads output, updates 3D view
```

**Pros:**
- Single application with full filesystem access — no browser limitations
- Can spawn solver subprocess directly — tightest possible integration
- Distribution as a native app (.exe, .app, .dmg)
- Could include a built-in spec editor alongside the 3D view

**Cons:**
- **Enormous complexity increase.** Electron bundles Chromium (~150MB). Tauri is lighter but still a significant framework.
- Desktop app development, packaging, and distribution are each major efforts
- Cross-platform testing (Windows, macOS, Linux) multiplies QA work
- Overkill for the current stage — adds months of work before the first useful iteration
- The solver's headless mode still needs to exist separately for CI/batch
- Locks the renderer into a desktop form factor — can't share a design via URL

## Recommendation

**Option B: Go Binary as Local Dev Server.**

This gives the best iteration experience (edit YAML → see result automatically) with the simplest distribution (one binary). The key insight is that the Go binary already exists as the solver — adding an HTTP server and embedding the renderer's compiled JS is a modest increment, not a new project.

### Why Not Option A (Pure CLI + Static)

The design loop is everything. This tool's value is the tight feedback cycle: tweak a parameter, see the result, tweak again. Option A's manual file-passing breaks that loop. Every friction point (switch windows, run a command, pick a file, wait) compounds across 50+ iterations per session into a genuinely unpleasant experience. The spec describes a "design loop" — Option A doesn't deliver a loop, it delivers a sequence of manual steps.

### Why Not Option C (Separate Dev Servers)

Option C is the right choice *during active renderer development* — Vite's hot module replacement is essential when iterating on Three.js code. But it's wrong as the production architecture. End users shouldn't need Node.js installed, shouldn't manage two terminals, and shouldn't coordinate two file-watching processes.

**The resolution: use Option C during renderer development, then build/embed the result into Option B for production.** The renderer is developed with Vite (Option C), compiled to static JS/HTML, and embedded into the Go binary (Option B). Development and production use different modes of the same architecture.

### Why Not Option D (Desktop App)

Premature. Desktop packaging is a major effort with marginal benefit over a localhost web app. The browser provides the rendering surface Three.js needs. If a desktop app is wanted later, Tauri can wrap the existing web renderer with minimal changes — but that's a distribution decision, not an architecture decision, and it can wait.

### The CLI Remains First-Class

Critically, Option B does not replace the headless CLI. The Go binary supports both modes:

```
cityplanner solve my-city/        → headless, writes city.scene.json (for CI, batch, scripting)
cityplanner validate my-city/     → headless, prints validation report
cityplanner cost my-city/         → headless, prints cost report
cityplanner serve my-city/        → starts local server, opens browser (for interactive design)
```

The serve mode is a convenience layer over the same solver that runs headless. There is no separate "server version" — it's one binary, multiple entry points.

### Build Pipeline

```
                                    ┌─────────────────────┐
  renderer/src/*.ts  ──→  Vite  ──→ │  renderer/dist/     │
                           build    │  (static JS/HTML)   │
                                    └────────┬────────────┘
                                             │ go:embed
                                             ▼
  solver/**.go  ─────────→  go build  ──→  cityplanner binary
                                           (solver + embedded renderer)
```

During development:
- Renderer developer runs `cd renderer && npm run dev` (Vite dev server with HMR)
- Solver developer runs `cityplanner serve my-city/` (Go binary serves solver API, renderer fetches from it)
- Both can run simultaneously — the Vite dev server proxies API calls to the Go server

For distribution:
- `npm run build` compiles the renderer to static assets
- `go build` embeds those assets and produces a single binary
- The binary is the complete tool — no Node.js, no npm, no separate files

### API Design

The HTTP API is deliberately minimal:

| Endpoint | Method | Description |
|---|---|---|
| `/` | GET | Serves the renderer (embedded HTML/JS) |
| `/api/scene` | GET | Returns current scene graph JSON |
| `/api/cost` | GET | Returns current cost report JSON |
| `/api/validation` | GET | Returns current validation report JSON |
| `/api/solve` | POST | Triggers re-solve with current spec, returns scene graph |
| `/api/spec` | GET | Returns current spec YAML (for display in renderer UI) |
| `/api/ws` | WebSocket | Pushes events: `scene_updated`, `solve_started`, `solve_failed` |

The solver caches results in memory. `GET /api/scene` returns the cached scene graph instantly. `POST /api/solve` re-runs the solver and updates the cache. The file watcher triggers `POST /api/solve` automatically on spec file changes and pushes the update via WebSocket.

## Consequences

- The Go binary is both CLI tool and local dev server — one codebase, two modes
- Must define the HTTP API contract (endpoints above) in addition to the scene graph JSON contract (ADR-008)
- The renderer's TypeScript code talks to a REST API + WebSocket — standard web development patterns
- The renderer build (Vite → static assets) is a prerequisite for the Go build — the CI pipeline must run TypeScript build before Go build
- Go's `embed` package makes bundling static assets trivial — no complex packaging
- Distribution is a single binary per platform (cross-compiled with `GOOS`/`GOARCH`)
- Watch mode with WebSocket push is the key UX feature — spec file changes should appear in the browser within 1-2 seconds
- Option C (Vite dev server) is the renderer development workflow, not the production architecture — both coexist naturally
- The localhost server architecture is a natural stepping stone to Option C-style remote/collaborative use if needed later — just expose the port

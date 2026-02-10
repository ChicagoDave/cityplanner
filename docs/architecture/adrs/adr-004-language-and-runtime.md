# ADR-004: Programming Language and Runtime

**Status:** Proposed
**Date:** 2026-02-09
**Deciders:** David Cornelson

---

## Context

The tool has two major computational components: the solver (constraint satisfaction, spatial layout, infrastructure routing) and the renderer (3D visualization). ADR-001 decouples them. ADR-003 recommends Three.js for the renderer, which requires TypeScript/JavaScript. The solver has no graphics dependency and could be written in any language.

The solver is computation-heavy: Voronoi tessellation, network flow optimization across multiple infrastructure systems, procedural building placement with collision detection, and constraint iteration loops. The spec's design loop (write → solve → view → adjust → re-solve) means solve time directly impacts usability. A 30-second solve in TypeScript vs. a 1-second solve in a compiled language is a meaningful UX difference across dozens of iterations.

Development is AI-assisted (Claude Code as primary author, David reviewing and directing). The agent writes any language fluently, so the choice is driven by **runtime performance** and **human readability** of the output, not developer velocity.

## Decision Drivers

- **Solver performance is the priority** — the design loop demands fast solves (target: under 2 seconds)
- Voronoi tessellation, graph-based network flow, spatial indexing, and procedural generation are all CPU-intensive
- The solver must handle the current 50,000-person city and scale to larger configurations
- The renderer is TypeScript/Three.js (per ADR-003) — the solver is a separate process (per ADR-001)
- David needs to read and understand the solver code
- Compile times affect the agent's iteration speed during development
- The solver communicates with the renderer via serialized JSON (scene graph) — language interop is at the file boundary, not the function boundary

## Options

### Option A: TypeScript Throughout

Solver, scene graph, and renderer all in TypeScript. Solver runs in Node.js headless.

**Pros:**
- One language, one toolchain — simplest to maintain
- Shared types between solver and renderer
- Near-instant compilation for fast agent iteration

**Cons:**
- Node.js is 10-50x slower than compiled languages for computation-heavy work
- No true parallelism — Worker threads exist but are clunky for shared-state computation
- GC pauses during large solver runs degrade predictability
- Scales poorly: acceptable at 50K population, likely painful at 100K+
- Voronoi, network flow, and spatial indexing algorithms are exactly the kind of work where interpreted languages struggle

**Note on tsc-go (TypeScript 7):** Microsoft is porting the TypeScript compiler itself from TypeScript to Go (Project Corsa / TypeScript 7, targeting mid-2026). This makes TypeScript *compilation* 10x faster but does **not** change Node.js/V8 runtime performance. The solver's bottleneck is execution speed, not compilation speed. tsc-go does not close the gap. In fact, it reinforces the point — Microsoft faced the same "JavaScript is too slow for this computation" problem with their own compiler and chose Go to solve it.

### Option B: Rust Solver + TypeScript Renderer

Solver in Rust, compiled to native CLI. Renderer in TypeScript/Three.js. Communication via JSON scene graph file.

**Pros:**
- Maximum performance — 10-100x faster than TypeScript for solver workload
- True parallelism via Rayon for infrastructure routing (route water, sewage, electrical concurrently)
- Memory-efficient: no GC overhead, predictable performance
- WASM compilation enables running the solver in-browser later
- Rich ecosystem for computational geometry (geo, delaunator, petgraph)

**Cons:**
- Rust is harder for David to read — ownership, lifetimes, trait bounds add noise to domain logic
- Compile times (30s-2min) slow the agent's development iteration loop
- Two toolchains: cargo + npm
- Steeper learning curve if David wants to make solver changes directly

### Option C: Go Solver + TypeScript Renderer

Solver in Go, compiled to native CLI. Renderer in TypeScript/Three.js. Communication via JSON scene graph file.

**Pros:**
- Fast: compiled native code, typically 5-20x faster than TypeScript for computation
- Excellent concurrency via goroutines — natural fit for parallel infrastructure routing
- **Highly readable** — Go reads close to pseudocode, minimal syntax noise
- Sub-second compile times — fast agent iteration loop, comparable to TypeScript
- First-class JSON marshaling — clean serialization to scene graph format
- Simple toolchain: single `go build` command, no package manager complexity
- David can read Go solver code with minimal learning curve

**Cons:**
- Not as fast as Rust for tight numerical loops (GC exists, though tunable)
- No WASM compilation path (Go's WASM output is large and slow)
- Smaller computational geometry ecosystem than Rust (but sufficient: libraries exist for Voronoi, Delaunay, R-trees)
- Two languages in the project (Go + TypeScript), though cleanly separated at the JSON boundary

### Option D: TypeScript with WASM Hot Paths

Solver in TypeScript with performance-critical algorithms in Rust compiled to WASM.

**Pros:**
- Mostly TypeScript for readability
- Rust WASM modules for proven bottlenecks only

**Cons:**
- WASM interop overhead (data copying across boundary)
- Debugging across TypeScript/WASM boundary is painful
- Hybrid complexity without the full benefit of either language
- Premature architecture — adds Rust before knowing which algorithms are bottlenecks

## Recommendation

**Option C: Go Solver + TypeScript Renderer.**

Performance must come first for the solver — the design loop depends on it. Between Rust and Go, Go is the better fit for this project.

### Why Performance Rules Out TypeScript for the Solver

The initial version of this ADR recommended TypeScript throughout, reasoning that the solver's workload (~13 pods, a few thousand buildings) was manageable in Node.js. On reassessment, this underweights several factors:

- **The design loop is iterative.** A user may re-solve 50+ times in a session, tweaking parameters. The difference between a 5-second solve and a 0.5-second solve compounds across those iterations into minutes of waiting vs. near-instant feedback.
- **The city may scale.** 50,000 is the starting target. A 200,000-person city with 50+ pods and proportionally more infrastructure is a plausible future configuration. TypeScript's performance ceiling becomes a hard wall.
- **Computational geometry is V8's weak spot.** Tight loops over floating-point arrays, spatial indexing, and graph traversal are exactly where interpreted languages lose the most ground to compiled code. This isn't web app logic where V8 excels — it's numerical computation.

### Why Go Over Rust

The TypeScript team's experience porting the TypeScript compiler (Project Corsa) provides a direct precedent. They faced the same decision — JavaScript/TypeScript was too slow for their computational workload, and they needed a compiled language. Anders Hejlsberg's team evaluated Rust, Go, and C#, and chose Go. Their reasoning aligns closely with ours:

1. **Performance is sufficient without Rust's ceiling.** The TypeScript compiler port achieved 10x speedups in Go. Our solver needs similar gains (5-20x over TypeScript), not the 100x that Rust enables for systems-level work. Go's compiled speed comfortably clears our 2-second solve target. Rust's additional performance margin is real but unnecessary for computational geometry and graph algorithms at this scale.

2. **Structural similarity and readability.** The TypeScript team found that Go's programming style closely resembled their existing TypeScript codebase, making porting tractable and the result readable. The same applies here: Go code reads like the domain. David can read `func layoutPods(rings []Ring, seeds []Point) []Pod` and understand it immediately. Rust's `fn layout_pods(rings: &[Ring], seeds: &mut Vec<Point>) -> Result<Vec<Pod>, SolverError>` adds ownership semantics, lifetime annotations, and trait bounds that obscure the domain logic behind language mechanics.

3. **GC is fine for this workload.** The TypeScript team explicitly decided Rust's manual memory management wasn't worth the complexity for their compiler. Our solver has a similar profile: allocate data structures, compute, output JSON, exit. There are no long-lived processes where GC pauses matter. Go's GC is tunable and sub-millisecond for typical heap sizes in our workload.

4. **Compile speed matches the workflow.** Go compiles in under a second. Rust builds take 30 seconds to 2 minutes. In AI-assisted development where the agent runs write → compile → test → adjust loops hundreds of times, this difference compounds enormously. The TypeScript team cited the same concern — fast iteration mattered for their development process.

5. **Concurrency is built in.** Go's goroutines are ideal for parallel infrastructure routing: `go routeWater(pods)`, `go routeSewage(pods)`, `go routeElectrical(pods)`. Rust has excellent concurrency via Rayon, but Go's model is simpler and more natural for coarse-grained parallelism across independent solver stages.

6. **Clean separation makes two languages painless.** The solver is a CLI that reads a YAML spec and writes a JSON scene graph. The renderer is a TypeScript/Three.js app that reads the JSON. They share no code — the JSON schema is the contract. Two languages is fine when the boundary is a file, not a function call. The TypeScript team made the same architectural choice: the Go compiler is a separate tool that produces JavaScript output consumed by the existing ecosystem.

### What About AI-Assisted Development?

The original version of this ADR weighted "easier to hire" and "team onboarding" as decision factors. With an AI agent as the primary developer, those factors are irrelevant — the agent writes Go, Rust, and TypeScript equally well. This shifts the decision entirely to:

- **Runtime performance** (favors Go or Rust over TypeScript)
- **Human readability** of the output (favors Go over Rust — David reviews and maintains the code)
- **Compile speed** for the agent's iteration loop (favors Go over Rust)

The AI agent eliminates the traditional "Go/Rust is slower to develop in" argument. What remains is purely: how fast does the solver run, how fast does it compile, and can the human owner read the code? Go wins on all three against TypeScript, and wins on readability + compile speed against Rust while providing sufficient runtime performance.

### Project Structure

```
solver/              — Go module
  cmd/solver/        — CLI entry point
  pkg/spec/          — Spec parsing and validation
  pkg/analytics/     — Phase 1: analytical resolution
  pkg/layout/        — Pod layout, building placement
  pkg/routing/       — Infrastructure network routing
  pkg/scene/         — Scene graph generation and JSON output
  pkg/cost/          — Cost model computation

renderer/            — TypeScript/Three.js (npm workspace or standalone)
  src/               — Three.js application, camera, controls, UI

shared/
  schema/            — JSON Schema for spec and scene graph (language-neutral)
```

## Consequences

- The solver is a standalone Go binary — runs anywhere, no runtime dependencies
- David reads Go for solver logic and TypeScript for renderer logic — both are accessible languages
- The agent iterates at full speed with sub-second Go compilation
- Parallel infrastructure routing via goroutines is straightforward to implement
- The JSON scene graph schema (ADR-008) becomes the critical contract between solver and renderer — must be well-specified and versioned
- If Go's performance ceiling is ever reached (unlikely for this problem scale), specific algorithms can be rewritten in Rust as Go CGo calls or as a separate pre-processing step
- The Go ecosystem has adequate libraries for Voronoi tessellation, Delaunay triangulation, R-trees, and graph algorithms
- This decision follows the same path as Microsoft's TypeScript 7 (Project Corsa), which ported the TypeScript compiler from TypeScript to Go for the same performance, readability, and concurrency reasons

## References

- [Progress on TypeScript 7 - December 2025](https://devblogs.microsoft.com/typescript/progress-on-typescript-7-december-2025/) — Status of the Go-based TypeScript compiler
- [Microsoft Ports TypeScript to Go for 10x Native Performance Gains](https://visualstudiomagazine.com/articles/2025/03/11/microsoft-ports-typescript-to-go-for-10x-native-performance-gains.aspx) — Announcement and rationale
- [TypeScript Migrates to Go: What's Really Behind That 10x Claim](https://www.architecture-weekly.com/p/typescript-migrates-to-go-whats-really) — Analysis of the Rust vs Go decision
- [A closer look at the details behind the Go port](https://2ality.com/2025/03/typescript-in-go.html) — Technical deep dive on the porting approach

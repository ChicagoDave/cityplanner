# ADR-002: Specification Format and Schema

**Status:** Proposed
**Date:** 2026-02-09
**Deciders:** David Cornelson

---

## Context

The city is defined entirely through a declarative specification. The technical spec uses YAML examples for all city parameters: population, demographics, zone definitions, pod assignments, infrastructure systems, and vehicle/logistics configuration. This spec is the single source of truth that drives the solver.

We need to decide the file format, how the schema is defined and validated, how defaults work when optional values are omitted, and how the format evolves over time.

## Decision Drivers

- The spec is human-authored and frequently edited (the core design loop)
- The spec must be machine-parseable with strong validation
- Errors in the spec should produce clear, actionable messages
- The format must support future extension without breaking existing specs
- The technical spec already uses YAML in its examples

## Options

### Option A: YAML Only

Spec files are written in YAML. Schema is enforced programmatically after parsing.

**Pros:**
- Matches the existing spec document examples exactly
- Human-friendly: comments, minimal punctuation, readable nesting
- Well-suited to the "write spec, adjust, regenerate" workflow

**Cons:**
- YAML parsing has well-known gotchas (Norway problem, implicit type coercion)
- No native schema standard as mature as JSON Schema
- Whitespace sensitivity can cause subtle errors

### Option B: JSON Only

Spec files are written in JSON. Schema is enforced via JSON Schema.

**Pros:**
- Unambiguous parsing — no implicit coercion
- JSON Schema is a mature, well-tooled standard
- Direct interop with JavaScript/TypeScript (no parser needed)

**Cons:**
- Verbose for a human-edited configuration file
- No comments (unless using JSON5 or JSONC)
- Poor ergonomics for the iterative design loop

### Option C: YAML Input with JSON Schema Validation

Spec files are authored in YAML. The build pipeline parses YAML to a JavaScript object, then validates against a JSON Schema. Internally the solver works with typed objects derived from the schema.

**Pros:**
- Best authoring experience (YAML) with strongest validation (JSON Schema)
- JSON Schema generates TypeScript types automatically (e.g., via `json-schema-to-typescript`)
- Schema serves as documentation, validation, and type generation in one artifact
- Editors can provide autocomplete and validation via YAML Language Server + JSON Schema

**Cons:**
- Two formats to understand (YAML for authoring, JSON Schema for validation)
- YAML parsing gotchas still apply at the input layer

### Option D: TypeScript DSL

Spec files are TypeScript modules that export a city configuration object. Type safety is enforced by the TypeScript compiler.

**Pros:**
- Full IDE support: autocomplete, inline errors, refactoring
- Can include computed values (e.g., `population * 100` for water capacity)
- Type safety without a separate schema layer

**Cons:**
- Requires TypeScript tooling to author a spec — higher barrier to entry
- Blurs the line between spec (data) and code (logic)
- Harder to version and diff as a pure data artifact

## Recommendation

**Option C: YAML Input with JSON Schema Validation.**

YAML is the natural authoring format given the existing spec examples and the iterative design workflow. JSON Schema provides rigorous validation, auto-generated TypeScript types, and editor integration. The schema file becomes the canonical definition of what a valid city spec looks like.

### Schema Design Principles

- Every field has a default value or is explicitly required
- Defaults match the values in the technical spec (e.g., `walk_radius: 400` default)
- Validation produces errors with field paths and human-readable explanations
- The schema includes `description` fields so it doubles as documentation
- A `spec_version` field at the root enables forward-compatible evolution

## Consequences

- Must author and maintain a JSON Schema file as the canonical spec definition
- Must select a YAML parser (e.g., `yaml` npm package) and a JSON Schema validator (e.g., `ajv`)
- Generated TypeScript types become the solver's input interface
- A `validate-spec` CLI command should exist for quick feedback without running the full solver
- The schema must be versioned; breaking changes require a migration path or version bump

---
name: implement
description: Implement or refactor scoped changes in the gopdsdk Go repository. Use when Codex is asked to add a feature, fix a bug, change package structure, add tests, or otherwise modify project code while preserving the feature-first architecture and cross-platform toolchain contract.
---

# Implement

Follow `AGENTS.md`; it is the repository contract.

## Workflow

1. Inspect `git status`, the target feature, its tests, and package comments.
2. Classify each component:
   - keep feature-specific code in `internal/features/<feature>`;
   - move code to `internal/shared/<component>` only after a second real feature
     consumer exists;
   - keep `cmd/<binary>` limited to composition and process concerns.
3. State assumptions and the package boundary before editing.
4. Implement the smallest complete behavior, including error paths.
5. Add `doc.go` for every new package with `// Package <name>`.
6. Add deterministic tests. Abstract filesystem or process behavior only where
   tests or platform boundaries require it.
7. Format and run the checks from `AGENTS.md` plus the relevant end-to-end path.
8. Inspect the diff for unrelated edits, platform assumptions, missing
   provenance, and false readiness claims.

Do not add dependencies, generated bindings, public API, or shared packages
speculatively. Do not copy implementation material from pdgo.

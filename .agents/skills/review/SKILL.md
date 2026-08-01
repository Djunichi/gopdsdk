---
name: review
description: Review gopdsdk Go changes for correctness, architecture, portability, tests, provenance, and readiness risks without fixing them.
---

# Review

Follow `AGENTS.md`. Remain read-only unless fixes are requested.

## Workflow

1. Scope from the request, `git status`, and diff; read complete changed
   functions and callers.
2. Use focused tests or read-only diagnostics to confirm suspected defects.
3. Check behavior/errors/cleanup; exit semantics; Windows/macOS/Linux process
   and path differences; feature cohesion; package comments; deterministic
   tests; probe integrity; external-module coverage for CLI/module changes;
   third-party provenance; security and destructive behavior.
4. Reject readiness claims that promote CI, Docker, cross-compilation, or
   dry-runs to SDK, Simulator, USB, or hardware evidence.
5. Report actionable findings by severity with precise lines, failing scenario,
   and impact. If none, say so and list material verification gaps.

Do not report ungrounded style preferences. Do not claim a defect without a
concrete execution path or violated invariant.

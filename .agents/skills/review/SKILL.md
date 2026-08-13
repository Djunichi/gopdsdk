---
name: review
description: Review gopdsdk Go changes for correctness and risk without fixing them.
---

# Review

Follow `AGENTS.md`. Remain read-only unless fixes are requested.

## Workflow

1. Scope from the request, status, and diff; read changed functions and callers.
2. Use focused tests or read-only diagnostics to confirm suspected defects.
3. Check behavior, errors, cleanup, exit semantics, platform process/path
   differences, architecture, deterministic tests, probe integrity,
   external-module coverage, provenance, security, and destructive behavior.
4. Report actionable findings by severity with precise lines, failing scenario,
   and impact. If none, say so and list material verification gaps.

Do not report ungrounded style preferences. Do not claim a defect without a
concrete execution path or violated invariant.

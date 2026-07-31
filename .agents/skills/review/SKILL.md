---
name: review
description: Review gopdsdk Go code, diffs, or proposed changes for correctness, regressions, architecture, portability, test coverage, and independent-implementation risks. Use when Codex is asked to review code, audit a change before commit, assess a refactor, or provide actionable findings without implementing fixes.
---

# Review

Follow `AGENTS.md`. Review read-only unless the user explicitly requests fixes.

## Workflow

1. Establish scope from the request, `git status`, and the relevant diff.
2. Read complete changed functions and their callers, not only diff fragments.
3. Run focused tests or read-only diagnostics to confirm suspected problems.
4. Check, in order:
   - behavior, errors, cleanup, and exit semantics;
   - Windows, macOS, and Linux path, executable, and process differences;
   - feature cohesion and premature `internal/shared` extraction;
   - required `doc.go` package comments;
   - deterministic tests and capability probe integrity;
   - accidental copying or structural imitation of third-party implementation;
   - security, destructive behavior, and secret exposure.
5. Report actionable findings ordered by severity, with precise file and line
   references. Explain the failing scenario and impact.
6. If no findings remain, say so and list material verification gaps briefly.

Do not report ungrounded style preferences. Do not claim a defect without a
concrete execution path or violated invariant.

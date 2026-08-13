---
name: update-docs
description: Synchronize gopdsdk public docs, roadmap, changelog, migrations, and evidence after implemented changes; not code or releases.
---

# Update Docs

Follow `AGENTS.md`. Treat implementation and verification output as evidence.
Do not change code, generated files, tests, or native assets.

## Workflow

1. Inspect `git status`, the relevant diff, current documentation, and the
   exact verification or acceptance results. Preserve unrelated user edits.
2. Identify the public behavior that actually changed, its active roadmap
   scope, and the strongest verified evidence level. Ask only when a missing
   fact would materially change a claim.
3. Update only applicable files under the `AGENTS.md` routing. Use
   `MIGRATING.md` only for required user action after an intentional breaking
   change. Include exact dates, tool versions, and artifact sizes when known.
4. Keep evidence at its verified level. Never infer visual or hardware
   acceptance from builds, probes, CI, Docker, or dry-runs. Record physical
   results only when explicitly confirmed by the user, and name relevant skipped
   gates such as soak, performance, memory growth, or device-log inspection.
5. Search for stale scope names, descriptions, dates, versions, and contradictory
   claims. Run `git diff --check`, inspect the final docs diff, and run broader
   checks only when generation or the request requires them.

## Boundaries

- Use `$release` for release declarations, release-only compatibility evidence,
  readiness audits, or publishing.
- Do not commit or modify hosted state unless explicitly requested.

## Handoff

Report changed docs, recorded evidence, unverified gates, and `git diff --check`.

---
name: update-docs
description: Synchronize gopdsdk documentation after a feature, fix, or acceptance check is already implemented. Use for updating public API contracts, example-oriented README guidance, unreleased changelog entries, roadmap scope status, migration notes, and precisely scoped Simulator or physical-device evidence without changing code or invoking the implementation workflow. Do not use for release declaration, tagging, publishing, or release-readiness audits; use release instead.
---

# Update Docs

Follow `AGENTS.md`. Treat the implementation and its verification output as
source evidence; do not change code, generated files, tests, or native assets.

## Workflow

1. Inspect `git status`, the relevant diff, current documentation, and the
   exact verification or acceptance results. Preserve unrelated user edits.
2. Identify the public behavior that actually changed, its active roadmap
   scope, and the strongest verified evidence level. Ask only when a missing
   fact would materially change a claim.
3. Update only the applicable documents:
   - `API.md`: public contracts, ownership, errors, lifecycle, and capability
     discovery.
   - `README.md`: product- and example-oriented usage. Update an existing
     example description or add a focused example section for new capability.
   - `CHANGELOG.md`: unreleased behavior changes and durable acceptance
     evidence, including exact dates, tool versions, and measured artifact
     sizes when known.
   - `docs/ROADMAP.md`: active scope status and remaining work. Keep only
     future work after a release is declared.
   - `MIGRATING.md`: user action required by an intentional breaking change.
   - `COMPATIBILITY.md`: released host/target compatibility only; do not add
     unreleased per-scope status.
4. Keep evidence precise. Distinguish unit tests, external-consumer CLI,
   native CI, official SDK integration, Simulator execution, USB deployment,
   and user-confirmed physical-device behavior. Never promote builds, probes,
   CI, Docker, or dry-runs to visual or hardware acceptance.
5. Name skipped gates such as soak, performance, memory-growth measurement,
   or post-run device-log inspection. Record user-reported physical results
   only when the user explicitly confirms them.
6. Search for stale scope names, example descriptions, dates, versions, and
   contradictory readiness claims. Run `git diff --check` and inspect the final
   documentation diff. Run broader verification only when documentation
   generation or the user's request requires it.

## Boundaries

- Do not use `$implement` merely to edit documentation after completed work.
- Do not infer acceptance from a process starting or an artifact building.
- Do not declare a release or edit release-only compatibility evidence unless
  the user requested release work; hand that scope to `$release`.
- Do not commit, tag, push, publish, or modify hosted state unless explicitly
  requested.

## Handoff

Report the documents changed, the exact evidence recorded, all intentionally
unverified gates, and the result of `git diff --check`.

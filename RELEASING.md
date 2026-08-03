# Release procedure

The current candidate is `v0.5.0`, covering P5 on top of published `v0.4.0`.
Go module versions come from Git tags; no separate VERSION file is maintained.

## Candidate gates

1. Confirm the worktree contains only reviewed release changes and that no
   local or remote `v0.5.0` tag already exists.
2. Run formatting, `go test ./...`, `go vet ./...`, `git diff --check`, and
   `gopdsdk doctor` with workspace-local Go caches.
3. Observe green native CI on Windows, macOS, and Linux, including the external
   consumer and Linux race detector.
4. On the accepted Windows profile, run `doctor --probe`, then build and launch
   the P5 audio and microphone consumers in the official Simulator.
5. Build both consumers for conservative hard-float device, install through
   USB, and repeat sample/file playback, completion/fade, routing, synth,
   sequence/effect, microphone permission, stop/start, WAV save, and audible PCM
   playback acceptance. Record skipped denial/revocation scenarios explicitly.
6. Run the required physical regression soak and memory-growth checks. Inspect
   post-run crashlogs only when explicitly requested for that release run.
7. Review [API.md](API.md), [COMPATIBILITY.md](COMPATIBILITY.md),
   [MIGRATING.md](MIGRATING.md), README commands, public Go documentation, and
   [CHANGELOG.md](CHANGELOG.md). Confirm the API snapshot contains only intended
   P5 additions and no readiness claim exceeds its evidence.

## Version and consumer check

Before the tag exists, acceptance uses `require ... v0.5.0` with a local
`replace` to the candidate checkout. After publication, verify from a clean
directory without a `replace`:

```sh
go mod init example.com/release-check
go get github.com/Djunichi/gopdsdk@v0.5.0
go run github.com/Djunichi/gopdsdk/cmd/gopdsdk@v0.5.0 doctor
```

Then compile and dry-run a standalone application against both targets. This
post-tag check is mandatory because a local replacement cannot prove module
proxy availability or tag contents.

Creating a commit, tag, push, or hosted release is an explicit publishing
action and is not part of routine implementation verification.

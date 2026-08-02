# Release procedure

The current candidate is `v0.3.0`, covering P3 on top of `v0.2.0`. Go module
versions come from Git tags; no separate VERSION file is
maintained.

## Candidate gates

1. Confirm the worktree contains only reviewed release changes.
   Check `git show-ref --tags v0.3.0` as well: a stale local tag from the
   pre-publication history must not be reused or moved without explicit
   approval, even when the corresponding GitHub tag was deleted.
2. Run formatting, `go test ./...`, `go vet ./...`, `git diff --check`, and
   `gopdsdk doctor` with workspace-local Go caches.
3. Observe green native CI on Windows, macOS, and Linux, including the external
   consumer and Linux race detector.
4. On the accepted Windows profile, run `doctor --probe`, launch the external
   consumer in Simulator, then build/install/run the same package on Playdate.
5. Run the required 60-second physical regression soak and compare both device
   logs with their pre-run sizes and timestamps.
6. Review [API.md](API.md), [COMPATIBILITY.md](COMPATIBILITY.md), README commands,
   public Go documentation, and [CHANGELOG.md](CHANGELOG.md) for unsupported
   claims. Confirm the API snapshot changed only by the intended P3 additions.

## Version and consumer check

Before the tag exists, acceptance uses `require ... v0.3.0` with a local
`replace` to the candidate checkout. After publication, verify from a clean
directory without a `replace`:

```sh
go mod init example.com/release-check
go get github.com/Djunichi/gopdsdk@v0.3.0
go run github.com/Djunichi/gopdsdk/cmd/gopdsdk@v0.3.0 doctor
```

Then compile and dry-run a standalone application against both targets. This
post-tag check is mandatory because a local replacement cannot prove module
proxy availability or tag contents.

Creating a commit, tag, push, or hosted release is an explicit publishing
action and is not part of routine implementation verification.

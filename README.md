# PlayDateSDK

An independent Go SDK and toolchain for building Playdate applications.

The project is currently in **P0: foundation and feasibility**. No public API is
stable yet. The official Playdate C API is the normative source; third-party
projects, including pdgo, may be studied only as behavioral and product
references. Their implementation is not copied.

The initial target matrix is Windows, macOS, and Linux for both Simulator and
physical-device workflows. Public documentation will be added as the API takes
shape; design documents remain internal during P0.

## Environment diagnostics

Run the read-only environment check:

```sh
go run ./cmd/gopdsdk doctor
```

If the Playdate SDK is not in a conventional location, set the official
`PLAYDATE_SDK_PATH` environment variable or provide it explicitly:

```sh
go run ./cmd/gopdsdk doctor --sdk /path/to/PlaydateSDK
```

The command assesses SDK discovery, development, Simulator compilation, device
compilation, and device deployment separately. A tool reported as `UNVERIFIED`
was found but has not yet passed an end-to-end probe.

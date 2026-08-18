# JSON codec acceptance

This example reads packaged JSON through `playdate.FileSystem`, decodes it with
explicit byte/depth/node/string limits, validates a small schema, appends a
member, and encodes into a fixed 512-byte writer. It exercises the portable
`playdate/json` replacement without reflection, C callbacks, or deferred
cleanup.

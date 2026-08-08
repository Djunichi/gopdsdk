package store

type storeError string

func (message storeError) Error() string { return string(message) }

type migrationError struct{ cause error }

func (migrationError) Error() string             { return "stored value migration failed" }
func (migrationError) Unwrap() error             { return ErrMigration }
func (failure migrationError) Is(err error) bool { return err == ErrMigration || err == failure.cause }

var (
	// ErrConfig indicates an invalid store path, version, size bound, or migration table.
	ErrConfig error = storeError("invalid store configuration")
	// ErrTooLarge indicates that a payload exceeds the configured size bound.
	ErrTooLarge error = storeError("store payload is too large")
	// ErrCorrupt indicates a truncated, malformed, or checksum-invalid stored value.
	ErrCorrupt error = storeError("stored value is corrupt")
	// ErrFutureVersion indicates data written by a newer unsupported schema.
	ErrFutureVersion error = storeError("stored value has an unsupported future version")
	// ErrMigration indicates a missing or failed schema migration.
	ErrMigration error = storeError("stored value migration failed")
)

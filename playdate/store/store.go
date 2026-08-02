// Package store provides bounded, versioned persistence over the Playdate filesystem.
package store

import (
	"io"
	"path"
	"strings"

	"github.com/Djunichi/gopdsdk/playdate"
)

const (
	headerSize       = 16
	maximumStoreSize = 16 * 1024 * 1024
)

var magic = [4]byte{'G', 'P', 'D', 'S'}

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

// Migration upgrades a payload from one schema version to the next.
type Migration func(payload []byte) ([]byte, error)

// VersionMigration assigns one migration to its source schema version.
type VersionMigration struct {
	From    uint32
	Migrate Migration
}

// Config defines one versioned save or configuration value.
type Config struct {
	Path        string
	Version     uint32
	MaximumSize uint32
	Migrations  []VersionMigration
}

// Store atomically loads and replaces one versioned value.
type Store struct {
	files       playdate.FileSystem
	path        string
	temporary   string
	backup      string
	version     uint32
	maximumSize uint32
	migrations  []VersionMigration
}

// New validates config and creates a store. MaximumSize bounds payload bytes,
// excluding the store envelope.
func New(files playdate.FileSystem, config Config) (*Store, error) {
	if files == nil || !validPath(config.Path) || config.Version == 0 || config.MaximumSize == 0 || config.MaximumSize > maximumStoreSize {
		return nil, ErrConfig
	}
	migrations := append([]VersionMigration(nil), config.Migrations...)
	for index, migration := range migrations {
		if migration.From == 0 || migration.From >= config.Version || migration.Migrate == nil {
			return nil, ErrConfig
		}
		for previous := 0; previous < index; previous++ {
			if migrations[previous].From == migration.From {
				return nil, ErrConfig
			}
		}
	}
	return &Store{
		files:       files,
		path:        config.Path,
		temporary:   config.Path + ".tmp",
		backup:      config.Path + ".bak",
		version:     config.Version,
		maximumSize: config.MaximumSize,
		migrations:  migrations,
	}, nil
}

func validPath(filePath string) bool {
	if strings.IndexByte(filePath, 0) >= 0 || strings.Contains(filePath, "\\") || strings.HasPrefix(filePath, "/") {
		return false
	}
	cleaned := path.Clean(filePath)
	return cleaned != "." && cleaned != ".." && !strings.HasPrefix(cleaned, "../") && cleaned == filePath
}

// Save atomically replaces the stored value at the configured schema version.
func (store *Store) Save(payload []byte) error {
	if uint64(len(payload)) > uint64(store.maximumSize) {
		return ErrTooLarge
	}
	encoded := encode(store.version, payload)
	output, err := store.files.OpenFile(store.temporary, playdate.FileWrite)
	if err != nil {
		return err
	}
	written := 0
	for written < len(encoded) && err == nil {
		var count int
		count, err = output.Write(encoded[written:])
		written += count
		if err == nil && count == 0 {
			err = io.ErrShortWrite
		}
	}
	if err == nil {
		err = output.Flush()
	}
	if closeErr := output.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = store.files.Rename(store.temporary, store.path); err == nil {
		_ = store.files.Remove(store.backup, false)
		return nil
	}
	_ = store.files.Remove(store.backup, false)
	if backupErr := store.files.Rename(store.path, store.backup); backupErr != nil {
		return err
	}
	if publishErr := store.files.Rename(store.temporary, store.path); publishErr != nil {
		_ = store.files.Rename(store.backup, store.path)
		return publishErr
	}
	_ = store.files.Remove(store.backup, false)
	return nil
}

// Load reads, validates, and migrates the stored value. Successful migrations
// are atomically persisted before the upgraded payload is returned.
func (store *Store) Load() ([]byte, error) {
	info, err := store.files.Stat(store.path)
	if err != nil {
		originalErr := err
		if _, backupErr := store.files.Stat(store.backup); backupErr != nil {
			return nil, originalErr
		}
		if err = store.files.Rename(store.backup, store.path); err != nil {
			return nil, err
		}
		info, err = store.files.Stat(store.path)
		if err != nil {
			return nil, err
		}
	}
	if info.IsDir || info.Size < headerSize || info.Size > store.maximumSize+headerSize {
		return nil, ErrCorrupt
	}
	input, err := store.files.OpenFile(store.path, playdate.FileReadData)
	if err != nil {
		return nil, err
	}
	encoded := make([]byte, int(info.Size))
	read := 0
	for read < len(encoded) && err == nil {
		var count int
		count, err = input.Read(encoded[read:])
		read += count
		if count == 0 && err == nil {
			err = ErrCorrupt
		}
	}
	if err != nil && read < len(encoded) {
		if _, nativeFailure := err.(playdate.FileOperationError); !nativeFailure {
			err = ErrCorrupt
		}
	} else if read == len(encoded) {
		err = nil
	}
	if closeErr := input.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}
	version, payload, err := decode(encoded, store.maximumSize)
	if err != nil {
		return nil, err
	}
	if version > store.version {
		return nil, ErrFutureVersion
	}
	if version == store.version {
		_ = store.files.Remove(store.backup, false)
		return payload, nil
	}
	for version < store.version {
		var migration Migration
		for _, candidate := range store.migrations {
			if candidate.From == version {
				migration = candidate.Migrate
				break
			}
		}
		if migration == nil {
			return nil, ErrMigration
		}
		payload, err = migration(append([]byte(nil), payload...))
		if err != nil {
			return nil, migrationError{cause: err}
		}
		if uint64(len(payload)) > uint64(store.maximumSize) {
			return nil, migrationError{cause: ErrTooLarge}
		}
		version++
	}
	if err = store.Save(payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func encode(version uint32, payload []byte) []byte {
	encoded := make([]byte, headerSize+len(payload))
	copy(encoded[:4], magic[:])
	putUint32(encoded[4:8], version)
	putUint32(encoded[8:12], uint32(len(payload)))
	putUint32(encoded[12:16], checksum(payload))
	copy(encoded[headerSize:], payload)
	return encoded
}

func decode(encoded []byte, maximumSize uint32) (uint32, []byte, error) {
	if len(encoded) < headerSize || string(encoded[:4]) != string(magic[:]) {
		return 0, nil, ErrCorrupt
	}
	version := uint32At(encoded[4:8])
	size := uint32At(encoded[8:12])
	if version == 0 || size > maximumSize {
		return 0, nil, ErrCorrupt
	}
	if uint64(size)+headerSize != uint64(len(encoded)) {
		return 0, nil, ErrCorrupt
	}
	payload := append([]byte(nil), encoded[headerSize:]...)
	if uint32At(encoded[12:16]) != checksum(payload) {
		return 0, nil, ErrCorrupt
	}
	return version, payload, nil
}

func putUint32(target []byte, value uint32) {
	target[0] = byte(value)
	target[1] = byte(value >> 8)
	target[2] = byte(value >> 16)
	target[3] = byte(value >> 24)
}

func uint32At(source []byte) uint32 {
	return uint32(source[0]) |
		uint32(source[1])<<8 |
		uint32(source[2])<<16 |
		uint32(source[3])<<24
}

func checksum(payload []byte) uint32 {
	value := uint32(2166136261)
	for _, octet := range payload {
		value ^= uint32(octet)
		value *= 16777619
	}
	return value
}

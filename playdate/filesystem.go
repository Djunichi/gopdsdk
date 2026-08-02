package playdate

import "io"

// FileOptions selects the source or destination used by OpenFile.
type FileOptions uint8

const (
	// FileReadPackage reads only from the packaged PDX.
	FileReadPackage FileOptions = 1 << iota
	// FileReadData reads from the game's Data directory.
	FileReadData
	// FileWrite creates or truncates a file in the game's Data directory.
	FileWrite
	// FileAppend appends to a file in the game's Data directory.
	FileAppend FileOptions = 2 << 2
)

// FileSystem is the optional capability for Playdate files and directories.
type FileSystem interface {
	OpenFile(path string, options FileOptions) (File, error)
	Stat(path string) (FileInfo, error)
	List(path string, showHidden bool) ([]string, error)
	Mkdir(path string) error
	Remove(path string, recursive bool) error
	Rename(from, to string) error
}

// File is an owned Playdate file handle. Close releases the native handle.
type File interface {
	io.Reader
	io.Writer
	io.Seeker
	Flush() error
	Close() error
}

// FileInfo describes a Playdate filesystem entry.
type FileInfo struct {
	IsDir  bool
	Size   uint32
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Second int
}

package playdate

// Filesystem errors.

type fileError string

func (message fileError) Error() string { return string(message) }

// FileOperationError preserves the diagnostic returned by the Playdate API.
type FileOperationError struct {
	Operation string
	Path      string
	Message   string
}

func (failure FileOperationError) Error() string {
	if failure.Message == "" {
		return failure.Operation + " " + failure.Path + ": " + ErrFileIO.Error()
	}
	return failure.Operation + " " + failure.Path + ": " + ErrFileIO.Error() + ": " + failure.Message
}

func (FileOperationError) Unwrap() error { return ErrFileIO }

var (
	// ErrFileClosed indicates an operation on a closed owned file.
	ErrFileClosed error = fileError("file is closed")
	// ErrFilePath indicates a non-relative or parent-traversing path.
	ErrFilePath error = fileError("invalid Playdate file path")
	// ErrFileMode indicates an unsupported combination of open flags.
	ErrFileMode error = fileError("invalid Playdate file mode")
	// ErrFileOffset indicates a seek outside the native signed 32-bit range.
	ErrFileOffset error = fileError("file offset is outside the Playdate range")
	// ErrFileIO categorizes a failure reported by the Playdate filesystem.
	ErrFileIO error = fileError("Playdate filesystem operation failed")
	// ErrFileUnavailable indicates a context without the optional filesystem capability.
	ErrFileUnavailable error = fileError("filesystem capability is unavailable")
)

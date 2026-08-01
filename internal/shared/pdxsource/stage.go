// Package pdxsource stages application resources for the Playdate compiler.
package pdxsource

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Stage copies non-Go application resources from applicationDir into sourceDir.
// The caller writes the authoritative pdxinfo separately.
func Stage(applicationDir, sourceDir string) error {
	return filepath.WalkDir(applicationDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(applicationDir, path)
		if err != nil || relative == "." {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("stage PDX resources: symbolic link is unsupported: %s", relative)
		}
		if entry.IsDir() {
			return nil
		}
		if excluded(relative) {
			return nil
		}
		destination := filepath.Join(sourceDir, relative)
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return fmt.Errorf("stage PDX resource %s: %w", relative, err)
		}
		if err := copyFile(path, destination); err != nil {
			return fmt.Errorf("stage PDX resource %s: %w", relative, err)
		}
		return nil
	})
}

func excluded(relative string) bool {
	name := filepath.Base(relative)
	return strings.EqualFold(filepath.Ext(name), ".go") ||
		strings.EqualFold(name, "go.mod") || strings.EqualFold(name, "go.sum") ||
		strings.EqualFold(name, "pdxinfo")
}

func copyFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

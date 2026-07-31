// Package gomodule locates and renders metadata for repository-local probe modules.
package gomodule

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Info identifies the active gopdsdk module.
type Info struct {
	Root      string
	Path      string
	GoVersion string
}

// Locate reads the active module selected by the Go command.
func Locate(ctx context.Context) (Info, error) {
	output, err := exec.CommandContext(ctx, "go", "env", "GOMOD").CombinedOutput()
	if err != nil {
		return Info{}, fmt.Errorf("locate project module: %w: %s", err, strings.TrimSpace(string(output)))
	}
	goModPath := strings.TrimSpace(string(output))
	if goModPath == "" || goModPath == os.DevNull {
		return Info{}, fmt.Errorf("locate project module: command is not running inside a Go module")
	}
	contents, err := os.ReadFile(goModPath)
	if err != nil {
		return Info{}, fmt.Errorf("read project module: %w", err)
	}
	info := Info{Root: filepath.Dir(goModPath)}
	for _, line := range strings.Split(string(contents), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		switch fields[0] {
		case "module":
			info.Path = fields[1]
		case "go":
			info.GoVersion = fields[1]
		}
	}
	if info.Path == "" || info.GoVersion == "" {
		return Info{}, fmt.Errorf("read project module: module and go directives are required")
	}
	return info, nil
}

// RenderProbe returns a go.mod for a temporary child module.
func RenderProbe(info Info, suffix string) string {
	return fmt.Sprintf("module %s/%s\n\ngo %s\n\nrequire %s v0.0.0\n\nreplace %s => %s\n",
		info.Path, suffix, info.GoVersion, info.Path, info.Path,
		strconv.Quote(filepath.ToSlash(info.Root)))
}

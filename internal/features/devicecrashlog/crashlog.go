// Package devicecrashlog retrieves crash reports from a connected Playdate.
package devicecrashlog

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Read mounts the Playdate data disk and returns its crash log.
func Read(ctx context.Context, sdkPath string) ([]byte, string, error) {
	if sdkPath == "" {
		return nil, "", fmt.Errorf("Playdate SDK path is required")
	}
	sdkPath, err := filepath.Abs(filepath.Clean(sdkPath))
	if err != nil {
		return nil, "", fmt.Errorf("resolve Playdate SDK path: %w", err)
	}
	pdutilName := "pdutil"
	if runtime.GOOS == "windows" {
		pdutilName += ".exe"
	}
	pdutil := filepath.Join(sdkPath, "bin", pdutilName)
	if info, statErr := os.Stat(pdutil); statErr != nil || info.IsDir() {
		return nil, "", fmt.Errorf("required file %s is unavailable", pdutil)
	}
	if mountPath, ok := findMountedPlaydate(); ok {
		return readCrashlog(mountPath)
	}
	output, err := exec.CommandContext(ctx, pdutil, "datadisk").CombinedOutput()
	if err != nil {
		return nil, "", commandError(err, output)
	}
	mountPath, err := parseMountPath(string(output))
	if err != nil {
		return nil, "", err
	}
	return readCrashlog(mountPath)
}

func readCrashlog(mountPath string) ([]byte, string, error) {
	crashlogPath := filepath.Join(mountPath, "crashlog.txt")
	contents, err := os.ReadFile(crashlogPath)
	if err != nil {
		return nil, "", fmt.Errorf("read Playdate crash log %s: %w", crashlogPath, err)
	}
	return contents, crashlogPath, nil
}

func findMountedPlaydate() (string, bool) {
	for _, root := range mountCandidates() {
		if isPlaydateRoot(root) {
			return root, true
		}
	}
	return "", false
}

func mountCandidates() []string {
	switch runtime.GOOS {
	case "windows":
		roots := make([]string, 0, 24)
		for drive := 'C'; drive <= 'Z'; drive++ {
			roots = append(roots, fmt.Sprintf("%c:\\", drive))
		}
		return roots
	case "darwin":
		roots, _ := filepath.Glob("/Volumes/*")
		return roots
	default:
		var roots []string
		if user := os.Getenv("USER"); user != "" {
			for _, pattern := range []string{filepath.Join("/media", user, "*"), filepath.Join("/run/media", user, "*")} {
				matches, _ := filepath.Glob(pattern)
				roots = append(roots, matches...)
			}
		}
		matches, _ := filepath.Glob("/mnt/*")
		return append(roots, matches...)
	}
}

func isPlaydateRoot(root string) bool {
	for _, directory := range []string{"Data", "Games", "System"} {
		info, err := os.Stat(filepath.Join(root, directory))
		if err != nil || !info.IsDir() {
			return false
		}
	}
	info, err := os.Stat(filepath.Join(root, "crashlog.txt"))
	return err == nil && !info.IsDir()
}

func parseMountPath(output string) (string, error) {
	const marker = "Playdate data disk mounted as "
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if path, ok := strings.CutPrefix(line, marker); ok {
			path = strings.TrimSpace(path)
			if path == "" || !filepath.IsAbs(path) {
				return "", fmt.Errorf("pdutil returned invalid data disk path %q", path)
			}
			return filepath.Clean(path), nil
		}
	}
	return "", fmt.Errorf("pdutil did not report the mounted Playdate data disk: %s", strings.TrimSpace(output))
}

func commandError(err error, output []byte) error {
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return fmt.Errorf("mount Playdate data disk: %w", err)
	}
	return fmt.Errorf("mount Playdate data disk: %w: %s", err, detail)
}

// Package build compiles gopdsdk applications into Playdate artifacts.
package build

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/Djunichi/gopdsdk/internal/features/runtime/simabi"
)

const sdkModule = "github.com/Djunichi/gopdsdk"

// Config identifies a Simulator application build.
type Config struct {
	SDKPath string
	Package string
	Output  string
	Replace bool
}

// Result describes the produced Simulator artifact.
type Result struct {
	PackageImport string
	Output        string
}

type module struct {
	Path      string
	Dir       string
	GoVersion string
}

type packageInfo struct {
	ImportPath string
	Name       string
	Module     *module
}

// Simulator builds an importable Go package into a Playdate Simulator .pdx.
func Simulator(ctx context.Context, config Config) (Result, error) {
	if runtime.GOOS != "windows" {
		return Result{}, fmt.Errorf("Simulator build is not implemented for host %s", runtime.GOOS)
	}
	if config.SDKPath == "" {
		return Result{}, fmt.Errorf("Playdate SDK path is required")
	}
	if config.Package == "" {
		config.Package = "."
	}

	app, err := inspectPackage(ctx, config.Package)
	if err != nil {
		return Result{}, err
	}
	if app.Name == "main" {
		return Result{}, fmt.Errorf("inspect application package: %s must be importable, not package main", app.ImportPath)
	}
	if app.Module == nil {
		return Result{}, fmt.Errorf("inspect application package: %s is not in a Go module", app.ImportPath)
	}
	sdk, err := inspectModule(ctx, sdkModule)
	if err != nil {
		return Result{}, err
	}

	sdkPath, err := filepath.Abs(filepath.Clean(config.SDKPath))
	if err != nil {
		return Result{}, fmt.Errorf("resolve Playdate SDK path: %w", err)
	}
	pdc := filepath.Join(sdkPath, "bin", "pdc.exe")
	apiHeader := filepath.Join(sdkPath, "C_API", "pd_api.h")
	for _, path := range []string{pdc, apiHeader} {
		if info, statErr := os.Stat(path); statErr != nil || info.IsDir() {
			return Result{}, fmt.Errorf("required file %s is unavailable", path)
		}
	}
	gcc, err := exec.LookPath("gcc")
	if err != nil {
		return Result{}, fmt.Errorf("find gcc: %w", err)
	}

	output := config.Output
	if output == "" {
		output = filepath.Join("build", app.Name+".pdx")
	}
	output, err = filepath.Abs(filepath.Clean(output))
	if err != nil {
		return Result{}, fmt.Errorf("resolve output path: %w", err)
	}
	outputExists := false
	if info, statErr := os.Stat(output); statErr == nil {
		if !config.Replace {
			return Result{}, fmt.Errorf("output already exists: %s", output)
		}
		if !info.IsDir() {
			return Result{}, fmt.Errorf("output path is not a directory: %s", output)
		}
		outputExists = true
	} else if !os.IsNotExist(statErr) {
		return Result{}, fmt.Errorf("inspect output path: %w", statErr)
	}

	workDir, err := os.MkdirTemp("", "gopdsdk-build-")
	if err != nil {
		return Result{}, fmt.Errorf("create build directory: %w", err)
	}
	defer os.RemoveAll(workDir)
	sourceDir := filepath.Join(workDir, "Source")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create Source directory: %w", err)
	}

	sources, err := simabi.Render(simabi.Config{
		APIHeader:         apiHeader,
		PublicAPIImport:   sdkModule + "/playdate",
		RuntimeImport:     sdkModule + "/internal/features/runtime",
		ApplicationImport: app.ImportPath,
	})
	if err != nil {
		return Result{}, err
	}
	files := []struct {
		name     string
		contents string
	}{
		{name: "main.go", contents: sources.Go},
		{name: "bridge.c", contents: sources.C},
		{name: "go.mod", contents: renderGoMod(sdk, *app.Module)},
	}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(workDir, file.name), []byte(file.contents), 0o644); err != nil {
			return Result{}, fmt.Errorf("write %s: %w", file.name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "pdxinfo"), []byte(renderPDXInfo(app.Name)), 0o644); err != nil {
		return Result{}, fmt.Errorf("write pdxinfo: %w", err)
	}

	dllPath := filepath.Join(sourceDir, "pdex.dll")
	command := exec.CommandContext(ctx, "go", "build", "-buildmode=c-shared", "-o", dllPath, ".")
	command.Dir = workDir
	command.Env = append(os.Environ(), "CGO_ENABLED=1", "CC="+gcc, "CGO_CFLAGS=-DTARGET_EXTENSION=1 -DTARGET_SIMULATOR=1")
	if outputBytes, runErr := command.CombinedOutput(); runErr != nil {
		return Result{}, commandError("build Simulator library", runErr, outputBytes)
	}
	_ = os.Remove(filepath.Join(sourceDir, "pdex.h"))

	temporaryPDX := filepath.Join(workDir, "Application.pdx")
	command = exec.CommandContext(ctx, pdc, "-sdkpath", sdkPath, sourceDir, temporaryPDX)
	if outputBytes, runErr := command.CombinedOutput(); runErr != nil {
		return Result{}, commandError("package application", runErr, outputBytes)
	}
	if outputExists {
		if err := os.RemoveAll(output); err != nil {
			return Result{}, fmt.Errorf("replace output: %w", err)
		}
	}
	if err := copyDirectory(temporaryPDX, output); err != nil {
		_ = os.RemoveAll(output)
		return Result{}, fmt.Errorf("write output: %w", err)
	}
	return Result{PackageImport: app.ImportPath, Output: output}, nil
}

func inspectPackage(ctx context.Context, pattern string) (packageInfo, error) {
	output, err := exec.CommandContext(ctx, "go", "list", "-json", pattern).CombinedOutput()
	if err != nil {
		return packageInfo{}, commandError("inspect application package", err, output)
	}
	var info packageInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return packageInfo{}, fmt.Errorf("inspect application package: decode go list output: %w", err)
	}
	return info, nil
}

func inspectModule(ctx context.Context, path string) (module, error) {
	output, err := exec.CommandContext(ctx, "go", "list", "-m", "-json", path).CombinedOutput()
	if err != nil {
		return module{}, commandError("inspect gopdsdk module", err, output)
	}
	var info module
	if err := json.Unmarshal(output, &info); err != nil {
		return module{}, fmt.Errorf("inspect gopdsdk module: decode go list output: %w", err)
	}
	return info, nil
}

func renderGoMod(sdk, app module) string {
	goVersion := sdk.GoVersion
	if goVersion == "" {
		goVersion = app.GoVersion
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "module %s/build\n\ngo %s\n\nrequire %s v0.0.0\n", sdkModule, goVersion, sdkModule)
	if app.Path != sdk.Path {
		fmt.Fprintf(&builder, "require %s v0.0.0\n", app.Path)
	}
	fmt.Fprintf(&builder, "\nreplace %s => %s\n", sdkModule, strconv.Quote(filepath.ToSlash(sdk.Dir)))
	if app.Path != sdk.Path {
		fmt.Fprintf(&builder, "replace %s => %s\n", app.Path, strconv.Quote(filepath.ToSlash(app.Dir)))
	}
	return builder.String()
}

func renderPDXInfo(name string) string {
	return fmt.Sprintf("name=%s\nauthor=gopdsdk\nbundleID=sdk.gopdsdk.%s\nversion=0.0.0\nbuildNumber=1\n", name, name)
}

func commandError(action string, err error, output []byte) error {
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return fmt.Errorf("%s: %w", action, err)
	}
	return fmt.Errorf("%s: %w: %s", action, err, detail)
}

func copyDirectory(source, target string) error {
	if err := os.MkdirAll(target, 0o755); err != nil {
		return err
	}
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil || relative == "." {
			return err
		}
		destination := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		return copyFile(path, destination)
	})
}

func copyFile(source, target string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
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

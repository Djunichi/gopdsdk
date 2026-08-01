package initproject

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strings"
)

// Run executes the init command.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("gopdsdk init", flag.ContinueOnError)
	flags.SetOutput(stderr)
	modulePath := flags.String("module", "", "Go module path")
	name := flags.String("name", "", "display name")
	author := flags.String("author", "", "author name")
	bundleID := flags.String("bundle-id", "", "reverse-DNS Playdate bundle ID")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return fmt.Errorf("expected one project directory")
	}
	sdkDir, sdkVersion, err := discoverSDKDependency(ctx)
	if err != nil {
		return err
	}
	result, err := Create(Config{
		Path:       flags.Arg(0),
		Module:     *modulePath,
		Name:       *name,
		Author:     *author,
		BundleID:   *bundleID,
		SDKDir:     sdkDir,
		SDKVersion: sdkVersion,
		GoVersion:  strings.TrimPrefix(runtime.Version(), "go"),
	})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "Created %s\nModule: %s\n", result.Path, result.Module)
	return err
}

func discoverSDKDependency(ctx context.Context) (directory, version string, err error) {
	if info, ok := debug.ReadBuildInfo(); ok {
		if version := publishedSDKVersion(info); version != "" {
			return "", version, nil
		}
	}
	command := exec.CommandContext(ctx, "go", "list", "-m", "-f", "{{.Dir}}", sdkModule)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("locate local gopdsdk module: %w: %s", err, strings.TrimSpace(string(output)))
	}
	directory = strings.TrimSpace(string(output))
	if directory == "" {
		return "", "", fmt.Errorf("locate local gopdsdk module: module directory is empty")
	}
	return directory, "", nil
}

func publishedSDKVersion(info *debug.BuildInfo) string {
	if info.Main.Path != sdkModule || !releaseVersionPattern.MatchString(info.Main.Version) {
		return ""
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.modified" && setting.Value == "true" {
			return ""
		}
	}
	return info.Main.Version
}

package deviceprobe

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Run executes the device probe command.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) < 2 || args[1] != "device" || (args[0] != "probe" && args[0] != "run") {
		return fmt.Errorf("expected \"gopdsdk probe device\" or \"gopdsdk run device\"")
	}
	runDevice := args[0] == "run"
	commandName := "gopdsdk probe device"
	if runDevice {
		commandName = "gopdsdk run device"
	}
	flags := flag.NewFlagSet(commandName, flag.ContinueOnError)
	flags.SetOutput(stderr)
	sdkPath := flags.String("sdk", os.Getenv("PLAYDATE_SDK_PATH"), "path to the Playdate SDK")
	install := flags.Bool("install", false, "install the verified probe package on a connected Playdate")
	artifactsDir := flags.String("artifacts", "", "directory for diagnostic build artifacts")
	if err := flags.Parse(args[2:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}
	if *sdkPath == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			*sdkPath = filepath.Join(home, "Documents", "PlaydateSDK")
		}
	}
	result, err := Probe(ctx, Config{SDKPath: *sdkPath, Install: *install || runDevice, Run: runDevice, ArtifactsDir: *artifactsDir})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "Device package stage: READY\nTinyGo:              %s\nCompiler:            %s\nELF:                 %s\nExport:              %s\nPackage:             %s\nDeployment:          %s\nExecution:           %s\nStill unverified:    %s\n",
		result.TinyGo, result.GCC, result.Format, result.Export, result.Package, result.Deploy, result.Run, result.Pending)
	return err
}

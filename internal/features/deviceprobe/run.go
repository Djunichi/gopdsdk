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
	if len(args) < 2 || args[0] != "probe" || args[1] != "device" {
		return fmt.Errorf("expected \"gopdsdk probe device\"")
	}
	flags := flag.NewFlagSet("gopdsdk probe device", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sdkPath := flags.String("sdk", os.Getenv("PLAYDATE_SDK_PATH"), "path to the Playdate SDK")
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
	result, err := Probe(ctx, Config{SDKPath: *sdkPath})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "Device package stage: READY\nTinyGo:              %s\nCompiler:            %s\nELF:                 %s\nExport:              %s\nPackage:             %s\nStill unverified:    %s\n",
		result.TinyGo, result.GCC, result.Format, result.Export, result.Package, result.Pending)
	return err
}

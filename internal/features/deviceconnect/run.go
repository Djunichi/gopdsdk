package deviceconnect

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Run executes the read-only device connection probe command.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) < 2 || args[0] != "probe" || args[1] != "connection" {
		return fmt.Errorf("expected \"gopdsdk probe connection\"")
	}
	flags := flag.NewFlagSet("gopdsdk probe connection", flag.ContinueOnError)
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
	_, err = fmt.Fprintf(stdout, "Device connection: READY\npdutil:           %s\nEvidence:         %s\n", result.Tool, result.Status)
	return err
}

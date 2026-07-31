package devicecrashlog

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Run executes the crashlog command.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] != "crashlog" {
		return fmt.Errorf("expected \"gopdsdk crashlog\"")
	}
	flags := flag.NewFlagSet("gopdsdk crashlog", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sdkPath := flags.String("sdk", os.Getenv("PLAYDATE_SDK_PATH"), "path to the Playdate SDK")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}
	if *sdkPath == "" {
		if home, err := os.UserHomeDir(); err == nil {
			*sdkPath = filepath.Join(home, "Documents", "PlaydateSDK")
		}
	}
	contents, path, err := Read(ctx, *sdkPath)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(stderr, "Crash log: %s\n", path); err != nil {
		return err
	}
	_, err = stdout.Write(contents)
	return err
}

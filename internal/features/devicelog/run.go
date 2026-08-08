package devicelog

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Run executes the crashlog or errorlog command.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("expected \"gopdsdk crashlog\" or \"gopdsdk errorlog\"")
	}
	kind, ok := commandKind(args[0])
	if !ok {
		return fmt.Errorf("expected \"gopdsdk crashlog\" or \"gopdsdk errorlog\"")
	}
	flags := flag.NewFlagSet("gopdsdk "+args[0], flag.ContinueOnError)
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
	contents, path, err := Read(ctx, *sdkPath, kind)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(stderr, "%s: %s\n", displayLabel(kind), path); err != nil {
		return err
	}
	_, err = stdout.Write(contents)
	return err
}

func commandKind(command string) (Kind, bool) {
	switch command {
	case "crashlog":
		return CrashLog, true
	case "errorlog":
		return ErrorLog, true
	default:
		return "", false
	}
}

func displayLabel(kind Kind) string {
	if kind == ErrorLog {
		return "Error log"
	}
	return "Crash log"
}

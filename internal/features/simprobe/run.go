package simprobe

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

// Run executes the simulator probe command.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) < 2 || args[0] != "probe" || args[1] != "simulator" {
		return fmt.Errorf("expected \"gopdsdk probe simulator\"")
	}

	flags := flag.NewFlagSet("gopdsdk probe simulator", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sdkPath := flags.String("sdk", os.Getenv("PLAYDATE_SDK_PATH"), "path to the Playdate SDK")
	runSimulator := flags.Bool("run", false, "launch Simulator and verify kEventInit")
	timeout := flags.Duration("timeout", 15*time.Second, "maximum time to wait for Simulator initialization")
	if err := flags.Parse(args[2:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}
	if *sdkPath == "" {
		return fmt.Errorf("playdate SDK path is required; set PLAYDATE_SDK_PATH or pass --sdk")
	}

	result, err := Probe(ctx, Config{SDKPath: *sdkPath, RunSimulator: *runSimulator, Timeout: *timeout})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "Simulator probe: READY\nSDK:             %s\nCompiler:        %s\nExport:          %s\nPackage:         %s\n",
		result.SDKVersion, result.Compiler, result.Export, result.Package)
	if err == nil && result.Event != "" {
		_, err = fmt.Fprintf(stdout, "Simulator event: %s\n", result.Event)
	}
	return err
}

package doctor

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
)

// Options supplies capability probes at the application composition boundary.
type Options struct {
	SimulatorProbe func(context.Context, string) error
}

// Run executes the doctor subcommand CLI.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer, options Options) error {
	if len(args) == 0 {
		return fmt.Errorf("expected a command (try \"gopdsdk doctor\")")
	}
	if args[0] != "doctor" {
		return fmt.Errorf("unknown command %q (try \"gopdsdk doctor\")", args[0])
	}

	flags := flag.NewFlagSet("gopdsdk doctor", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sdkPath := flags.String("sdk", "", "path to the Playdate SDK")
	probe := flags.Bool("probe", false, "run build probes for discovered toolchains")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}

	report, err := Inspect(ctx, Config{SDKPath: *sdkPath})
	if err != nil {
		return err
	}
	if *probe {
		runSimulatorProbe(ctx, &report, options.SimulatorProbe)
	}
	return writeReport(stdout, report)
}

func runSimulatorProbe(ctx context.Context, report *Report, probe func(context.Context, string) error) {
	for index := range report.Capabilities {
		capability := &report.Capabilities[index]
		if capability.Name != "simulator" || capability.Status != StatusUnverified {
			continue
		}
		if probe == nil {
			capability.Status = StatusIncompatible
			capability.Summary = "simulator probe is not configured"
			return
		}
		if err := probe(ctx, report.SDKPath); err != nil {
			capability.Status = StatusIncompatible
			capability.Summary = err.Error()
			return
		}
		capability.Status = StatusReady
		capability.Summary = "c-shared build, eventHandler export, and .pdx packaging verified"
		return
	}
}

func writeReport(out io.Writer, report Report) error {
	if _, err := fmt.Fprintf(out, "Host:         %s\n", report.Host); err != nil {
		return err
	}
	if report.SDKPath == "" {
		if _, err := fmt.Fprintln(out, "Playdate SDK: not found"); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(out, "Playdate SDK: %s\nSDK path:     %s\n", report.SDKVersion, report.SDKPath); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(out, "\nCapabilities:"); err != nil {
		return err
	}
	for _, capability := range report.Capabilities {
		if _, err := fmt.Fprintf(out, "  %-14s %-12s %s\n", capability.Name, strings.ToUpper(string(capability.Status)), capability.Summary); err != nil {
			return err
		}
	}
	return nil
}

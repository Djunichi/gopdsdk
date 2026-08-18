package deviceprobe

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
)

// RunProfileAudit executes the isolated device Go-profile compile audit.
func RunProfileAudit(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	if len(args) < 2 || args[0] != "probe" || args[1] != "device-profile" {
		return fmt.Errorf("expected \"gopdsdk probe device-profile\"")
	}
	flags := flag.NewFlagSet("gopdsdk probe device-profile", flag.ContinueOnError)
	flags.SetOutput(stderr)
	jsonOutput := flags.Bool("json", false, "write the machine-readable audit matrix")
	if err := flags.Parse(args[2:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}
	results, err := AuditProfile(ctx)
	if err != nil {
		return err
	}
	if *jsonOutput {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(results)
	}
	if _, err := fmt.Fprintln(stdout, "Device Go profile compile audit"); err != nil {
		return err
	}
	for _, result := range results {
		if _, err := fmt.Fprintf(stdout, "%-16s %-10s %s\n", result.Name, result.Status, result.Evidence); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(stdout, "Evidence: compile and linked-symbol inspection only; Simulator and physical-device execution unverified")
	return err
}

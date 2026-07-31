package deviceprobe

import (
	"context"
	"fmt"
	"io"
)

// Run executes the device probe command.
func Run(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) != 2 || args[0] != "probe" || args[1] != "device" {
		return fmt.Errorf("expected \"gopdsdk probe device\"")
	}
	result, err := Probe(ctx)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "Device compile stage: READY\nTinyGo:              %s\nCompiler:            %s\nObject:              %s\nExport:              %s\nStill unverified:    %s\n",
		result.TinyGo, result.GCC, result.Format, result.Export, result.Pending)
	return err
}

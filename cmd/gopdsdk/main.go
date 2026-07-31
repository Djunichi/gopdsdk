package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Djunichi/gopdsdk/internal/features/doctor"
	"github.com/Djunichi/gopdsdk/internal/features/simprobe"
)

func main() {
	args := os.Args[1:]
	var err error
	if len(args) == 0 {
		err = fmt.Errorf("expected a command (try \"gopdsdk doctor\")")
	} else {
		switch args[0] {
		case "doctor":
			err = doctor.Run(context.Background(), args, os.Stdout, os.Stderr, doctor.Options{
				SimulatorProbe: func(ctx context.Context, sdkPath string) error {
					_, probeErr := simprobe.Probe(ctx, simprobe.Config{SDKPath: sdkPath})
					return probeErr
				},
			})
		case "probe":
			err = simprobe.Run(context.Background(), args, os.Stdout, os.Stderr)
		default:
			err = fmt.Errorf("unknown command %q", args[0])
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "gopdsdk:", err)
		os.Exit(2)
	}
}

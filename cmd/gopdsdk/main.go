package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Djunichi/gopdsdk/internal/features/doctor"
)

func main() {
	if err := doctor.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "gopdsdk:", err)
		os.Exit(2)
	}
}

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"ceoharness/internal/eval"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := eval.RunCLI(ctx, os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

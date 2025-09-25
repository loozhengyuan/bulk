package main

import (
	"fmt"
	"os"

	"github.com/loozhengyuan/bulk/internal/cmd"
)

func run() error {
	cli, err := cmd.New()
	if err != nil {
		return err
	}
	if err := cli.Execute(); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

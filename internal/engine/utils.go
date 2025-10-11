package engine

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func dirExecContext(ctx context.Context, dir, cmd string, args ...string) error {
	// TODO: Should we not pipe stdout/err?
	c := exec.CommandContext(ctx, cmd, args...)
	c.Dir = dir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("run command: %w", err)
	}
	return nil
}

func dirExec(dir, cmd string, args ...string) error {
	return dirExecContext(context.Background(), dir, cmd, args...)
}

func promptConfirm(prompt string) (bool, error) {
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [y/N]: ", prompt)

		input, err := r.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("read input: %w", err)
		}

		input = strings.TrimSpace(strings.ToLower(input))
		switch input {
		case "y", "yes":
			return true, nil
		case "n", "no", "":
			return false, nil
		}
	}
}

package engine

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func execCommand(dir, cmd string, args ...string) error {
	// TODO: Implement CommandContext?
	// TODO: Should we not pipe stdout/err?
	c := exec.Command(cmd, args...)
	c.Dir = dir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("run command: %w", err)
	}
	return nil
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

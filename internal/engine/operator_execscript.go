package engine

import (
	"fmt"
	"os"
)

func execScript(dir, script string) error {
	// TODO: Support running in container?
	f, err := os.CreateTemp("", "script-*.sh")
	if err != nil {
		return fmt.Errorf("create temp script file: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(script); err != nil {
		return fmt.Errorf("write script to file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close script file: %w", err)
	}

	// TODO: Support other shells?
	return dirExec(dir, "bash", "-euo", "pipefail", f.Name())
}

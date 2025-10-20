package engine

import (
	"fmt"
	"os"
)

var _ Operator = (*OperatorExecScript)(nil)

type OperatorExecScript struct {
	Run string `json:"run"`
}

func (op *OperatorExecScript) Validate() error {
	if op.Run == "" {
		return fmt.Errorf("run is not specified")
	}
	return nil
}

func (op *OperatorExecScript) Apply(ctx OperatorContext) error {
	// TODO: Support running in container?
	f, err := os.CreateTemp("", "script-*.sh")
	if err != nil {
		return fmt.Errorf("create temp script file: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(op.Run); err != nil {
		return fmt.Errorf("write script to file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close script file: %w", err)
	}

	// TODO: Support other shells?
	return dirExec(ctx.Dir, "bash", "-euo", "pipefail", f.Name())
}

package apply

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/loozhengyuan/bulk/internal/engine"
)

type options struct {
	// TODO: Verbose or Debug mode? Or both?
	// TODO: Do we need `--dry-run` and/or `--interactive` flags?
	force bool
}

func New() *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Applies configuration onto repositories.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := engine.NewFromFile(args[0])
			if err != nil {
				return fmt.Errorf("create engine: %w", err)
			}
			e.SetForce(opts.force)
			if err := e.Execute(); err != nil {
				return fmt.Errorf("execute plan: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "skips any interactive prompts")
	return cmd
}

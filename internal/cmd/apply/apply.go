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
	key   string
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
			e.SetKey(opts.key)
			if err := e.Execute(); err != nil {
				return fmt.Errorf("execute plan: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "skips any interactive prompts")
	cmd.Flags().StringVarP(&opts.key, "key", "k", "", "override the default id key")
	return cmd
}

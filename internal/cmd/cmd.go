package cmd

import (
	"github.com/spf13/cobra"

	"github.com/loozhengyuan/bulk/internal/cmd/apply"
	"github.com/loozhengyuan/bulk/internal/cmd/version"
)

func New() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:           "bulk",
		Short:         "Apply bulk changes across repositories.",
		SilenceErrors: true,
		SilenceUsage:  true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	cmd.AddCommand(apply.New())
	cmd.AddCommand(version.New())
	return cmd, nil
}

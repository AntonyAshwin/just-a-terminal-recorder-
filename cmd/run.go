package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"jtr/internal/executor"
	"jtr/internal/storage"
)

var runCmd = &cobra.Command{
	Use:   "run <name>",
	Short: "Execute a stored recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !storage.ValidName(name) {
			return fmt.Errorf("invalid recording name %q", name)
		}
		return executor.RunScript(name)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

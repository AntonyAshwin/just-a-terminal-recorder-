package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"jtr/internal/storage"
)

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Open a recording in your editor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !storage.ValidName(name) {
			return fmt.Errorf("invalid recording name %q", name)
		}
		return storage.EditScript(name)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"jtr/internal/storage"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !storage.ValidName(name) {
			return fmt.Errorf("invalid recording name %q", name)
		}
		return storage.DeleteScript(name)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

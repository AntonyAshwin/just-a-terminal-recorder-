package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"jtr/internal/storage"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all recordings",
	RunE: func(cmd *cobra.Command, args []string) error {
		items, err := storage.ListScripts()
		if err != nil {
			return err
		}
		for _, item := range items {
			fmt.Println(item)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"jtr/internal/storage"
)

var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print REC scripts directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := storage.EnsureBinDir()
		if err != nil {
			return err
		}
		fmt.Println(dir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pathCmd)
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"jtr/internal/storage"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <name>",
	Short: "Show a recording without running it",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !storage.ValidName(name) {
			return fmt.Errorf("invalid recording name %q", name)
		}

		content, err := storage.ReadScript(name)
		if err != nil {
			return err
		}

		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#!") {
				continue
			}
			fmt.Println(line)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}

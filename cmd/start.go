package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"jtr/internal/history"
	"jtr/internal/shell"
	"jtr/internal/storage"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a recording session",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		exists, err := storage.ActiveSessionExists()
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("an active session already exists; run: rec stop <name>")
		}

		shellType := shell.DetectShell()
		baseline, err := history.ReadLastN(shellType, 5000)
		if err != nil {
			return fmt.Errorf("failed to start session: unable to read history: %w", err)
		}

		startDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to start session: unable to detect current directory: %w", err)
		}

		if err := storage.CreateActiveSession(baseline, startDir); err != nil {
			return err
		}

		cmd.Println("started recording session")
		cmd.Println("run your terminal commands, then execute: rec stop <name>")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

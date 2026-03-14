package cmd

import "github.com/spf13/cobra"

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "rec",
	Short: "Just a Terminal Recorder",
	Long:  "REC records recent shell commands into reusable scripts.",
}

func Execute() error {
	return rootCmd.Execute()
}

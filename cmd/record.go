package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"jtr/internal/history"
	"jtr/internal/recorder"
	"jtr/internal/shell"
	"jtr/internal/storage"
)

var (
	recordIncludePath bool
	recordIgnorePath  bool
	recordAbsolute    bool
	recordDryRun      bool
)

var recordCmd = &cobra.Command{
	Use:   "record [n] <name>",
	Short: "Record last shell commands as a reusable script",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 || len(args) > 2 {
			return fmt.Errorf("usage: rec record [n] <name>")
		}
		if len(args) == 2 {
			if _, err := strconv.Atoi(args[0]); err != nil {
				return fmt.Errorf("invalid n value %q; expected integer", args[0])
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if recordIncludePath && recordIgnorePath {
			return fmt.Errorf("--include-path and --ignore-path cannot be used together")
		}

		n := 1
		name := args[0]
		if len(args) == 2 {
			parsedN, _ := strconv.Atoi(args[0])
			if parsedN <= 0 {
				return fmt.Errorf("n must be greater than 0")
			}
			n = parsedN
			name = args[1]
		}

		if !storage.ValidName(name) {
			return fmt.Errorf("invalid recording name %q", name)
		}

		shellType := shell.DetectShell()
		commands, err := history.ReadLastN(shellType, n+5)
		if err != nil {
			return err
		}

		processed := recorder.Process(recorder.Options{
			Commands:      commands,
			Name:          name,
			Count:         n,
			IgnorePath:    recordIgnorePath,
			AbsolutePaths: recordAbsolute,
			IncludePath:   recordIncludePath,
			CurrentDir:    mustGetwd(),
		})

		if len(processed) == 0 {
			return fmt.Errorf("no commands available to record")
		}

		if recordDryRun {
			for _, line := range processed {
				fmt.Println(line)
			}
			return nil
		}

		exists, err := storage.ScriptExists(name)
		if err != nil {
			return err
		}
		if exists {
			ok, err := storage.ConfirmOverwrite(name)
			if err != nil {
				return err
			}
			if !ok {
				fmt.Println("record canceled")
				return nil
			}
		}

		path, err := storage.SaveScript(name, processed)
		if err != nil {
			return err
		}

		fmt.Printf("recorded %d command(s) as %q\n", len(processed), name)
		fmt.Printf("script: %s\n", path)
		return nil
	},
}

func init() {
	recordCmd.Flags().BoolVar(&recordIncludePath, "include-path", false, "Preserve path-changing commands like cd")
	recordCmd.Flags().BoolVar(&recordIgnorePath, "ignore-path", false, "Remove path-changing commands like cd")
	recordCmd.Flags().BoolVar(&recordIgnorePath, "ip", false, "Alias for --ignore-path")
	recordCmd.Flags().BoolVar(&recordAbsolute, "absolute-paths", false, "Convert relative cd paths to absolute paths")
	recordCmd.Flags().BoolVar(&recordDryRun, "dry-run", false, "Print commands that would be recorded without saving")

	rootCmd.AddCommand(recordCmd)
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}

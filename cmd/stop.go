package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"jtr/internal/history"
	"jtr/internal/shell"
	"jtr/internal/storage"
)

var stopIgnorePath bool

var stopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop active session and save it as a recording",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !storage.ValidName(name) {
			return fmt.Errorf("invalid recording name %q", name)
		}

		exists, err := storage.ActiveSessionExists()
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("no active session; run: rec start")
		}

		session, err := storage.ReadActiveSession()
		if err != nil {
			return err
		}

		shellType := shell.DetectShell()
		all, err := history.ReadLastN(shellType, 5000)
		if err != nil {
			return err
		}

		newLines, reliable := diffAfterBaseline(session.Baseline, all)
		if !reliable {
			fmt.Println("unable to safely detect commands since start")
			fmt.Println("session remains active; try running 'rec stop <name>' again after 'fc -W'")
			fmt.Println("if it persists, restart with: rec start")
			return nil
		}
		captured := make([]string, 0, len(newLines))
		for _, line := range newLines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if isJTRCommand(trimmed) {
				continue
			}
			if stopIgnorePath && isCDCommand(trimmed) {
				continue
			}
			captured = append(captured, trimmed)
		}

		if !stopIgnorePath && strings.TrimSpace(session.StartDir) != "" {
			captured = prependStartDir(captured, session.StartDir)
		}

		if len(captured) == 0 {
			fmt.Printf("no commands captured for session %q yet\n", name)
			fmt.Println("session remains active; run more commands and retry stop")
			fmt.Println("for zsh, enable: setopt INC_APPEND_HISTORY SHARE_HISTORY")
			return nil
		}

		scriptExists, err := storage.ScriptExists(name)
		if err != nil {
			return err
		}
		if scriptExists {
			ok, err := storage.ConfirmOverwrite(name)
			if err != nil {
				return err
			}
			if !ok {
				fmt.Println("stop canceled")
				return nil
			}
		}

		path, err := storage.SaveScript(name, captured)
		if err != nil {
			return err
		}

		if err := storage.DeleteActiveSession(); err != nil {
			return err
		}

		fmt.Printf("saved %d command(s) to %q\n", len(captured), name)
		fmt.Printf("script: %s\n", path)
		return nil
	},
}

func init() {
	stopCmd.Flags().BoolVarP(&stopIgnorePath, "ignore-path", "i", false, "Do not include path context or cd commands")
	stopCmd.Flags().BoolVar(&stopIgnorePath, "ip", false, "Alias for --ignore-path")
	rootCmd.AddCommand(stopCmd)
}

func isJTRCommand(line string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}
	first := filepath.Base(fields[0])
	return first == "rec" || first == "jtr"
}

func isCDCommand(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "cd" || strings.HasPrefix(trimmed, "cd ")
}

func prependStartDir(commands []string, startDir string) []string {
	if len(commands) == 0 {
		return commands
	}
	cdLine := "cd " + shellSingleQuote(startDir)
	out := make([]string, 0, len(commands)+1)
	out = append(out, cdLine)
	out = append(out, commands...)
	return out
}

func shellSingleQuote(input string) string {
	if input == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(input, "'", "'\"'\"'") + "'"
}

func diffAfterBaseline(baseline, current []string) ([]string, bool) {
	if len(current) == 0 {
		return []string{}, true
	}
	if len(baseline) == 0 {
		return []string{}, false
	}

	max := len(baseline)
	if len(current) < max {
		max = len(current)
	}

	// Strategy 1: rolling-window overlap (baseline tail == current head).
	for k := max; k >= 1; k-- {
		baseTail := baseline[len(baseline)-k:]
		currHead := current[:k]
		if equalSlices(baseTail, currHead) {
			return current[k:], true
		}
	}

	// Strategy 2: find the latest occurrence of the longest baseline suffix
	// anywhere in current history, then take everything after that match.
	bestStart := -1
	bestLen := 0
	for k := max; k >= 1; k-- {
		target := baseline[len(baseline)-k:]
		for i := len(current) - k; i >= 0; i-- {
			if equalSlices(target, current[i:i+k]) {
				bestStart = i
				bestLen = k
				break
			}
		}
		if bestStart >= 0 {
			break
		}
	}

	if bestStart >= 0 {
		return current[bestStart+bestLen:], true
	}

	return []string{}, false
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

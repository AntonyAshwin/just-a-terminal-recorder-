package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"jtr/internal/shell"
	"jtr/internal/storage"
)

const (
	setupStartMarker = "# >>> rec setup >>>"
	setupEndMarker   = "# <<< rec setup <<<"
)

var (
	setupShell   string
	setupProfile string
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure shell integration automatically",
	Long:  "Configure PATH and shell integration in your shell profile so recordings can be executed as path-aware shell functions.",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetShell := strings.TrimSpace(strings.ToLower(setupShell))
		if targetShell == "" {
			targetShell = string(shell.DetectShell())
		}

		switch targetShell {
		case "zsh", "bash":
		default:
			return fmt.Errorf("unsupported shell %q (expected zsh or bash)", targetShell)
		}

		if _, err := storage.EnsureBinDir(); err != nil {
			return err
		}

		profilePath := strings.TrimSpace(setupProfile)
		if profilePath == "" {
			var err error
			profilePath, err = defaultProfileForShell(targetShell)
			if err != nil {
				return err
			}
		}

		execName := filepath.Base(os.Args[0])
		if execName == "" {
			execName = "rec"
		}

		block := buildSetupBlock(execName, targetShell)

		raw, err := os.ReadFile(profilePath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to read profile %q: %w", profilePath, err)
		}

		updated, changed := upsertSetupBlock(string(raw), block)
		if !changed {
			fmt.Printf("setup is already configured in %s\n", profilePath)
			fmt.Printf("reload your shell: source %s\n", profilePath)
			return nil
		}

		if err := os.WriteFile(profilePath, []byte(updated), 0o644); err != nil {
			return fmt.Errorf("failed to update profile %q: %w", profilePath, err)
		}

		fmt.Printf("updated %s\n", profilePath)
		fmt.Printf("reload your shell: source %s\n", profilePath)
		fmt.Println("new recordings can be loaded with: rec_load_recordings")

		return nil
	},
}

func init() {
	setupCmd.Flags().StringVar(&setupShell, "shell", "", "Shell to configure (zsh or bash)")
	setupCmd.Flags().StringVar(&setupProfile, "profile", "", "Profile file to update (defaults to ~/.zshrc or ~/.bashrc)")
	rootCmd.AddCommand(setupCmd)
}

func defaultProfileForShell(shellName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to detect home directory: %w", err)
	}

	switch shellName {
	case "zsh":
		return filepath.Join(home, ".zshrc"), nil
	case "bash":
		return filepath.Join(home, ".bashrc"), nil
	default:
		return "", fmt.Errorf("unsupported shell %q", shellName)
	}
}

func buildSetupBlock(execName, shellName string) string {
	return fmt.Sprintf(`%s
export PATH="$HOME/.jtr/bin:$PATH"
eval "$(command %s init %s)"
%s
`, setupStartMarker, execName, shellName, setupEndMarker)
}

func upsertSetupBlock(content, block string) (string, bool) {
	start := strings.Index(content, setupStartMarker)
	end := strings.Index(content, setupEndMarker)

	if start >= 0 && end >= start {
		end += len(setupEndMarker)
		for end < len(content) && content[end] == '\n' {
			end++
		}
		replaced := content[:start] + block + content[end:]
		return replaced, replaced != content
	}

	trimmed := strings.TrimRight(content, "\n")
	if trimmed == "" {
		return block, true
	}

	return trimmed + "\n\n" + block, true
}

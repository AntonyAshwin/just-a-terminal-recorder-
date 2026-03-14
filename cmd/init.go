package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"jtr/internal/shell"
)

var initShell string

var initCmd = &cobra.Command{
	Use:   "init [zsh|bash]",
	Short: "Print shell integration for path-aware recordings",
	Long:  "Print shell integration that loads recordings as shell functions so commands like cd affect the current shell.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := strings.TrimSpace(initShell)
		if len(args) == 1 {
			target = strings.TrimSpace(args[0])
		}

		if target == "" {
			target = string(shell.DetectShell())
		}

		switch strings.ToLower(target) {
		case "zsh", "bash":
			execName := filepath.Base(os.Args[0])
			if execName == "" {
				execName = "rec"
			}
			fmt.Print(shellInitScript(execName))
			return nil
		default:
			return fmt.Errorf("unsupported shell %q (expected zsh or bash)", target)
		}
	},
}

func init() {
	initCmd.Flags().StringVar(&initShell, "shell", "", "Shell to initialize (zsh or bash)")
	rootCmd.AddCommand(initCmd)
}

func shellInitScript(execName string) string {
	return fmt.Sprintf(`# rec shell integration
# Load once per shell startup:
#   eval "$(%s init zsh)"

rec_load_recordings() {
  local rec_dir="$HOME/.jtr/bin"
  [[ -d "$rec_dir" ]] || return 0

  local rec_file rec_name rec_existing
  for rec_file in "$rec_dir"/*; do
    [[ -f "$rec_file" ]] || continue
    rec_name="${rec_file##*/}"

    case "$rec_name" in
      *[!A-Za-z0-9._-]*|"")
        continue
        ;;
    esac

    rec_existing="$(command -v -- "$rec_name" 2>/dev/null || true)"
    if [[ -n "$rec_existing" && "$rec_existing" != "$rec_file" ]]; then
      continue
    fi

    eval "${rec_name}() { source \"${rec_file}\" \"\$@\"; }"
  done
}

rec_load_recordings

%s() {
	command %s "$@"
	local rec_rc=$?
	if [[ $rec_rc -eq 0 ]]; then
		case "${1:-}" in
			record|stop|delete)
				rec_load_recordings
				;;
		esac
	fi
	return $rec_rc
}
`, execName, execName, execName)
}

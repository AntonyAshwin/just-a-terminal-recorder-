package history

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"jtr/internal/shell"
)

var bashPrefix = regexp.MustCompile(`^\s*\d+\s+`)

// ReadLastN returns the last n commands from supported shell history.
// It first tries shell built-ins, then falls back to history files.
func ReadLastN(shellType shell.Type, n int) ([]string, error) {
	if n <= 0 {
		return nil, fmt.Errorf("n must be greater than 0")
	}

	// Try shell builtin first
	cmds, err := readViaShellBuiltin(shellType, n)
	if err == nil && len(cmds) > 0 {
		return cmds, nil
	}

	// Fallback to history file
	cmds, ferr := readViaHistoryFile(shellType, n)
	if ferr != nil {
		if err != nil {
			return nil, fmt.Errorf("failed to read history (builtin: %v, file: %v)", err, ferr)
		}
		return nil, ferr
	}
	return cmds, nil
}

func readViaShellBuiltin(shellType shell.Type, n int) ([]string, error) {
	var c *exec.Cmd

	switch shellType {
	case shell.Zsh:
		// fc is zsh builtin; use zsh -ic to get interactive history context
		c = exec.Command("zsh", "-ic", fmt.Sprintf("fc -ln -%d", n))
	case shell.Bash:
		c = exec.Command("bash", "-ic", fmt.Sprintf("history | tail -n %d", n))
	default:
		c = exec.Command("bash", "-ic", fmt.Sprintf("history | tail -n %d", n))
	}

	out, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("shell history command failed: %w (%s)", err, strings.TrimSpace(string(out)))
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	return normalizeHistoryLines(lines), nil
}

func readViaHistoryFile(shellType shell.Type, n int) ([]string, error) {
	path := os.Getenv("HISTFILE")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		switch shellType {
		case shell.Zsh:
			path = filepath.Join(home, ".zsh_history")
		case shell.Bash:
			path = filepath.Join(home, ".bash_history")
		default:
			path = filepath.Join(home, ".bash_history")
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var all []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		// zsh extended history format: ": 1700000000:0;command"
		if strings.HasPrefix(line, ": ") {
			if idx := strings.Index(line, ";"); idx >= 0 && idx+1 < len(line) {
				line = strings.TrimSpace(line[idx+1:])
			}
		}
		all = append(all, line)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	if len(all) == 0 {
		return nil, fmt.Errorf("history file is empty")
	}

	if n > len(all) {
		n = len(all)
	}
	return normalizeHistoryLines(all[len(all)-n:]), nil
}

func normalizeHistoryLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		line = bashPrefix.ReplaceAllString(line, "")

		// Bash 'history' output often has line numbers: " 123  git status"
		if i := strings.Index(line, "  "); i > 0 {
			prefix := strings.TrimSpace(line[:i])
			isNum := true
			for _, r := range prefix {
				if r < '0' || r > '9' {
					isNum = false
					break
				}
			}
			if isNum {
				line = strings.TrimSpace(line[i:])
			}
		}
		out = append(out, line)
	}
	return out
}

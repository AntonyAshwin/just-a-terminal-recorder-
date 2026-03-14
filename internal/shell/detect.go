package shell

import (
	"os"
	"path/filepath"
	"strings"
)

type Type string

const (
	Zsh  Type = "zsh"
	Bash Type = "bash"
)

func DetectShell() Type {
	shellEnv := strings.TrimSpace(os.Getenv("SHELL"))
	base := strings.ToLower(filepath.Base(shellEnv))

	switch base {
	case "zsh":
		return Zsh
	case "bash":
		return Bash
	default:
		return Bash
	}
}

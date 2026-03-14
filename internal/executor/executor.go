package executor

import (
	"fmt"
	"os"
	"os/exec"

	"jtr/internal/storage"
)

func RunScript(name string) error {
	path, err := storage.ScriptPath(name)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("recording %q does not exist", name)
		}
		return fmt.Errorf("failed to access recording %q: %w", name, err)
	}

	cmd := exec.Command(path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("recording %q failed: %w", name, err)
	}

	return nil
}

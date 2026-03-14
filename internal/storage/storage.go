package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type SessionData struct {
	Name      string   `json:"name"`
	StartedAt string   `json:"started_at"`
	StartDir  string   `json:"start_dir"`
	Baseline  []string `json:"baseline"`
}

var namePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// ValidName enforces safe script names to prevent path traversal.
func ValidName(name string) bool {
	return namePattern.MatchString(name)
}

// EnsureBinDir creates ~/.jtr/bin when it does not already exist.
func EnsureBinDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to detect home directory: %w", err)
	}
	dir := filepath.Join(home, ".jtr", "bin")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create jtr bin directory: %w", err)
	}
	return dir, nil
}

func ensureSessionDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to detect home directory: %w", err)
	}
	dir := filepath.Join(home, ".jtr", "sessions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create jtr session directory: %w", err)
	}
	return dir, nil
}

func sessionPath(name string) (string, error) {
	if !ValidName(name) {
		return "", fmt.Errorf("invalid session name %q", name)
	}
	dir, err := ensureSessionDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".session.json"), nil
}

func activeSessionPath() (string, error) {
	dir, err := ensureSessionDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "active.session.json"), nil
}

func ActiveSessionExists() (bool, error) {
	path, err := activeSessionPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check active session: %w", err)
}

func CreateActiveSession(baseline []string, startDir string) error {
	path, err := activeSessionPath()
	if err != nil {
		return err
	}

	exists, err := ActiveSessionExists()
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("an active session already exists")
	}

	payload := SessionData{
		Name:      "active",
		StartedAt: time.Now().Format(time.RFC3339),
		StartDir:  startDir,
		Baseline:  baseline,
	}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize active session: %w", err)
	}

	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return fmt.Errorf("failed to create active session: %w", err)
	}

	return nil
}

func ReadActiveSession() (*SessionData, error) {
	path, err := activeSessionPath()
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no active session")
		}
		return nil, fmt.Errorf("failed to read active session: %w", err)
	}

	var s SessionData
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("failed to parse active session: %w", err)
	}

	return &s, nil
}

func DeleteActiveSession() error {
	path, err := activeSessionPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete active session: %w", err)
	}
	return nil
}

func SessionExists(name string) (bool, error) {
	path, err := sessionPath(name)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check session %q: %w", name, err)
}

func CreateSession(name string, baseline []string) error {
	path, err := sessionPath(name)
	if err != nil {
		return err
	}
	exists, err := SessionExists(name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("session %q already exists", name)
	}

	payload := SessionData{
		Name:      name,
		StartedAt: time.Now().Format(time.RFC3339),
		Baseline:  baseline,
	}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize session %q: %w", name, err)
	}

	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return fmt.Errorf("failed to create session %q: %w", name, err)
	}
	return nil
}

func ReadSession(name string) (*SessionData, error) {
	path, err := sessionPath(name)
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session %q does not exist", name)
		}
		return nil, fmt.Errorf("failed to read session %q: %w", name, err)
	}

	var s SessionData
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("failed to parse session %q: %w", name, err)
	}

	return &s, nil
}

func DeleteSession(name string) error {
	path, err := sessionPath(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete session %q: %w", name, err)
	}
	return nil
}

func ScriptPath(name string) (string, error) {
	if !ValidName(name) {
		return "", fmt.Errorf("invalid script name %q", name)
	}
	dir, err := EnsureBinDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

func ScriptExists(name string) (bool, error) {
	path, err := ScriptPath(name)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check recording %q: %w", name, err)
}

// SaveScript writes the recording as an executable POSIX shell script.
func SaveScript(name string, commands []string) (string, error) {
	if len(commands) == 0 {
		return "", fmt.Errorf("no commands to save")
	}

	path, err := ScriptPath(name)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.WriteString("#!/bin/sh\n")
	for _, command := range commands {
		builder.WriteString(command)
		builder.WriteString("\n")
	}

	if err := os.WriteFile(path, []byte(builder.String()), 0o755); err != nil {
		return "", fmt.Errorf("failed to write script %q: %w", name, err)
	}

	if err := os.Chmod(path, 0o755); err != nil {
		return "", fmt.Errorf("failed to make script executable: %w", err)
	}

	return path, nil
}

func ListScripts() ([]string, error) {
	dir, err := EnsureBinDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to list recordings: %w", err)
	}

	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		out = append(out, entry.Name())
	}
	sort.Strings(out)
	return out, nil
}

func DeleteScript(name string) error {
	path, err := ScriptPath(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("recording %q does not exist", name)
		}
		return fmt.Errorf("failed to delete recording %q: %w", name, err)
	}
	return nil
}

func EditScript(name string) error {
	path, err := ScriptPath(name)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("recording %q does not exist", name)
		}
		return fmt.Errorf("failed to stat recording %q: %w", name, err)
	}

	editor := os.Getenv("VISUAL")
	if strings.TrimSpace(editor) == "" {
		editor = os.Getenv("EDITOR")
	}
	if strings.TrimSpace(editor) == "" {
		editor = "vi"
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	return nil
}

func ReadScript(name string) (string, error) {
	path, err := ScriptPath(name)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("recording %q does not exist", name)
		}
		return "", fmt.Errorf("failed to read recording %q: %w", name, err)
	}

	return string(content), nil
}

func ConfirmOverwrite(name string) (bool, error) {
	fmt.Printf("recording %q already exists. overwrite? [y/N]: ", name)

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return false, fmt.Errorf("failed to read confirmation input: %w", err)
	}

	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

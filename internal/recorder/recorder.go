package recorder

import (
	"path/filepath"
	"strings"
)

type Options struct {
	Commands      []string
	Name          string
	Count         int
	IgnorePath    bool
	IncludePath   bool
	AbsolutePaths bool
	CurrentDir    string
}

// Process filters and transforms history lines into the final script body.
// It preserves raw command text to keep pipes/quotes/arguments intact.
func Process(opts Options) []string {
	if len(opts.Commands) == 0 || opts.Count <= 0 {
		return nil
	}

	filtered := make([]string, 0, len(opts.Commands))
	for _, cmd := range opts.Commands {
		trimmed := strings.TrimSpace(cmd)
		if trimmed == "" {
			continue
		}
		if isJTRRecordCommand(trimmed) {
			continue
		}
		filtered = append(filtered, trimmed)
	}

	if len(filtered) == 0 {
		return nil
	}

	if len(filtered) > opts.Count {
		filtered = filtered[len(filtered)-opts.Count:]
	}

	result := make([]string, 0, len(filtered))
	currentDir := "."

	for _, line := range filtered {
		if opts.IgnorePath && isCDCommand(line) {
			continue
		}

		if opts.AbsolutePaths {
			line = absolutizeCD(line, currentDir)
		}

		if isCDCommand(line) {
			currentDir = applyCD(currentDir, line)
		}

		result = append(result, line)
	}

	if !opts.IgnorePath {
		result = canonicalizeSingleRelativeCD(result, opts.CurrentDir)
	}

	return result
}

func canonicalizeSingleRelativeCD(lines []string, currentDir string) []string {
	if len(lines) != 1 || strings.TrimSpace(currentDir) == "" {
		return lines
	}

	pathArg, ok := extractCDPath(lines[0])
	if !ok {
		return lines
	}

	if pathArg == "" || strings.HasPrefix(pathArg, "~") || filepath.IsAbs(pathArg) {
		return lines
	}

	return []string{"cd " + shellSingleQuote(filepath.Clean(currentDir))}
}

func extractCDPath(cmd string) (string, bool) {
	trimmed := strings.TrimSpace(cmd)
	if !isCDCommand(trimmed) {
		return "", false
	}

	parts := strings.SplitN(trimmed, " ", 2)
	if len(parts) != 2 {
		return "", true
	}

	pathArg := strings.TrimSpace(parts[1])
	if strings.HasPrefix(pathArg, "\"") && strings.HasSuffix(pathArg, "\"") && len(pathArg) >= 2 {
		pathArg = strings.Trim(pathArg, "\"")
	} else if strings.HasPrefix(pathArg, "'") && strings.HasSuffix(pathArg, "'") && len(pathArg) >= 2 {
		pathArg = strings.Trim(pathArg, "'")
	}

	return pathArg, true
}

func shellSingleQuote(input string) string {
	if input == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(input, "'", "'\"'\"'") + "'"
}

func isJTRRecordCommand(cmd string) bool {
	if cmd == "jtr record" || cmd == "rec record" {
		return true
	}
	return strings.HasPrefix(cmd, "jtr record ") || strings.HasPrefix(cmd, "rec record ")
}

func isCDCommand(cmd string) bool {
	trimmed := strings.TrimSpace(cmd)
	return trimmed == "cd" || strings.HasPrefix(trimmed, "cd ")
}

func absolutizeCD(cmd string, baseDir string) string {
	trimmed := strings.TrimSpace(cmd)
	if !isCDCommand(trimmed) {
		return cmd
	}

	parts := strings.SplitN(trimmed, " ", 2)
	if len(parts) != 2 {
		return cmd
	}

	pathArg := strings.TrimSpace(parts[1])
	quote := ""
	if strings.HasPrefix(pathArg, "\"") && strings.HasSuffix(pathArg, "\"") && len(pathArg) >= 2 {
		quote = "\""
		pathArg = strings.Trim(pathArg, "\"")
	} else if strings.HasPrefix(pathArg, "'") && strings.HasSuffix(pathArg, "'") && len(pathArg) >= 2 {
		quote = "'"
		pathArg = strings.Trim(pathArg, "'")
	}

	if pathArg == "" || strings.HasPrefix(pathArg, "~") || filepath.IsAbs(pathArg) {
		return cmd
	}

	abs := filepath.Clean(filepath.Join(baseDir, pathArg))
	if quote != "" {
		return "cd " + quote + abs + quote
	}
	return "cd " + abs
}

func applyCD(baseDir, cmd string) string {
	trimmed := strings.TrimSpace(cmd)
	if !isCDCommand(trimmed) {
		return baseDir
	}

	parts := strings.SplitN(trimmed, " ", 2)
	if len(parts) != 2 {
		return baseDir
	}

	pathArg := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
	if pathArg == "" {
		return baseDir
	}
	if filepath.IsAbs(pathArg) {
		return filepath.Clean(pathArg)
	}
	return filepath.Clean(filepath.Join(baseDir, pathArg))
}

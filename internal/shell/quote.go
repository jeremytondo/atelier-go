package shell

import "strings"

// Quote returns a shell-escaped version of the string.
// It wraps the string in single quotes and escapes any single quotes within the string.
// This ensures the string is treated as a single literal argument by the shell.
func Quote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

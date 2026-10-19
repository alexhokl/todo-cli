package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// runEditor runs the prepared editor command. It is a variable so that tests
// can substitute it without launching a real editor.
var runEditor = func(c *exec.Cmd) error {
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// editorCommand returns the editor executable and its arguments, taken from
// $VISUAL, then $EDITOR, falling back to vi. The value is split on whitespace
// so settings such as "code --wait" work.
func editorCommand() (string, []string) {
	for _, name := range []string{"VISUAL", "EDITOR"} {
		if fields := strings.Fields(os.Getenv(name)); len(fields) > 0 {
			return fields[0], fields[1:]
		}
	}
	return "vi", nil
}

// editInEditor opens initial in the user's editor and returns the edited text
// with trailing newlines removed (editors append one on save).
func editInEditor(initial string) (string, error) {
	file, err := os.CreateTemp("", "todo-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create a temporary file: %w", err)
	}
	path := file.Name()
	defer func() { _ = os.Remove(path) }()

	if _, err := file.WriteString(initial); err != nil {
		_ = file.Close()
		return "", fmt.Errorf("failed to write the temporary file: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("failed to write the temporary file: %w", err)
	}

	name, args := editorCommand()
	// The editor is configured by the user through their own environment.
	c := exec.Command(name, append(args, path)...) // #nosec G204
	if err := runEditor(c); err != nil {
		return "", fmt.Errorf("failed to run editor %q: %w", name, err)
	}

	edited, err := os.ReadFile(path) // #nosec G304 -- path created above
	if err != nil {
		return "", fmt.Errorf("failed to read the edited file: %w", err)
	}
	return strings.TrimRight(string(edited), "\r\n"), nil
}

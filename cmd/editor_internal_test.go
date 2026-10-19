package cmd

import (
	"errors"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func TestEditorCommand(t *testing.T) {
	tests := []struct {
		name         string
		visual       string
		editor       string
		expectedName string
		expectedArgs []string
	}{
		{"default", "", "", "vi", nil},
		{"editor", "", "nano", "nano", []string{}},
		{"visual wins", "emacs", "nano", "emacs", []string{}},
		{"arguments", "", "code --wait", "code", []string{"--wait"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("VISUAL", test.visual)
			t.Setenv("EDITOR", test.editor)
			name, args := editorCommand()
			if name != test.expectedName {
				t.Errorf("expected %q but got %q", test.expectedName, name)
			}
			if len(args) != len(test.expectedArgs) || (len(args) > 0 && !reflect.DeepEqual(args, test.expectedArgs)) {
				t.Errorf("expected %v but got %v", test.expectedArgs, args)
			}
		})
	}
}

func TestEditInEditor(t *testing.T) {
	original := runEditor
	t.Cleanup(func() { runEditor = original })

	t.Run("returns edited text", func(t *testing.T) {
		runEditor = func(c *exec.Cmd) error {
			path := c.Args[len(c.Args)-1]
			seed, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if string(seed) != "old" {
				t.Errorf("expected seed %q but got %q", "old", seed)
			}
			return os.WriteFile(path, []byte("new\nlines\n"), 0o600)
		}
		got, err := editInEditor("old")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "new\nlines" {
			t.Errorf("expected %q but got %q", "new\nlines", got)
		}
	})

	t.Run("editor failure", func(t *testing.T) {
		runEditor = func(*exec.Cmd) error { return errors.New("boom") }
		if _, err := editInEditor("old"); err == nil {
			t.Errorf("expected an error but got none")
		}
	})

	t.Run("temp file removed", func(t *testing.T) {
		var path string
		runEditor = func(c *exec.Cmd) error {
			path = c.Args[len(c.Args)-1]
			return nil
		}
		if _, err := editInEditor("old"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("expected temporary file to be removed")
		}
	})
}

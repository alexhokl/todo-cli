package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestVersionCommandOutput(t *testing.T) {
	// versionCmd.Run writes via fmt.Printf directly to os.Stdout, so the
	// command's SetOut is not honoured. Replace os.Stdout with a pipe,
	// run the command, and read the captured output.
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	defer func() {
		os.Stdout = origStdout
		_ = r.Close()
		_ = w.Close()
	}()
	os.Stdout = w

	versionCmd.Run(versionCmd, nil)
	_ = w.Close()

	buf := make([]byte, 0, 256)
	tmp := make([]byte, 256)
	for {
		n, err := r.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}
	output := string(buf)

	expected := AppName + " " + version + " (" + commit + ")\n"
	if output != expected {
		t.Errorf("expected %q but got %q", expected, output)
	}
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("expected a trailing newline but got %q", output)
	}
}

func TestVersionCommandDoesNotRequireService(t *testing.T) {
	if requiresService(versionCmd) {
		t.Errorf("version command must not require a service URI")
	}
}
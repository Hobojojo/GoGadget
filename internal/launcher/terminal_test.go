package launcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"tui-app-launcher/internal/interfaces"
)

func TestTerminalLaunchUsesWorkingDirectory(t *testing.T) {
	dir := t.TempDir()
	terminal := filepath.Join(dir, "terminal")
	output := filepath.Join(dir, "launch.txt")
	if err := os.WriteFile(terminal, []byte("#!/bin/sh\n{ pwd; printf '%s\\n' \"$@\"; } > \"$LAUNCH_TEST_OUTPUT.tmp\"\nmv \"$LAUNCH_TEST_OUTPUT.tmp\" \"$LAUNCH_TEST_OUTPUT\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TERMINAL", terminal)
	t.Setenv("LAUNCH_TEST_OUTPUT", output)
	if err := NewLauncher().LaunchApplication(interfaces.Application{Name: "CLI", Exec: "echo hello", Terminal: true, Path: dir}); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(time.Second)
	for {
		if content, err := os.ReadFile(output); err == nil {
			if got := string(content); got != dir+"\n-e\necho\nhello\n" {
				t.Fatalf("unexpected terminal invocation: %q", got)
			}
			break
		}
		select {
		case <-deadline:
			t.Fatal("terminal did not run")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestTerminalAndWorkingDirectoryValidation(t *testing.T) {
	dir := t.TempDir()
	terminal := filepath.Join(dir, "terminal")
	if err := os.WriteFile(terminal, []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TERMINAL", terminal)
	l := NewLauncher()
	if err := l.ValidateApplication(interfaces.Application{Name: "CLI", Exec: "echo hello", Terminal: true, Path: dir}); err != nil {
		t.Fatalf("valid terminal app rejected: %v", err)
	}
	if err := l.ValidateApplication(interfaces.Application{Name: "CLI", Exec: "echo", Path: filepath.Join(dir, "absent")}); err == nil {
		t.Fatal("nonexistent working directory accepted")
	}
}

package scanner

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestScanCancelledBeforeAndDuringWalk(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.desktop"), []byte("[Desktop Entry]\nName=App\nExec=echo\n"), 0600); err != nil {
		t.Fatal(err)
	}
	s := NewScannerWithPaths([]string{root})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if apps, err := s.ScanApplicationsContext(ctx); !errors.Is(err, context.Canceled) || apps != nil {
		t.Fatalf("cancelled scan returned apps=%v err=%v", apps, err)
	}
	apps, err := s.ScanApplicationsContext(context.Background())
	if err != nil || len(apps) != 1 || apps[0].Name != "App" {
		t.Fatalf("normal scan returned apps=%v err=%v", apps, err)
	}
}

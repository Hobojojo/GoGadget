package logging

import (
	"os"
	"strings"
	"sync"
	"testing"
)

func TestConcurrentLoggingAcrossRotation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	logger, err := NewLogger()
	if err != nil {
		t.Fatal(err)
	}
	logger.SetMaxSize(128)
	logger.SetMaxFiles(3)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				logger.Info("message %d", j)
			}
		}()
	}
	wg.Wait()
	if err := logger.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(logger.GetLogPath())
	if err != nil || !strings.Contains(string(data), "message") {
		t.Fatalf("missing log output: %v, %q", err, data)
	}
}

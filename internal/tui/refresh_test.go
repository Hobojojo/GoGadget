package tui

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"tui-app-launcher/internal/interfaces"
)

type blockingScanner struct {
	started chan struct{}
	stopped chan struct{}
}

func (s *blockingScanner) ScanApplications() ([]interfaces.Application, error)  { return nil, nil }
func (s *blockingScanner) RefreshApplications() error                           { return nil }
func (s *blockingScanner) RefreshApplicationsContext(ctx context.Context) error { return ctx.Err() }
func (s *blockingScanner) ScanApplicationsContext(ctx context.Context) ([]interfaces.Application, error) {
	close(s.started)
	<-ctx.Done()
	close(s.stopped)
	return nil, ctx.Err()
}

func TestRefreshCancelsOnExit(t *testing.T) {
	s := &blockingScanner{started: make(chan struct{}), stopped: make(chan struct{})}
	m := testModel(t, 0)
	m.scanner = s
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	m = next.(Model)
	if !m.isRefreshing || cmd == nil {
		t.Fatal("refresh did not start")
	}
	result := make(chan tea.Msg, 1)
	go func() { result <- cmd() }()
	select {
	case <-s.started:
	case <-time.After(time.Second):
		t.Fatal("scan never started")
	}
	exited, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = exited.(Model)
	select {
	case <-s.stopped:
	case <-time.After(time.Second):
		t.Fatal("scan was not cancelled on exit")
	}
	if msg := (<-result).(refreshCompleteMsg); !errors.Is(msg.err, context.Canceled) {
		t.Fatalf("cancelled scan returned %v", msg.err)
	}
}

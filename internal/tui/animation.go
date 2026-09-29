package tui

import (
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
)

const animationFPS = 30

// Each motion is model-owned: sessions never share animation state.
type motion struct {
	spring                  harmonica.Spring
	value, velocity, target float64
	active                  bool
}

func newMotion(frequency, damping float64) motion {
	return motion{spring: harmonica.NewSpring(harmonica.FPS(animationFPS), frequency, damping)}
}

func (s *motion) snap(value float64) {
	s.value, s.target, s.velocity, s.active = value, value, 0, false
}

func (s *motion) aim(target float64) {
	s.target = target
	s.active = math.Abs(s.value-target) > 0.001 || math.Abs(s.velocity) > 0.01
}

func (s *motion) enter(from, target float64) {
	s.snap(from)
	s.aim(target)
}

func (s *motion) step() {
	if !s.active {
		return
	}
	s.value, s.velocity = s.spring.Update(s.value, s.velocity, s.target)
	if math.Abs(s.value-s.target) < 0.001 && math.Abs(s.velocity) < 0.01 {
		s.snap(s.target)
	}
}

type animationState struct {
	cursor, list, help, error, pulse motion
	pending                          bool
	frameID                          uint64
}

func newAnimations() animationState {
	return animationState{
		cursor: newMotion(26, 1),
		list:   newMotion(22, 0.85),
		help:   newMotion(24, 1),
		error:  newMotion(26, 0.85),
		pulse:  newMotion(24, 0.9),
	}
}

func (a animationState) active() bool {
	return a.cursor.active || a.list.active || a.help.active || a.error.active || a.pulse.active
}

func (a *animationState) stop() {
	a.cursor.snap(a.cursor.target)
	a.list.snap(0)
	a.help.snap(a.help.target)
	a.error.snap(0)
	a.pulse.snap(0)
	a.pending = false
	a.frameID++ // Invalidate a tick that was already in flight.
}

// Only one tick is ever outstanding. Settled animations consume no idle ticks.
func (m *Model) animationTick() tea.Cmd {
	if m.animations.pending || !m.animations.active() {
		return nil
	}
	m.animations.pending = true
	m.animations.frameID++
	id := m.animations.frameID
	return tea.Tick(time.Second/animationFPS, func(time.Time) tea.Msg {
		return animationFrameMsg{id: id}
	})
}

// Update keeps animation scheduling separate from application behavior.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if frame, ok := msg.(animationFrameMsg); ok {
		if !m.animations.pending || frame.id != m.animations.frameID {
			return m, nil
		}
		m.animations.pending = false
		m.animations.cursor.step()
		m.animations.list.step()
		m.animations.help.step()
		m.animations.error.step()
		m.animations.pulse.step()
		cmd := m.animationTick()
		return m, cmd
	}

	model, cmd := m.update(msg)
	next := model.(Model)
	quitting := false
	switch event := msg.(type) {
	case tea.KeyMsg:
		quitting = event.String() == "esc" || event.String() == "ctrl+c"
	case launchSuccessMsg:
		quitting = true
	case launchErrorMsg:
		quitting = cmd != nil // Unrecoverable errors return tea.Quit.
	}
	if quitting {
		next.animations.stop()
		return next, cmd
	}

	next.reconcileAnimations(m, msg)
	tick := next.animationTick()
	if tick == nil {
		return next, cmd
	}
	if cmd == nil {
		return next, tick
	}
	return next, tea.Batch(cmd, tick)
}

func (m *Model) reconcileAnimations(previous Model, msg tea.Msg) {
	// Error appearance changes list height; clamp before choosing visual targets.
	m.ensureSelectionVisible()
	target := float64((m.selectedIndex - m.scrollOffset) / m.columns())
	key, keyEvent := msg.(tea.KeyMsg)
	navigating := keyEvent && (key.String() == "up" || key.String() == "down")
	_, resizing := msg.(tea.WindowSizeMsg)

	if navigating {
		m.animations.list.snap(0) // Navigation never waits for an entrance animation.
		shift := float64(previous.scrollOffset-m.scrollOffset) / float64(m.columns())
		m.animations.cursor.value = math.Max(0, math.Min(float64(max(0, m.visibleRows()-1)), m.animations.cursor.value+shift))
		if math.Abs(float64(previous.selectedIndex-m.selectedIndex)) > 1 {
			m.animations.cursor.snap(target) // Wraparound should not sweep the entire list.
		} else {
			m.animations.cursor.aim(target)
		}
	} else if resizing || previous.selectedIndex != m.selectedIndex || previous.scrollOffset != m.scrollOffset || previous.searchQuery != m.searchQuery {
		m.animations.cursor.snap(target)
		m.animations.list.snap(0)
	}

	if previous.showHelp != m.showHelp {
		target := 0.0
		if m.showHelp {
			target = 1
		}
		m.animations.help.aim(target)
	}

	entrance := false
	switch event := msg.(type) {
	case initMsg:
		entrance = m.scanner == nil
	case refreshCompleteMsg:
		entrance = event.err == nil
	}
	if entrance && len(m.filteredApps) > 0 {
		// Never slide the logically selected row below the viewport.
		distance := min(3, max(0, m.visibleRows()-1-int(target)))
		m.animations.list.enter(float64(distance), 0)
		m.animations.cursor.snap(target)
	}
	if resizing {
		m.animations.list.snap(0)
	}

	if m.errorMessage == "" {
		m.animations.error.snap(0)
	} else if m.errorMessage != previous.errorMessage {
		m.animations.error.enter(6, 0)
	}
	if keyEvent && previous.searchQuery == "" && m.searchQuery != "" {
		m.animations.pulse.enter(1, 0)
	} else if m.searchQuery == "" {
		m.animations.pulse.snap(0)
	}
}

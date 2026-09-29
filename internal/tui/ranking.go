package tui

import (
	"sort"
	"time"

	"tui-app-launcher/internal/interfaces"
)

// Frequency decays with a seven-day half-life approximation. Unused apps score
// zero; future timestamps are clamped so clock changes cannot inflate ranking.
func frecency(record interfaces.LaunchRecord, now time.Time) float64 {
	if record.Count <= 0 || record.LastUsed.IsZero() {
		return 0
	}
	days := max(0, now.Sub(record.LastUsed).Hours()/24)
	return float64(record.Count) / (1 + days/7)
}

func (m *Model) rankByHistory(results []interfaces.SearchResult) {
	manager, ok := m.configManager.(interfaces.LaunchHistoryManager)
	if !ok {
		return
	}
	history, now := manager.LaunchHistory(), time.Now()
	if results == nil {
		// Never sort the scanner's backing slice in place.
		m.filteredApps = append([]interfaces.Application(nil), m.filteredApps...)
		sort.SliceStable(m.filteredApps, func(i, j int) bool {
			return frecency(history[m.filteredApps[i].HistoryID()], now) > frecency(history[m.filteredApps[j].HistoryID()], now)
		})
		return
	}
	// Text relevance is primary; usage breaks ties without hiding better matches.
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return frecency(history[results[i].Application.HistoryID()], now) > frecency(history[results[j].Application.HistoryID()], now)
	})
}

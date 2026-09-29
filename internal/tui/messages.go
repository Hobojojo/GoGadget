package tui

import "tui-app-launcher/internal/interfaces"

// animationFrameMsg advances only the currently scheduled presentation frame.
type animationFrameMsg struct{ id uint64 }

// initMsg is sent when the model is initialized
type initMsg struct{}

// launchMsg is sent when an application should be launched
type launchMsg struct {
	app interfaces.Application
}

// launchSuccessMsg is sent when an application launches successfully
type launchSuccessMsg struct{}

// launchErrorMsg is sent when an application fails to launch
type launchErrorMsg struct {
	err error
}

// toggleFavoriteMsg is sent when favorite status should be toggled
type toggleFavoriteMsg struct {
	app interfaces.Application
}

// refreshMsg is sent when applications should be refreshed
type refreshMsg struct{}

// refreshCompleteMsg is sent when refresh is complete
type refreshCompleteMsg struct {
	apps []interfaces.Application
	err  error
}

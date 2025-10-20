package errors

import (
	"fmt"
)

// ErrorType represents different categories of errors in the application
type ErrorType int

const (
	// ConfigError represents configuration-related errors
	ConfigError ErrorType = iota
	// ScanError represents application scanning errors
	ScanError
	// LaunchError represents application launch errors
	LaunchError
	// TUIError represents terminal/interface errors
	TUIError
	// SearchError represents search-related errors
	SearchError
)

// String returns a string representation of the error type
func (et ErrorType) String() string {
	switch et {
	case ConfigError:
		return "Configuration Error"
	case ScanError:
		return "Scan Error"
	case LaunchError:
		return "Launch Error"
	case TUIError:
		return "Interface Error"
	case SearchError:
		return "Search Error"
	default:
		return "Unknown Error"
	}
}

// LauncherError represents a structured error with type and context
type LauncherError struct {
	Type    ErrorType
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *LauncherError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type.String(), e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type.String(), e.Message)
}

// Unwrap returns the underlying error for error unwrapping
func (e *LauncherError) Unwrap() error {
	return e.Cause
}

// GetUserFriendlyMessage returns a user-friendly error message for display in TUI
func (e *LauncherError) GetUserFriendlyMessage() string {
	switch e.Type {
	case ConfigError:
		return "Configuration issue - check settings"
	case ScanError:
		return "Failed to scan applications - some apps may not appear"
	case LaunchError:
		return e.getUserFriendlyLaunchMessage()
	case TUIError:
		return "Interface error - try resizing terminal"
	case SearchError:
		return "Search temporarily unavailable"
	default:
		return "An unexpected error occurred"
	}
}

// getUserFriendlyLaunchMessage returns user-friendly launch error messages
func (e *LauncherError) getUserFriendlyLaunchMessage() string {
	if e.Cause != nil {
		causeMsg := e.Cause.Error()

		// Check for common error patterns and provide friendly messages
		switch {
		case contains(causeMsg, "not found in PATH"):
			return "Application not found or not installed"
		case contains(causeMsg, "permission denied"):
			return "Permission denied - check file permissions"
		case contains(causeMsg, "no such file"):
			return "Application executable not found"
		case contains(causeMsg, "exec format error"):
			return "Application format not supported"
		case contains(causeMsg, "text file busy"):
			return "Application is currently being updated"
		default:
			return fmt.Sprintf("Launch failed: %s", e.Message)
		}
	}
	return e.Message
}

// contains checks if a string contains a substring (case-insensitive helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				findSubstring(s, substr)))
}

// findSubstring performs case-insensitive substring search
func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// toLower converts a byte to lowercase
func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

// NewConfigError creates a new configuration error
func NewConfigError(message string, cause error) *LauncherError {
	return &LauncherError{
		Type:    ConfigError,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// NewScanError creates a new scan error
func NewScanError(message string, cause error) *LauncherError {
	return &LauncherError{
		Type:    ScanError,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// NewLaunchError creates a new launch error
func NewLaunchError(message string, cause error) *LauncherError {
	return &LauncherError{
		Type:    LaunchError,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// NewTUIError creates a new TUI error
func NewTUIError(message string, cause error) *LauncherError {
	return &LauncherError{
		Type:    TUIError,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// NewSearchError creates a new search error
func NewSearchError(message string, cause error) *LauncherError {
	return &LauncherError{
		Type:    SearchError,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context information to an error
func (e *LauncherError) WithContext(key string, value interface{}) *LauncherError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// GetContext retrieves context information from an error
func (e *LauncherError) GetContext(key string) (interface{}, bool) {
	if e.Context == nil {
		return nil, false
	}
	value, exists := e.Context[key]
	return value, exists
}

// IsRecoverable determines if an error is recoverable (application can continue)
func (e *LauncherError) IsRecoverable() bool {
	switch e.Type {
	case ConfigError:
		// Most config errors are recoverable (use defaults)
		return true
	case ScanError:
		// Scan errors are recoverable (continue with partial results)
		return true
	case LaunchError:
		// Launch errors are recoverable (stay in interface)
		return true
	case SearchError:
		// Search errors are recoverable (fall back to no search)
		return true
	case TUIError:
		// TUI errors may or may not be recoverable
		return false
	default:
		return false
	}
}

// ErrorRecovery provides recovery strategies for different error types
type ErrorRecovery struct{}

// NewErrorRecovery creates a new error recovery handler
func NewErrorRecovery() *ErrorRecovery {
	return &ErrorRecovery{}
}

// RecoverFromError attempts to recover from an error and return a recovery action
func (er *ErrorRecovery) RecoverFromError(err error) RecoveryAction {
	if launcherErr, ok := err.(*LauncherError); ok {
		switch launcherErr.Type {
		case ConfigError:
			return RecoveryAction{
				Type:        UseDefaults,
				Message:     "Using default configuration",
				CanContinue: true,
			}
		case ScanError:
			return RecoveryAction{
				Type:        PartialOperation,
				Message:     "Continuing with available applications",
				CanContinue: true,
			}
		case LaunchError:
			return RecoveryAction{
				Type:        ShowError,
				Message:     launcherErr.GetUserFriendlyMessage(),
				CanContinue: true,
			}
		case SearchError:
			return RecoveryAction{
				Type:        FallbackMode,
				Message:     "Search disabled, showing all applications",
				CanContinue: true,
			}
		case TUIError:
			return RecoveryAction{
				Type:        Restart,
				Message:     "Interface error, please restart",
				CanContinue: false,
			}
		}
	}

	// Unknown error type
	return RecoveryAction{
		Type:        ShowError,
		Message:     "An unexpected error occurred",
		CanContinue: true,
	}
}

// RecoveryActionType represents different recovery strategies
type RecoveryActionType int

const (
	// UseDefaults uses default values/configuration
	UseDefaults RecoveryActionType = iota
	// PartialOperation continues with partial functionality
	PartialOperation
	// ShowError displays error to user and continues
	ShowError
	// FallbackMode switches to a simpler mode of operation
	FallbackMode
	// Restart requires application restart
	Restart
)

// RecoveryAction represents an action to take when recovering from an error
type RecoveryAction struct {
	Type        RecoveryActionType
	Message     string
	CanContinue bool
}

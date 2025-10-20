package launcher

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"tui-app-launcher/internal/errors"
	"tui-app-launcher/internal/interfaces"
	"tui-app-launcher/internal/logging"
)

// Launcher implements the ApplicationLauncher interface
type Launcher struct {
	logger *logging.Logger
}

// NewLauncher creates a new application launcher
func NewLauncher() *Launcher {
	return &Launcher{
		logger: logging.GetGlobalLogger(),
	}
}

// LaunchApplication launches the specified application
func (l *Launcher) LaunchApplication(app interfaces.Application) error {
	if app.Exec == "" {
		return errors.NewLaunchError("no executable specified", nil).
			WithContext("app_name", app.Name)
	}

	// Parse the Exec field to handle command arguments and field codes
	command, args, err := l.parseExecField(app.Exec)
	if err != nil {
		return errors.NewLaunchError("failed to parse exec field", err).
			WithContext("app_name", app.Name).
			WithContext("exec_field", app.Exec)
	}

	// Check if the command exists in PATH
	if _, err := exec.LookPath(command); err != nil {
		return errors.NewLaunchError("command not found in PATH", err).
			WithContext("app_name", app.Name).
			WithContext("command", command)
	}

	// Create the command
	cmd := exec.Command(command, args...)

	// Set up the command environment
	cmd.Env = os.Environ()

	// Detach from the current process group to prevent the launched app
	// from being killed when the launcher exits
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Redirect stdout and stderr to prevent output from interfering with TUI
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	// Start the application in the background
	err = cmd.Start()
	if err != nil {
		if l.logger != nil {
			l.logger.LogApplicationLaunch(app.Name, command, false, err)
		}
		return errors.NewLaunchError("failed to start application", err).
			WithContext("app_name", app.Name).
			WithContext("command", command)
	}

	// Log successful launch
	if l.logger != nil {
		l.logger.LogApplicationLaunch(app.Name, command, true, nil)
	}

	// Wait a brief moment to check if the process starts successfully
	// This helps catch immediate failures (like missing libraries)
	go func() {
		time.Sleep(100 * time.Millisecond)
		// Check if process is still running
		if cmd.Process != nil {
			// Send signal 0 to check if process exists (doesn't actually send a signal)
			err := cmd.Process.Signal(syscall.Signal(0))
			if err != nil {
				// Process has already exited, but we don't report this as an error
				// since the launcher should have already exited by now
			}
		}
	}()

	// The launcher will exit after starting the application
	return nil
}

// parseExecField parses the Exec field from desktop files
// Handles field codes like %f, %F, %u, %U, %i, %c, %k and removes them
// Returns the command and its arguments
func (l *Launcher) parseExecField(execField string) (string, []string, error) {
	if execField == "" {
		return "", nil, errors.NewLaunchError("empty exec field", nil)
	}

	// Split the exec field into parts, respecting quoted strings
	parts, err := l.splitExecField(execField)
	if err != nil {
		return "", nil, err
	}

	if len(parts) == 0 {
		return "", nil, errors.NewLaunchError("no command found in exec field", nil).
			WithContext("exec_field", execField)
	}

	command := parts[0]
	args := make([]string, 0)

	// Process arguments and remove field codes
	for i := 1; i < len(parts); i++ {
		arg := parts[i]

		// Skip desktop file field codes
		if l.isFieldCode(arg) {
			continue
		}

		args = append(args, arg)
	}

	return command, args, nil
}

// splitExecField splits the exec field into command and arguments
// Handles quoted strings properly
func (l *Launcher) splitExecField(execField string) ([]string, error) {
	var parts []string
	var current strings.Builder
	inQuotes := false
	escapeNext := false

	for _, char := range execField {
		if escapeNext {
			current.WriteRune(char)
			escapeNext = false
			continue
		}

		switch char {
		case '\\':
			if inQuotes {
				escapeNext = true
			} else {
				current.WriteRune(char)
			}
		case '"':
			inQuotes = !inQuotes
		case ' ', '\t':
			if inQuotes {
				current.WriteRune(char)
			} else {
				if current.Len() > 0 {
					parts = append(parts, current.String())
					current.Reset()
				}
			}
		default:
			current.WriteRune(char)
		}
	}

	// Add the last part if there's content
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	// Check for unclosed quotes
	if inQuotes {
		return nil, errors.NewLaunchError("unclosed quotes in exec field", nil).
			WithContext("exec_field", execField)
	}

	return parts, nil
}

// isFieldCode checks if the argument is a desktop file field code
// Field codes: %f %F %u %U %d %D %n %N %i %c %k %v %m
func (l *Launcher) isFieldCode(arg string) bool {
	fieldCodes := []string{
		"%f", "%F", "%u", "%U", "%d", "%D",
		"%n", "%N", "%i", "%c", "%k", "%v", "%m",
	}

	for _, code := range fieldCodes {
		if arg == code {
			return true
		}
	}

	return false
}

// ValidateApplication checks if an application can be launched
func (l *Launcher) ValidateApplication(app interfaces.Application) error {
	if app.Exec == "" {
		return errors.NewLaunchError("no executable specified", nil).
			WithContext("app_name", app.Name)
	}

	// Parse the Exec field to validate it
	command, _, err := l.parseExecField(app.Exec)
	if err != nil {
		return errors.NewLaunchError("invalid exec field", err).
			WithContext("app_name", app.Name).
			WithContext("exec_field", app.Exec)
	}

	// Check if the command exists in PATH
	if _, err := exec.LookPath(command); err != nil {
		return errors.NewLaunchError("command not found in PATH", err).
			WithContext("app_name", app.Name).
			WithContext("command", command)
	}

	return nil
}

// GetLaunchCommand returns the parsed command and arguments for an application
// This is useful for debugging or displaying what would be executed
func (l *Launcher) GetLaunchCommand(app interfaces.Application) (string, []string, error) {
	if app.Exec == "" {
		return "", nil, errors.NewLaunchError("no executable specified", nil).
			WithContext("app_name", app.Name)
	}

	return l.parseExecField(app.Exec)
}

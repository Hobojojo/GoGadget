package launcher

import (
	"strings"
	"testing"

	"tui-app-launcher/internal/interfaces"
)

func TestParseExecField(t *testing.T) {
	launcher := NewLauncher()

	tests := []struct {
		name        string
		execField   string
		wantCommand string
		wantArgs    []string
		wantError   bool
	}{
		{
			name:        "simple command",
			execField:   "firefox",
			wantCommand: "firefox",
			wantArgs:    []string{},
			wantError:   false,
		},
		{
			name:        "command with arguments",
			execField:   "firefox --new-window",
			wantCommand: "firefox",
			wantArgs:    []string{"--new-window"},
			wantError:   false,
		},
		{
			name:        "command with field codes",
			execField:   "firefox %u",
			wantCommand: "firefox",
			wantArgs:    []string{}, // Field codes should be removed
			wantError:   false,
		},
		{
			name:        "quoted arguments",
			execField:   `firefox "https://example.com"`,
			wantCommand: "firefox",
			wantArgs:    []string{"https://example.com"},
			wantError:   false,
		},
		{
			name:        "complex command with multiple field codes",
			execField:   "code --new-window %F --goto %f",
			wantCommand: "code",
			wantArgs:    []string{"--new-window", "--goto"},
			wantError:   false,
		},
		{
			name:        "empty exec field",
			execField:   "",
			wantCommand: "",
			wantArgs:    nil,
			wantError:   true,
		},
		{
			name:        "unclosed quotes",
			execField:   `firefox "unclosed quote`,
			wantCommand: "",
			wantArgs:    nil,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, args, err := launcher.parseExecField(tt.execField)

			if tt.wantError {
				if err == nil {
					t.Errorf("parseExecField() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("parseExecField() unexpected error: %v", err)
				return
			}

			if command != tt.wantCommand {
				t.Errorf("parseExecField() command = %v, want %v", command, tt.wantCommand)
			}

			if len(args) != len(tt.wantArgs) {
				t.Errorf("parseExecField() args length = %v, want %v", len(args), len(tt.wantArgs))
				return
			}

			for i, arg := range args {
				if arg != tt.wantArgs[i] {
					t.Errorf("parseExecField() args[%d] = %v, want %v", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestIsFieldCode(t *testing.T) {
	launcher := NewLauncher()

	tests := []struct {
		arg  string
		want bool
	}{
		{"%f", true},
		{"%F", true},
		{"%u", true},
		{"%U", true},
		{"%d", true},
		{"%D", true},
		{"%n", true},
		{"%N", true},
		{"%i", true},
		{"%c", true},
		{"%k", true},
		{"%v", true},
		{"%m", true},
		{"--new-window", false},
		{"firefox", false},
		{"%x", false}, // Not a valid field code
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			got := launcher.isFieldCode(tt.arg)
			if got != tt.want {
				t.Errorf("isFieldCode(%q) = %v, want %v", tt.arg, got, tt.want)
			}
		})
	}
}

func TestLaunchApplication_ValidationErrors(t *testing.T) {
	launcher := NewLauncher()

	tests := []struct {
		name    string
		app     interfaces.Application
		wantErr string
	}{
		{
			name: "empty exec field",
			app: interfaces.Application{
				Name: "Test App",
				Exec: "",
			},
			wantErr: "no executable specified",
		},
		{
			name: "invalid exec field with unclosed quotes",
			app: interfaces.Application{
				Name: "Test App",
				Exec: `firefox "unclosed`,
			},
			wantErr: "failed to parse exec field",
		},
		{
			name: "nonexistent command",
			app: interfaces.Application{
				Name: "Test App",
				Exec: "nonexistent-command-12345",
			},
			wantErr: "not found in PATH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := launcher.LaunchApplication(tt.app)
			if err == nil {
				t.Errorf("LaunchApplication() expected error but got none")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("LaunchApplication() error = %v, want error containing %v", err, tt.wantErr)
			}
		})
	}
}

func TestSplitExecField(t *testing.T) {
	launcher := NewLauncher()

	tests := []struct {
		name      string
		execField string
		want      []string
		wantError bool
	}{
		{
			name:      "simple command",
			execField: "firefox",
			want:      []string{"firefox"},
			wantError: false,
		},
		{
			name:      "command with spaces",
			execField: "firefox --new-window",
			want:      []string{"firefox", "--new-window"},
			wantError: false,
		},
		{
			name:      "quoted argument",
			execField: `firefox "https://example.com"`,
			want:      []string{"firefox", "https://example.com"},
			wantError: false,
		},
		{
			name:      "multiple quoted arguments",
			execField: `code "file 1.txt" "file 2.txt"`,
			want:      []string{"code", "file 1.txt", "file 2.txt"},
			wantError: false,
		},
		{
			name:      "escaped quotes",
			execField: `echo "He said \"Hello\""`,
			want:      []string{"echo", `He said "Hello"`},
			wantError: false,
		},
		{
			name:      "unclosed quotes",
			execField: `firefox "unclosed`,
			want:      nil,
			wantError: true,
		},
		{
			name:      "empty string",
			execField: "",
			want:      []string{},
			wantError: false,
		},
		{
			name:      "only spaces",
			execField: "   ",
			want:      []string{},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := launcher.splitExecField(tt.execField)

			if tt.wantError {
				if err == nil {
					t.Errorf("splitExecField() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("splitExecField() unexpected error: %v", err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("splitExecField() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for i, part := range got {
				if part != tt.want[i] {
					t.Errorf("splitExecField() part[%d] = %v, want %v", i, part, tt.want[i])
				}
			}
		})
	}
}func 
TestValidateApplication(t *testing.T) {
	launcher := NewLauncher()

	tests := []struct {
		name    string
		app     interfaces.Application
		wantErr bool
	}{
		{
			name: "valid application with existing command",
			app: interfaces.Application{
				Name: "Echo Test",
				Exec: "echo hello",
			},
			wantErr: false,
		},
		{
			name: "empty exec field",
			app: interfaces.Application{
				Name: "Test App",
				Exec: "",
			},
			wantErr: true,
		},
		{
			name: "nonexistent command",
			app: interfaces.Application{
				Name: "Test App",
				Exec: "nonexistent-command-12345",
			},
			wantErr: true,
		},
		{
			name: "invalid exec field",
			app: interfaces.Application{
				Name: "Test App",
				Exec: `command "unclosed quote`,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := launcher.ValidateApplication(tt.app)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateApplication() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetLaunchCommand(t *testing.T) {
	launcher := NewLauncher()

	tests := []struct {
		name        string
		app         interfaces.Application
		wantCommand string
		wantArgs    []string
		wantErr     bool
	}{
		{
			name: "simple command",
			app: interfaces.Application{
				Name: "Firefox",
				Exec: "firefox",
			},
			wantCommand: "firefox",
			wantArgs:    []string{},
			wantErr:     false,
		},
		{
			name: "command with arguments",
			app: interfaces.Application{
				Name: "Firefox",
				Exec: "firefox --new-window %u",
			},
			wantCommand: "firefox",
			wantArgs:    []string{"--new-window"},
			wantErr:     false,
		},
		{
			name: "empty exec field",
			app: interfaces.Application{
				Name: "Test App",
				Exec: "",
			},
			wantCommand: "",
			wantArgs:    nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, args, err := launcher.GetLaunchCommand(tt.app)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLaunchCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return // Skip further checks if error was expected
			}

			if command != tt.wantCommand {
				t.Errorf("GetLaunchCommand() command = %v, want %v", command, tt.wantCommand)
			}

			if len(args) != len(tt.wantArgs) {
				t.Errorf("GetLaunchCommand() args length = %v, want %v", len(args), len(tt.wantArgs))
				return
			}

			for i, arg := range args {
				if arg != tt.wantArgs[i] {
					t.Errorf("GetLaunchCommand() args[%d] = %v, want %v", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}
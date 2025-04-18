package macro

import (
	"os"
	"testing"

	"github.com/ksysoev/wsget/pkg/core"
	"github.com/ksysoev/wsget/pkg/core/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMacro(t *testing.T) {
	tests := []struct {
		name    string
		want    *Repo
		domains []string
	}{
		{
			name:    "empty domains",
			domains: []string{},
			want: &Repo{
				macro:   make(map[string]*command.Templates),
				domains: []string{},
			},
		},
		{
			name:    "non-empty domains",
			domains: []string{"example.com", "google.com"},
			want: &Repo{
				macro:   make(map[string]*command.Templates),
				domains: []string{"example.com", "google.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.domains); got == nil {
				t.Errorf("New() = %v, want non-nil", got)
			} else if len(got.macro) != 0 {
				t.Errorf("New() = %v, want empty macro map", got)
			} else if len(got.domains) != len(tt.domains) {
				t.Errorf("New() = %v, want domains %v", got, tt.domains)
			}
		})
	}
}
func TestMacro_AddCommands(t *testing.T) {
	tests := []struct {
		name        string
		macro       *Repo
		commandName string
		commands    []string
		wantErr     bool
	}{
		{
			name:        "add new macro",
			macro:       New([]string{}),
			commandName: "test",
			commands:    []string{"edit hello"},
			wantErr:     false,
		},
		{
			name:        "add existing macro",
			macro:       &Repo{macro: map[string]*command.Templates{"test": nil}},
			commandName: "test",
			commands:    []string{"send hello"},
			wantErr:     true,
		},
		{
			name:        "empty macro",
			macro:       New([]string{}),
			commandName: "test",
			commands:    []string{},
			wantErr:     true,
		},
		{
			name:        "single command macro",
			macro:       New([]string{}),
			commandName: "test",
			commands:    []string{"exit hello"},
			wantErr:     false,
		},
		{
			name:        "multi command macro",
			macro:       New([]string{}),
			commandName: "test",
			commands:    []string{"send hello", "wait 5"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.macro.AddCommands(tt.commandName, tt.commands)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repo.AddCommands() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func TestMacro_Merge(t *testing.T) {
	tests := []struct {
		macro       *Repo
		otherMacro  *Repo
		name        string
		expectedLen int
		wantErr     bool
	}{
		{
			name: "merge empty macro with empty macro",
			macro: &Repo{
				macro:   make(map[string]*command.Templates),
				domains: []string{},
			},
			otherMacro: &Repo{
				macro:   make(map[string]*command.Templates),
				domains: []string{},
			},
			wantErr:     false,
			expectedLen: 0,
		},
		{
			name: "merge non-empty macro with empty macro",
			macro: &Repo{
				macro: map[string]*command.Templates{
					"test": nil,
				},
				domains: []string{},
			},
			otherMacro: &Repo{
				macro:   make(map[string]*command.Templates),
				domains: []string{},
			},
			wantErr:     false,
			expectedLen: 1,
		},
		{
			name: "merge empty macro with non-empty macro",
			macro: &Repo{
				macro:   make(map[string]*command.Templates),
				domains: []string{},
			},
			otherMacro: &Repo{
				macro: map[string]*command.Templates{
					"test": nil,
				},
				domains: []string{},
			},
			wantErr:     false,
			expectedLen: 1,
		},
		{
			name: "merge non-empty macro with non-empty macro",
			macro: &Repo{
				macro: map[string]*command.Templates{
					"test": nil,
				},
				domains: []string{},
			},
			otherMacro: &Repo{
				macro: map[string]*command.Templates{
					"test2": nil,
				},
				domains: []string{},
			},
			wantErr:     false,
			expectedLen: 2,
		},
		{
			name: "merge macro with duplicate macro name",
			macro: &Repo{
				macro: map[string]*command.Templates{
					"test": nil,
				},
				domains: []string{},
			},
			otherMacro: &Repo{
				macro: map[string]*command.Templates{
					"test": nil,
				},
				domains: []string{},
			},
			wantErr:     true,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.macro.merge(tt.otherMacro)
			if (err != nil) != tt.wantErr {
				t.Errorf("Repo.merge() error = %v, wantErr %v", err, tt.wantErr)
			} else if len(tt.macro.macro) != tt.expectedLen {
				t.Errorf("Repo.merge() expected length of macro map = %d, got %d", tt.expectedLen, len(tt.macro.macro))
			}
		})
	}
}
func TestMacro_Get(t *testing.T) {
	testTemplate, _ := command.NewMacro([]string{"exit"})
	tests := []struct {
		name    string
		macro   *Repo
		cmdName string
		wantCmd core.Executer
		errMsg  string
		wantErr bool
	}{
		{
			name:    "get existing command",
			macro:   &Repo{macro: map[string]*command.Templates{"test": testTemplate}},
			cmdName: "test",
			wantCmd: command.NewExit(),
			wantErr: false,
			errMsg:  "",
		},
		{
			name:    "get non-existing command",
			macro:   &Repo{macro: map[string]*command.Templates{}},
			cmdName: "test",
			wantCmd: nil,
			wantErr: true,
			errMsg:  "unknown command: test",
		},
		{
			name:    "get command with empty macro",
			macro:   &Repo{macro: map[string]*command.Templates{}},
			cmdName: "",
			wantCmd: nil,
			wantErr: true,
			errMsg:  "unknown command: ",
		},
		{
			name:    "get command with non-empty macro",
			macro:   &Repo{macro: map[string]*command.Templates{"test": nil}},
			cmdName: "",
			wantCmd: nil,
			wantErr: true,
			errMsg:  "unknown command: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := tt.macro.Get(tt.cmdName, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("Repo.Get() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("Repo.Get() error message = %v, want %v", err.Error(), tt.errMsg)
			}

			if cmd != tt.wantCmd {
				t.Errorf("Repo.Get() cmd = %v, want %v", cmd, tt.wantCmd)
			}
		})
	}
}
func TestLoadFromFile(t *testing.T) {
	macroDir := os.TempDir()
	domain := "example.com"

	// Create temporary test file
	tempFile, err := os.CreateTemp(macroDir, "macro.yaml")
	if err != nil {
		t.Fatalf("Failed to create temporary test file: %v", err)
	}

	// Write test data to the temporary test file
	_, err = tempFile.WriteString(`
version: 1
domains:
  - example.com
macro:
  test:
    - send hello
    - wait 5
`)

	if err != nil {
		t.Fatalf("Failed to write to temporary test file: %v", err)
	}

	// Load macro from temporary test file
	macro, err := LoadFromFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load macro from temporary test file: %v", err)
	}

	// Check if macro was loaded correctly
	if len(macro.macro) != 1 {
		t.Errorf("LoadFromFile() macro length = %d, want %d", len(macro.macro), 1)
	}

	if len(macro.domains) != 1 {
		t.Errorf("LoadFromFile() domains length = %d, want %d", len(macro.domains), 1)
	}

	if macro.domains[0] != domain {
		t.Errorf("LoadFromFile() domain = %s, want %s", macro.domains[0], domain)
	}

	cmd, err := macro.Get("test", "")

	if err != nil {
		t.Errorf("LoadFromFile() error = %v, want nil", err)
	}

	if cmd == nil {
		t.Errorf("LoadFromFile() cmd = %v, want non-nil", cmd)
	}
}

func TestLoadFromFile_InvalidFile(t *testing.T) {
	macroDir := os.TempDir()

	// Create temporary test file
	tempFile, err := os.CreateTemp(macroDir, "macro.yaml")
	if err != nil {
		t.Fatalf("Failed to create temporary test file: %v", err)
	}

	// Write test data to the temporary test file
	_, err = tempFile.WriteString("Some\n  - invalid\n    data")
	if err != nil {
		t.Fatalf("Failed to write to temporary test file: %v", err)
	}

	// Load macro from temporary test file
	_, err = LoadFromFile(tempFile.Name())
	if err == nil {
		t.Fatalf("LoadFromFile() error = %v, want non-nil", err)
	}
}

func TestLoadFromFile_InvalidVersion(t *testing.T) {
	macroDir := os.TempDir()

	// Create temporary test file
	tempFile, err := os.CreateTemp(macroDir, "macro.yaml")
	if err != nil {
		t.Fatalf("Failed to create temporary test file: %v", err)
	}

	// Write test data to the temporary test file
	_, err = tempFile.WriteString(`
version: 2
domains:
  - example.com
macro:
  test:
    - send hello
    - wait 5
`)
	require.NoError(t, err)

	// Load macro from temporary test file
	_, err = LoadFromFile(tempFile.Name())
	assert.ErrorContains(t, err, "unsupported macro version")
}

func TestLoadFromFile_NotExists(t *testing.T) {
	_, err := LoadFromFile("/tmp/TestLoadFromFile_NotExists.yaml")
	if err == nil {
		t.Fatalf("LoadFromFile() error = %v, want non-nil", err)
	}
}

func TestMacro_GetNames(t *testing.T) {
	tests := []struct {
		name  string
		macro *Repo
		want  []string
	}{
		{
			name: "empty macro",
			macro: &Repo{
				macro: map[string]*command.Templates{},
			},
			want: []string{},
		},
		{
			name: "single command macro",
			macro: &Repo{
				macro: map[string]*command.Templates{
					"test": nil,
				},
			},
			want: []string{"test"},
		},
		{
			name: "multiple command macro",
			macro: &Repo{
				macro: map[string]*command.Templates{
					"command1": nil,
					"command2": nil,
					"command3": nil,
				},
			},
			want: []string{"command1", "command2", "command3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.macro.GetNames()
			assert.ElementsMatch(t, tt.want, got, "GetNames() = %v, want %v", got, tt.want)
		})
	}
}

func TestMacro_LoadMacroForDomain(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(macroDir string) // setup function to prepare test environment
		domain      string
		expectedErr string
		expectedLen int // expected number of macros loaded
	}{
		{
			name: "no files in directory",
			setup: func(_ string) {
				// No setup needed, empty directory suffices
			},
			domain:      "example.com",
			expectedErr: "",
			expectedLen: 0,
		},
		{
			name: "matching domain in YAML file",
			setup: func(macroDir string) {
				err := os.WriteFile(macroDir+"/macro1.yaml", []byte(`
version: 1
domains:
  - example.com
macro:
  test:
    - send hello
    - wait 5
`), 0o600)

				assert.NoError(t, err)
			},
			domain:      "example.com",
			expectedErr: "",
			expectedLen: 1,
		},
		{
			name: "no matching domain in files",
			setup: func(macroDir string) {
				err := os.WriteFile(macroDir+"/macro1.yaml", []byte(`
version: 1
domains:
  - otherdomain.com
macro:
  test:
    - send hello
    - wait 5
`), 0o600)
				assert.NoError(t, err)
			},
			domain:      "example.com",
			expectedErr: "",
			expectedLen: 0,
		},
		{
			name: "invalid file contents",
			setup: func(macroDir string) {
				err := os.WriteFile(macroDir+"/macro1.yaml", []byte("Invalid YAML content"), 0o600)
				assert.NoError(t, err)
			},
			domain:      "example.com",
			expectedErr: "yaml: unmarshal errors",
			expectedLen: 0,
		},
		{
			name: "multiple files with partially matching domains",
			setup: func(macroDir string) {
				err := os.WriteFile(macroDir+"/macro1.yaml", []byte(`
version: 1
domains:
  - example.com
macro:
  test:
    - send hello
    - wait 5
`), 0o600)
				assert.NoError(t, err)
				err = os.WriteFile(macroDir+"/macro2.yaml", []byte(`
version: 1
domains:
  - anotherdomain.com
macro:
  other:
    - edit world
`), 0o600)
				assert.NoError(t, err)
			},
			domain:      "example.com",
			expectedErr: "",
			expectedLen: 1,
		},
		{
			name: "merge macros successfully",
			setup: func(macroDir string) {
				err := os.WriteFile(macroDir+"/macro1.yaml", []byte(`
version: 1
domains:
  - example.com
macro:
  test1:
    - send hello
`), 0o600)
				assert.NoError(t, err)
				err = os.WriteFile(macroDir+"/macro2.yaml", []byte(`
version: 1
domains:
  - example.com
macro:
  test2:
    - wait 5
`), 0o600)
				assert.NoError(t, err)
			},
			domain:      "example.com",
			expectedErr: "",
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			macroDir := t.TempDir()

			if tt.setup != nil {
				tt.setup(macroDir)
			}

			got, err := LoadMacroForDomain(macroDir, tt.domain)

			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedErr)
			}

			if tt.expectedLen == 0 {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, tt.expectedLen, len(got.macro))
			}
		})
	}
}

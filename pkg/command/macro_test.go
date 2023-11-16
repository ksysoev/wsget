package command

import (
	"testing"
)

func TestNewMacro(t *testing.T) {
	tests := []struct {
		name    string
		want    *Macro
		domains []string
	}{
		{
			name:    "empty domains",
			domains: []string{},
			want: &Macro{
				macro:   make(map[string]Executer),
				domains: []string{},
			},
		},
		{
			name:    "non-empty domains",
			domains: []string{"example.com", "google.com"},
			want: &Macro{
				macro:   make(map[string]Executer),
				domains: []string{"example.com", "google.com"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMacro(tt.domains); got == nil {
				t.Errorf("NewMacro() = %v, want non-nil", got)
			} else if len(got.macro) != 0 {
				t.Errorf("NewMacro() = %v, want empty macro map", got)
			} else if len(got.domains) != len(tt.domains) {
				t.Errorf("NewMacro() = %v, want domains %v", got, tt.domains)
			}
		})
	}
}
func TestMacro_AddCommands(t *testing.T) {
	tests := []struct {
		name        string
		macro       *Macro
		commandName string
		commands    []string
		wantErr     bool
	}{
		{
			name:        "add new macro",
			macro:       NewMacro([]string{}),
			commandName: "test",
			commands:    []string{"edit hello"},
			wantErr:     false,
		},
		{
			name:        "add existing macro",
			macro:       &Macro{macro: map[string]Executer{"test": nil}},
			commandName: "test",
			commands:    []string{"send hello"},
			wantErr:     true,
		},
		{
			name:        "empty macro",
			macro:       NewMacro([]string{}),
			commandName: "test",
			commands:    []string{},
			wantErr:     true,
		},
		{
			name:        "single command macro",
			macro:       NewMacro([]string{}),
			commandName: "test",
			commands:    []string{"exit hello"},
			wantErr:     false,
		},
		{
			name:        "multi command macro",
			macro:       NewMacro([]string{}),
			commandName: "test",
			commands:    []string{"send hello", "wait 5"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.macro.AddCommands(tt.commandName, tt.commands)
			if (err != nil) != tt.wantErr {
				t.Errorf("Macro.AddCommands() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func TestMacro_Merge(t *testing.T) {
	tests := []struct {
		macro       *Macro
		otherMacro  *Macro
		name        string
		expectedLen int
		wantErr     bool
	}{
		{
			name: "merge empty macro with empty macro",
			macro: &Macro{
				macro:   make(map[string]Executer),
				domains: []string{},
			},
			otherMacro: &Macro{
				macro:   make(map[string]Executer),
				domains: []string{},
			},
			wantErr:     false,
			expectedLen: 0,
		},
		{
			name: "merge non-empty macro with empty macro",
			macro: &Macro{
				macro: map[string]Executer{
					"test": nil,
				},
				domains: []string{},
			},
			otherMacro: &Macro{
				macro:   make(map[string]Executer),
				domains: []string{},
			},
			wantErr:     false,
			expectedLen: 1,
		},
		{
			name: "merge empty macro with non-empty macro",
			macro: &Macro{
				macro:   make(map[string]Executer),
				domains: []string{},
			},
			otherMacro: &Macro{
				macro: map[string]Executer{
					"test": nil,
				},
				domains: []string{},
			},
			wantErr:     false,
			expectedLen: 1,
		},
		{
			name: "merge non-empty macro with non-empty macro",
			macro: &Macro{
				macro: map[string]Executer{
					"test": nil,
				},
				domains: []string{},
			},
			otherMacro: &Macro{
				macro: map[string]Executer{
					"test2": nil,
				},
				domains: []string{},
			},
			wantErr:     false,
			expectedLen: 2,
		},
		{
			name: "merge macro with duplicate macro name",
			macro: &Macro{
				macro: map[string]Executer{
					"test": nil,
				},
				domains: []string{},
			},
			otherMacro: &Macro{
				macro: map[string]Executer{
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
				t.Errorf("Macro.merge() error = %v, wantErr %v", err, tt.wantErr)
			} else if len(tt.macro.macro) != tt.expectedLen {
				t.Errorf("Macro.merge() expected length of macro map = %d, got %d", tt.expectedLen, len(tt.macro.macro))
			}
		})
	}
}
func TestMacro_Get(t *testing.T) {
	tests := []struct {
		name    string
		macro   *Macro
		cmdName string
		wantCmd Executer
		errMsg  string
		wantErr bool
	}{
		{
			name:    "get existing command",
			macro:   &Macro{macro: map[string]Executer{"test": NewExit()}},
			cmdName: "test",
			wantCmd: NewExit(),
			wantErr: false,
			errMsg:  "",
		},
		{
			name:    "get non-existing command",
			macro:   &Macro{macro: map[string]Executer{}},
			cmdName: "test",
			wantCmd: nil,
			wantErr: true,
			errMsg:  "unknown command: test",
		},
		{
			name:    "get command with empty macro",
			macro:   &Macro{macro: map[string]Executer{}},
			cmdName: "",
			wantCmd: nil,
			wantErr: true,
			errMsg:  "unknown command: ",
		},
		{
			name:    "get command with non-empty macro",
			macro:   &Macro{macro: map[string]Executer{"test": nil}},
			cmdName: "",
			wantCmd: nil,
			wantErr: true,
			errMsg:  "unknown command: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := tt.macro.Get(tt.cmdName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Macro.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("Macro.Get() error message = %v, want %v", err.Error(), tt.errMsg)
			}
			if cmd != tt.wantCmd {
				t.Errorf("Macro.Get() cmd = %v, want %v", cmd, tt.wantCmd)
			}
		})
	}
}

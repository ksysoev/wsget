package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMacroTemplates(t *testing.T) {
	tests := []struct {
		name      string
		templates []string
		wantErr   bool
	}{
		{
			name:      "ValidTemplates",
			templates: []string{"{{.Value}}", "{{.Name}}: {{.Age}}"},
			wantErr:   false,
		},
		{
			name:      "InvalidTemplate",
			templates: []string{"{{.Value}", "{{.Name}}: {{.Age}}"},
			wantErr:   true,
		},
		{
			name:      "EmptyTemplates",
			templates: []string{},
			wantErr:   false,
		},
		{
			name:      "NilTemplates",
			templates: nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result, err := NewMacro(tt.templates)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.list, len(tt.templates))
			}
		})
	}
}

func TestTemplates_GetExecuter(t *testing.T) {
	tests := []struct {
		name         string
		templates    []string
		args         []string
		wantErr      bool
		expectedCmds int
	}{
		{
			name:         "SingleValidTemplate",
			templates:    []string{"send {{index .Args 0}}"},
			args:         []string{"Hello"},
			wantErr:      false,
			expectedCmds: 1,
		},
		{
			name:         "MultipleValidTemplates",
			templates:    []string{"send {{index .Args 0}}", "send {{index .Args 0}}"},
			args:         []string{"World"},
			wantErr:      false,
			expectedCmds: 2,
		},
		{
			name:         "TemplateExecutionError",
			templates:    []string{"echo {{.InvalidKey}}"},
			args:         []string{"Test"},
			wantErr:      true,
			expectedCmds: 0,
		},
		{
			name:         "CommandCreationError",
			templates:    []string{"invalid-command {{index .Args 0}}"},
			args:         []string{"Test"},
			wantErr:      true,
			expectedCmds: 0,
		},
		{
			name:         "EmptyTemplates",
			templates:    []string{},
			args:         []string{"Test"},
			wantErr:      false,
			expectedCmds: 0,
		},
		{
			name:         "NilTemplates",
			templates:    nil,
			args:         []string{"Test"},
			wantErr:      false,
			expectedCmds: 0,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			templates, err := NewMacro(tt.templates)
			assert.NoError(t, err)

			// Act
			executer, err := templates.GetExecuter(tt.args)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, executer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, executer)

				if tt.expectedCmds > 1 {
					_, ok := executer.(*Sequence)
					assert.True(t, ok)
				}
			}
		})
	}
}

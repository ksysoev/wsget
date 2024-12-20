package macro

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
			t.Parallel()

			// Act
			result, err := NewMacroTemplates(tt.templates)

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

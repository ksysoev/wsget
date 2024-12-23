package macro

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		validateErr error
		expectedErr string
	}{
		{
			name: "valid config",
			input: `
version: 1
domains: ["example.com"]
macro:
  test: ["exit"]
`,
		},
		{
			name:        "invalid YAML format",
			input:       "key: : value",
			expectedErr: "yaml: mapping values are not allowed in this context",
		},
		{
			name:        "validation error",
			input:       `version: 2`,
			expectedErr: "unsupported macro version: 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			buf.WriteString(tt.input)

			// Act
			result, err := newConfig(&buf)

			// Assert
			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestConfig_SetSource(t *testing.T) {
	// Arrange
	c := &config{}

	// Act
	c.SetSource("test")

	// Assert
	assert.Equal(t, "test", c.Source)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *config
		expectedErr string
	}{
		{
			name: "valid config",
			config: &config{
				Version: "1",
				Domains: []string{"example.com"},
				Macro: map[string][]string{
					"test": {"exit"},
				},
			},
		},
		{
			name: "unsupported version",
			config: &config{
				Version: "2",
			},
			expectedErr: "unsupported macro version: 2",
		},
		{
			name: "missing domains",
			config: &config{
				Version: "1",
			},
			expectedErr: "domains are required",
		},
		{
			name: "missing macro commands",
			config: &config{
				Version: "1",
				Domains: []string{"example.com"},
			},
			expectedErr: "macro commands are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.config.validate()

			// Assert
			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_CreateRepo(t *testing.T) {
	tests := []struct {
		name    string
		config  *config
		wantErr string
	}{
		{
			name: "valid config with commands",
			config: &config{
				Macro: map[string][]string{"test": {"exit"}},
			},
		},
		{
			name: "error adding commands",
			config: &config{
				Macro: map[string][]string{"test": {"invalid {{ command }"}},
			},
			wantErr: "fail to add macro: template: macro:1: function \"command\" not defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			repo, err := tt.config.CreateRepo()

			// Assert
			if tt.wantErr != "" {
				assert.Nil(t, repo)
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NotNil(t, repo)
				assert.NoError(t, err)
			}
		})
	}
}

// mockWriter is a mock implementation of io.Writer to simulate writing scenarios.
type mockWriter struct {
	output string
	fail   bool
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	if m.fail {
		return 0, assert.AnError
	}
	m.output += string(p)
	return len(p), nil
}

func TestConfig_Write(t *testing.T) {
	tests := []struct {
		name       string
		config     *config
		wantOutput string
		wantErr    error
	}{
		{
			name: "valid config writes successfully",
			config: &config{
				Version: "1",
				Domains: []string{"example.com"},
				Macro:   map[string][]string{"test": {"exit"}},
			},
			wantOutput: `version: "1"
macro:
    test:
        - exit
domains:
    - example.com
`,
		},
		{
			name: "empty config writes successfully",
			config: &config{
				Version: "",
				Domains: nil,
				Macro:   nil,
			},
			wantOutput: "version: \"\"\nmacro: {}\ndomains: []\n",
		},
		{
			name: "error during writing",
			config: &config{
				Version: "1",
				Domains: []string{"example.com"},
				Macro:   map[string][]string{"test": {"exit"}},
			},
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			writer := &mockWriter{}
			if tt.wantErr != nil {
				writer.fail = true
			}

			// Act
			err := tt.config.Write(writer)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOutput, writer.output)
			}
		})
	}
}

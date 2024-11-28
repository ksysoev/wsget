package cmd

import (
	"os"
	"testing"
	"time"

	"github.com/ksysoev/wsget/pkg/cli"
	"github.com/ksysoev/wsget/pkg/command"
	"github.com/stretchr/testify/assert"
)

func TestCreateCommands(t *testing.T) {
	tmpDir := os.TempDir()

	tests := []struct {
		name     string
		args     *flags
		expected []command.Executer
	}{
		{
			name: "Request with waitResponse",
			args: &flags{
				request:      "test request",
				waitResponse: 5,
			},
			expected: []command.Executer{
				command.NewSend("test request"),
				command.NewWaitForResp(5 * time.Second),
				command.NewExit(),
			},
		},
		{
			name: "Request without waitResponse",
			args: &flags{
				request:      "test request",
				waitResponse: -1,
			},
			expected: []command.Executer{
				command.NewSend("test request"),
			},
		},
		{
			name: "InputFile",
			args: &flags{
				inputFile: tmpDir + "/testfile.txt",
			},
			expected: []command.Executer{
				command.NewInputFileCommand(tmpDir + "/testfile.txt"),
			},
		},
		{
			name: "Default Edit",
			args: &flags{},
			expected: []command.Executer{
				command.NewEdit(""),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createCommands(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInitRunOptions(t *testing.T) {
	tmpDir := os.TempDir()

	tests := []struct {
		name        string
		args        *flags
		expected    *cli.RunOptions
		expectError bool
	}{
		{
			name: "OutputFile and Request",
			args: &flags{
				outputFile:   tmpDir + "/test_output.txt",
				request:      "test request",
				waitResponse: -1,
			},
			expected: &cli.RunOptions{
				OutputFile: func() *os.File {
					f, _ := os.Create(tmpDir + "/test_output.txt")
					return f
				}(),
				Commands: []command.Executer{
					command.NewSend("test request"),
				},
			},
			expectError: false,
		},
		{
			name: "Invalid OutputFile",
			args: &flags{
				outputFile: "/invalid/path/test_output.txt",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Request with waitResponse",
			args: &flags{
				request:      "test request",
				waitResponse: 5,
			},
			expected: &cli.RunOptions{
				Commands: []command.Executer{
					command.NewSend("test request"),
					command.NewWaitForResp(5 * time.Second),
					command.NewExit(),
				},
			},
			expectError: false,
		},
		{
			name: "InputFile",
			args: &flags{
				inputFile: tmpDir + "/testfile.txt",
			},
			expected: &cli.RunOptions{
				Commands: []command.Executer{
					command.NewInputFileCommand(tmpDir + "/testfile.txt"),
				},
			},
			expectError: false,
		},
		{
			name: "Default Edit",
			args: &flags{},
			expected: &cli.RunOptions{
				Commands: []command.Executer{
					command.NewEdit(""),
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := initRunOptions(tt.args)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Commands, opts.Commands)
				if tt.expected.OutputFile != nil {
					assert.NotNil(t, opts.OutputFile)
					opts.OutputFile.Close()
				}
			}
		})
	}
}

func TestValidateArgs(t *testing.T) {
	tests := []struct {
		name        string
		wsURL       string
		args        *flags
		expectedErr string
	}{
		{
			name:        "Empty URL",
			wsURL:       "",
			args:        &flags{},
			expectedErr: "url is required",
		},
		{
			name:  "WaitResponse without Request",
			wsURL: "ws://example.com",
			args: &flags{
				waitResponse: 5,
				request:      "",
			},
			expectedErr: "single response timeout could be used only with request",
		},
		{
			name:  "Valid Arguments",
			wsURL: "ws://example.com",
			args: &flags{
				waitResponse: 5,
				request:      "test request",
			},
			expectedErr: "",
		},
		{
			name:  "Valid Arguments without WaitResponse",
			wsURL: "ws://example.com",
			args: &flags{
				waitResponse: -1,
				request:      "test request",
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArgs(tt.wsURL, tt.args)
			if tt.expectedErr != "" {
				assert.EqualError(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

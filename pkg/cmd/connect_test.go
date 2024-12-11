package cmd

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/ksysoev/wsget/pkg/core"
	"github.com/ksysoev/wsget/pkg/core/command"
	"github.com/stretchr/testify/assert"
)

func createEchoWSHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}

		defer c.Close(websocket.StatusNormalClosure, "")

		for {
			_, wsr, err := c.Reader(r.Context())
			if err != nil {
				if err == io.EOF {
					return
				}

				return
			}

			wsw, err := c.Writer(r.Context(), websocket.MessageText)
			if err != nil {
				return
			}

			if _, err := io.Copy(wsw, wsr); err != nil {
				return
			}

			if err := wsw.Close(); err != nil {
				return
			}
		}
	})
}

func TestCreateCommands(t *testing.T) {
	tmpDir := os.TempDir()

	tests := []struct {
		name     string
		args     *flags
		expected []core.Executer
	}{
		{
			name: "Request with waitResponse",
			args: &flags{
				request:      "test request",
				waitResponse: 5,
			},
			expected: []core.Executer{
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
			expected: []core.Executer{
				command.NewSend("test request"),
			},
		},
		{
			name: "InputFile",
			args: &flags{
				inputFile: tmpDir + "/testfile.txt",
			},
			expected: []core.Executer{
				command.NewInputFileCommand(tmpDir + "/testfile.txt"),
			},
		},
		{
			name: "Default Edit",
			args: &flags{},
			expected: []core.Executer{
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
		args        *flags
		expected    *core.RunOptions
		name        string
		expectError bool
	}{
		{
			name: "OutputFile and Request",
			args: &flags{
				outputFile:   tmpDir + "/test_output.txt",
				request:      "test request",
				waitResponse: -1,
			},
			expected: &core.RunOptions{
				OutputFile: func() *os.File {
					f, _ := os.Create(tmpDir + "/test_output.txt")
					return f
				}(),
				Commands: []core.Executer{
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
			expected: &core.RunOptions{
				Commands: []core.Executer{
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
			expected: &core.RunOptions{
				Commands: []core.Executer{
					command.NewInputFileCommand(tmpDir + "/testfile.txt"),
				},
			},
			expectError: false,
		},
		{
			name: "Default Edit",
			args: &flags{},
			expected: &core.RunOptions{
				Commands: []core.Executer{
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

func TestCreateConnectRunner(t *testing.T) {
	runner := createConnectRunner(&flags{})
	assert.NotNil(t, runner)
}
func TestRunConnectCmd_FailToConnect(t *testing.T) {
	ctx := context.Background()
	err := runConnectCmd(ctx, &flags{}, []string{"ws://localhost:0"})
	assert.Error(t, err)
}

func TestRunConnectCmd_NoURL(t *testing.T) {
	ctx := context.Background()
	args := &flags{
		request: "test request",
	}
	err := runConnectCmd(ctx, args, []string{""})
	assert.Error(t, err)
}

func TestRunConnectCmd_SuccessConnect(t *testing.T) {
	server := httptest.NewServer(createEchoWSHandler())
	defer server.Close()

	url := "ws://" + server.Listener.Addr().String()

	ctx := context.Background()
	args := &flags{
		request:      "test request",
		waitResponse: 1,
	}

	// tty is not available in the test environment
	// so the test will fail in some cases and be successful in others
	err := runConnectCmd(ctx, args, []string{url})

	if err != nil {
		assert.ErrorContains(t, err, "keyboard run error: open /dev/tty: ")
	} else {
		assert.NoError(t, err)
	}
}

package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestNewExecutionContext(t *testing.T) {
	tests := []struct {
		cli        *CLI
		outputFile *bytes.Buffer
		name       string
	}{
		{
			name: "Valid CLI and OutputFile",
			cli: &CLI{
				inputStream: make(chan KeyEvent),
			},
			outputFile: &bytes.Buffer{},
		},
		{
			name:       "Nil CLI",
			cli:        nil,
			outputFile: &bytes.Buffer{},
		},
		{
			name:       "Nil OutputFile",
			cli:        &CLI{},
			outputFile: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executionContext := newExecutionContext(context.Background(), tt.cli, tt.outputFile)
			assert.Equal(t, tt.cli, executionContext.cli, "CLI should match the input CLI")
			assert.Equal(t, tt.outputFile, executionContext.outputFile, "Output file should match the input outputFile")
		})
	}
}

func TestExecutionContext_SendRequest(t *testing.T) {
	tests := []struct {
		name        string
		setupCLI    func(ctx context.Context) *CLI
		req         string
		expectError bool
	}{
		{
			name: "Valid request",
			setupCLI: func(ctx context.Context) *CLI {
				mockWsConn := NewMockConnectionHandler(t)
				mockWsConn.EXPECT().Send(ctx, "valid request").Return(nil)

				return &CLI{
					wsConn: mockWsConn,
				}
			},
			req:         "valid request",
			expectError: false,
		},
		{
			name: "Send failure",
			setupCLI: func(ctx context.Context) *CLI {
				mockWsConn := NewMockConnectionHandler(t)
				mockWsConn.EXPECT().Send(ctx, "invalid request").Return(fmt.Errorf("send error"))

				return &CLI{
					wsConn: mockWsConn,
				}
			},
			req:         "invalid request",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cli := tt.setupCLI(ctx)
			ec := &executionContext{
				cli: cli,
				ctx: ctx,
			}

			err := ec.SendRequest(tt.req)
			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Did not expect an error")
			}
		})
	}
}

func TestExecutionContext_Print(t *testing.T) {
	tests := []struct {
		setupCLI    func() *CLI
		name        string
		data        string
		attributes  []color.Attribute
		expectError bool
	}{
		{
			name:       "Valid case with no attributes",
			data:       "test data",
			attributes: nil,
			setupCLI: func() *CLI {
				return &CLI{
					output: &bytes.Buffer{},
				}
			},
		},
		{
			name:       "Valid case with attributes",
			data:       "colored data",
			attributes: []color.Attribute{color.FgBlue, color.Bold},
			setupCLI: func() *CLI {
				return &CLI{
					output: &bytes.Buffer{},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := tt.setupCLI()
			ec := &executionContext{
				cli: cli,
			}

			err := ec.Print(tt.data, tt.attributes...)
			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Did not expect an error")

				if cli.output != nil {
					output := cli.output.(*bytes.Buffer).String()
					assert.Contains(t, output, tt.data, "Expected output to contain data")
				}
			}
		})
	}
}

func TestExecutionContext_EditorMode(t *testing.T) {
	mockEditor := NewMockEditor(t)
	ctx := context.Background()

	mockEditor.EXPECT().Edit(ctx, "test").Return("test", nil)

	ec := &executionContext{
		ctx: ctx,
		cli: &CLI{
			editor: mockEditor,
		},
	}

	res, err := ec.EditorMode("test")
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, "test", res, "Expected response to match")
}

func TestExecutionContext_CommandMode(t *testing.T) {
	mockEditor := NewMockEditor(t)
	ctx := context.Background()

	mockEditor.EXPECT().CommandMode(ctx, "test").Return("test", nil)

	ec := &executionContext{
		ctx: ctx,
		cli: &CLI{
			editor: mockEditor,
		},
	}

	res, err := ec.CommandMode("test")
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, "test", res, "Expected response to match")
}

func TestExecutionContext_CreateCommand(t *testing.T) {
	mockFactory := NewMockCommandFactory(t)
	ctx := context.Background()

	expectCmd := NewMockExecuter(t)

	mockFactory.EXPECT().Create("test").Return(expectCmd, nil)

	ec := &executionContext{
		ctx: ctx,
		cli: &CLI{
			cmdFactory: mockFactory,
		},
	}

	cmd, err := ec.CreateCommand("test")
	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, expectCmd, cmd, "Expected command to match")
}

func TestExecutionContext_WaitForResponse(t *testing.T) {
	tests := []struct {
		setupCLI       func(ctx context.Context) *CLI
		name           string
		expectedResult Message
		timeout        time.Duration
		expectError    bool
	}{
		{
			name:    "Valid response within timeout",
			timeout: 2 * time.Second,
			setupCLI: func(_ context.Context) *CLI {
				msgChan := make(chan Message, 1)
				msgChan <- Message{Type: Response, Data: "Response Data"}

				return &CLI{
					messages: msgChan,
				}
			},
			expectedResult: Message{Type: Response, Data: "Response Data"},
			expectError:    false,
		},
		{
			name:    "Timeout exceeded",
			timeout: 1 * time.Millisecond,
			setupCLI: func(_ context.Context) *CLI {
				msgChan := make(chan Message, 1)
				go func() {
					time.Sleep(2 * time.Millisecond)
					msgChan <- Message{Type: Response, Data: "Response Data"}
				}()

				return &CLI{
					messages: msgChan,
				}
			},
			expectedResult: Message{},
			expectError:    true,
		},
		{
			name:    "Error from CLI",
			timeout: 1 * time.Millisecond,
			setupCLI: func(_ context.Context) *CLI {
				msgChan := make(chan Message, 1)
				go func() {
					time.Sleep(2 * time.Millisecond)
					msgChan <- Message{Type: Response, Data: "Response Data"}
				}()

				return &CLI{
					messages: msgChan,
				}
			},
			expectedResult: Message{},
			expectError:    true,
		},
		{
			name:    "Zero timeout with valid response",
			timeout: 0,
			setupCLI: func(_ context.Context) *CLI {
				msgChan := make(chan Message, 1)
				msgChan <- Message{Type: Request, Data: "Immediate Response"}

				return &CLI{
					messages: msgChan,
				}
			},
			expectedResult: Message{Type: Request, Data: "Immediate Response"},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cli := tt.setupCLI(ctx)
			ec := &executionContext{
				cli: cli,
				ctx: ctx,
			}

			result, err := ec.WaitForResponse(tt.timeout)
			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				assert.NoError(t, err, "Did not expect an error but got one")
				assert.Equal(t, tt.expectedResult, result, "Expected result to match")
			}
		})
	}
}

func TestExecutionContext_PrintToFile(t *testing.T) {
	tests := []struct {
		setupOutput    func() io.Writer
		name           string
		data           string
		expectedOutput string
		expectedError  bool
	}{
		{
			name: "Valid output file",
			setupOutput: func() io.Writer {
				return &bytes.Buffer{}
			},
			data:           "test data",
			expectedError:  false,
			expectedOutput: "test data\n",
		},
		{
			name: "Nil output file",
			setupOutput: func() io.Writer {
				return nil
			},
			data:           "test data",
			expectedError:  false,
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.setupOutput()

			ec := &executionContext{
				outputFile: output,
			}

			err := ec.PrintToFile(tt.data)
			if tt.expectedError {
				assert.Error(t, err, "Expected an error but didn't get one")
			} else {
				assert.NoError(t, err, "Did not expect an error but got one")
			}

			if buf, ok := output.(*bytes.Buffer); ok {
				assert.Equal(t, tt.expectedOutput, buf.String(), "Unexpected output in buffer")
			}
		})
	}
}

func TestExecutionContext_FormatMessage(t *testing.T) {
	tests := []struct {
		setupCLI    func() *CLI
		name        string
		expected    string
		message     Message
		noColor     bool
		expectError bool
	}{
		{
			name:    "Successful formatting for file without color",
			message: Message{Type: Response, Data: "File formatting"},
			noColor: true,
			setupCLI: func() *CLI {
				mockFormatter := NewMockFormater(t)
				mockFormatter.EXPECT().FormatForFile("Response", "File formatting").Return("Formatted for file", nil)

				return &CLI{
					formater: mockFormatter,
				}
			},
			expectError: false,
			expected:    "Formatted for file",
		},
		{
			name:    "Error during formatting for file without color",
			message: Message{Type: Request, Data: "File error case"},
			noColor: true,
			setupCLI: func() *CLI {
				mockFormatter := NewMockFormater(t)
				mockFormatter.EXPECT().FormatForFile("Request", "File error case").Return("", fmt.Errorf("formatting error"))

				return &CLI{
					formater: mockFormatter,
				}
			},
			expectError: true,
			expected:    "",
		},
		{
			name:    "Successful formatting for message with color",
			message: Message{Type: Response, Data: "Colored message"},
			noColor: false,
			setupCLI: func() *CLI {
				mockFormatter := NewMockFormater(t)
				mockFormatter.EXPECT().FormatMessage("Response", "Colored message").Return("Formatted with color", nil)

				return &CLI{
					formater: mockFormatter,
				}
			},
			expectError: false,
			expected:    "Formatted with color",
		},
		{
			name:    "Error during formatting for message with color",
			message: Message{Type: Request, Data: "Colored error case"},
			noColor: false,
			setupCLI: func() *CLI {
				mockFormatter := NewMockFormater(t)
				mockFormatter.EXPECT().FormatMessage("Request", "Colored error case").Return("", fmt.Errorf("color formatting error"))

				return &CLI{
					formater: mockFormatter,
				}
			},
			expectError: true,
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := tt.setupCLI()
			ec := &executionContext{
				cli: cli,
			}

			result, err := ec.FormatMessage(tt.message, tt.noColor)
			if tt.expectError {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Did not expect error but got one")
				assert.Equal(t, tt.expected, result, "Expected formatted message does not match")
			}
		})
	}
}

package core

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestNewExecutionContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cli        *CLI
		outputFile *bytes.Buffer
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
			executionContext := newExecutionContext(tt.cli, tt.outputFile)
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
	t.Parallel()

	tests := []struct {
		name        string
		data        string
		attributes  []color.Attribute
		setupCLI    func() *CLI
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

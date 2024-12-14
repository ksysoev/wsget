package command

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/ksysoev/wsget/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExit_Execute(t *testing.T) {
	c := NewExit()
	_, err := c.Execute(nil)

	if err == nil {
		t.Errorf("Exit.Execute() error = %v, wantErr %v", err, true)
	}

	if !errors.Is(err, core.ErrInterrupted) {
		t.Errorf("Exit.Execute() error = %v, wantErr interrupted", err)
	}
}

func TestPrintMsg_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		message          core.Message
		mockFormatError  error
		mockFormatOutput string
		mockPrintError   error
		expectedErr      string
	}{
		{
			name: "RequestMessage_Success",
			message: core.Message{
				Type: core.Request,
				Data: "test request",
			},
			mockFormatError:  nil,
			mockFormatOutput: "formatted request",
			mockPrintError:   nil,
			expectedErr:      "",
		},
		{
			name: "ResponseMessage_Success",
			message: core.Message{
				Type: core.Response,
				Data: "test response",
			},
			mockFormatError:  nil,
			mockFormatOutput: "formatted response",
			mockPrintError:   nil,
			expectedErr:      "",
		},
		{
			name: "UnsupportedMessageType",
			message: core.Message{
				Type: core.MessageType(3),
				Data: "unsupported",
			},
			expectedErr: "unsupported message type",
		},
		{
			name: "FormatMessage_Error",
			message: core.Message{
				Type: core.Request,
				Data: "test request",
			},
			mockFormatError: fmt.Errorf("format error"),
			expectedErr:     "fail to format message: format error",
		},
		{
			name: "Print_Error",
			message: core.Message{
				Type: core.Request,
				Data: "test request",
			},
			mockFormatError:  nil,
			mockFormatOutput: "formatted request",
			mockPrintError:   fmt.Errorf("print error"),
			expectedErr:      "fail to print message: print error",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			exCtx := core.NewMockExecutionContext(t)
			exCtx.EXPECT().
				FormatMessage(tt.message, false).
				Return(tt.mockFormatOutput, tt.mockFormatError).
				Maybe()

			exCtx.EXPECT().
				FormatMessage(tt.message, true).
				Return(tt.mockFormatOutput, tt.mockFormatError).
				Maybe()

			if tt.mockFormatError == nil {
				switch tt.message.Type {
				case core.Request:
					exCtx.EXPECT().
						Print("->", color.FgGreen).
						Return(tt.mockPrintError).
						Maybe()
				case core.Response:
					exCtx.EXPECT().
						Print("<-", color.FgRed).
						Return(tt.mockPrintError).
						Maybe()
				}
			}

			if tt.mockPrintError == nil && tt.mockFormatError == nil {
				exCtx.EXPECT().
					Print("%s\n" + tt.mockFormatOutput).
					Return(tt.mockPrintError).
					Maybe()
				exCtx.EXPECT().
					PrintToFile(tt.mockFormatOutput).
					Return(tt.mockPrintError).
					Maybe()
			}

			cmd := NewPrintMsg(tt.message)
			_, err := cmd.Execute(exCtx)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCmdEdit_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		mockPrintError      error
		mockPrintAfterError error
		mockCommandError    error
		mockCreateCmd       core.Executer
		mockCreateCmdErr    error
		expectedNextCmd     core.Executer
		expectedErr         error
		mockRawCommand      string
	}{
		{
			name:             "ValidCommand",
			mockPrintError:   nil,
			mockCommandError: nil,
			mockRawCommand:   "test-command",
			mockCreateCmd:    NewPrintMsg(core.Message{Type: core.Request, Data: "mock"}),
			expectedNextCmd:  NewPrintMsg(core.Message{Type: core.Request, Data: "mock"}),
			expectedErr:      nil,
		},
		{
			name:             "CreateCommandError",
			mockPrintError:   nil,
			mockCommandError: nil,
			mockRawCommand:   "invalid-command",
			mockCreateCmd:    nil,
			mockCreateCmdErr: assert.AnError,
			expectedNextCmd:  nil,
			expectedErr:      nil,
		},
		{
			name:             "EmptyRawCommand",
			mockPrintError:   nil,
			mockCommandError: nil,
			mockRawCommand:   "",
			mockCreateCmd:    nil,
			mockCreateCmdErr: nil,
			expectedNextCmd:  nil,
			expectedErr:      nil, // Assuming it's valid to return no command or error.
		},
		{
			name:             "PrintErrorAtStart",
			mockPrintError:   assert.AnError,
			mockCommandError: nil,
			mockRawCommand:   "test-command",
			mockCreateCmd:    nil,
			mockCreateCmdErr: nil,
			expectedNextCmd:  nil,
			expectedErr:      assert.AnError,
		},
		{
			name:                "PrintErrorOnCursorHide",
			mockPrintError:      nil,
			mockCommandError:    nil,
			mockPrintAfterError: assert.AnError,
			mockRawCommand:      "test-command",
			mockCreateCmd:       nil,
			mockCreateCmdErr:    nil,
			expectedNextCmd:     nil,
			expectedErr:         assert.AnError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			exCtx := core.NewMockExecutionContext(t)
			exCtx.EXPECT().Print(":\x1b[?25h").Return(tt.mockPrintError).Maybe()
			exCtx.EXPECT().CommandMode("").Return(tt.mockRawCommand, tt.mockCommandError).Maybe()
			exCtx.EXPECT().Print(LineClear + "\r" + HideCursor).Return(tt.mockPrintAfterError).Maybe()
			exCtx.EXPECT().CreateCommand(tt.mockRawCommand).Return(tt.mockCreateCmd, tt.mockCreateCmdErr).Maybe()
			exCtx.EXPECT().Print("Invalid command: "+tt.mockRawCommand+"\n", color.FgRed).Return(nil).Maybe()

			cmd := NewCmdEdit()
			nextCmd, err := cmd.Execute(exCtx)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, nextCmd)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedNextCmd, nextCmd)
			}
		})
	}
}

func TestNewWaitForResp_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expectedErr error
		name        string
		timeout     time.Duration
	}{
		{
			name:        "ValidTimeout",
			timeout:     5 * time.Second,
			expectedErr: nil,
		},
		{
			name:        "ZeroTimeout",
			timeout:     0,
			expectedErr: nil,
		},
		{
			name:        "ErrorTimeout",
			timeout:     5 * time.Second,
			expectedErr: errors.New("response timeout"),
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var expectedMsg core.Message
			if tt.expectedErr == nil {
				expectedMsg = core.Message{Type: core.Response, Data: "test"}
			}

			exCtx := core.NewMockExecutionContext(t)
			exCtx.EXPECT().WaitForResponse(tt.timeout).Return(expectedMsg, tt.expectedErr)

			cmd := NewWaitForResp(tt.timeout)

			cmd1, err := cmd.Execute(exCtx)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, cmd1)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, NewPrintMsg(expectedMsg), cmd1)
			}
		})
	}
}

func TestSequence_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expectedNextCmd  core.Executer
		expectedErr      bool
		mockExecutionCtx func(t *testing.T) core.ExecutionContext
		name             string
		subCommands      []core.Executer
	}{
		{
			name:            "AllSubCommandsExecuteSuccessfully",
			subCommands:     []core.Executer{NewSleepCommand(time.Millisecond), NewSleepCommand(time.Millisecond)},
			expectedNextCmd: nil,
			expectedErr:     false,
			mockExecutionCtx: func(t *testing.T) core.ExecutionContext {
				t.Helper()

				return core.NewMockExecutionContext(t)
			},
		},
		{
			name: "SubCommandReturnsAnotherCommand",
			subCommands: []core.Executer{
				NewSleepCommand(time.Millisecond),
				NewCmdEdit(),
			},
			expectedNextCmd: nil,
			expectedErr:     false,
			mockExecutionCtx: func(t *testing.T) core.ExecutionContext {
				t.Helper()

				exCtx := core.NewMockExecutionContext(t)

				exCtx.EXPECT().Print(":\x1b[?25h").Return(nil)
				exCtx.EXPECT().CommandMode("").Return("sleep 0", nil)
				exCtx.EXPECT().Print("\x1b[2K\r\x1b[?25l").Return(nil)
				exCtx.EXPECT().CreateCommand("sleep 0").Return(NewSleepCommand(0), nil)

				return exCtx
			},
		},
		{
			name: "SubCommandFailsWithError",
			subCommands: []core.Executer{
				NewExit(),
			},
			expectedNextCmd: nil,
			expectedErr:     true,
			mockExecutionCtx: func(t *testing.T) core.ExecutionContext {
				t.Helper()

				return core.NewMockExecutionContext(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			exCtx := tt.mockExecutionCtx(t)
			seq := NewSequence(tt.subCommands)
			nextCmd, err := seq.Execute(exCtx)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, nextCmd)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedNextCmd, nextCmd)
			}
		})
	}
}

func TestRepeat_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		subCommand           core.Executer
		expectedErr          error
		mockExecutionContext func(t *testing.T) core.ExecutionContext
		name                 string
		times                int
	}{
		{
			name:        "SubCommandExecutesOnce",
			times:       1,
			subCommand:  NewSleepCommand(1 * time.Millisecond),
			expectedErr: nil,
			mockExecutionContext: func(t *testing.T) core.ExecutionContext {
				t.Helper()

				return core.NewMockExecutionContext(t)
			},
		},
		{
			name:        "SubCommandExecutesMultipleTimes",
			times:       3,
			subCommand:  NewSleepCommand(1 * time.Millisecond),
			expectedErr: nil,
			mockExecutionContext: func(t *testing.T) core.ExecutionContext {
				t.Helper()

				return core.NewMockExecutionContext(t)
			},
		},
		{
			name:        "SubCommandFails",
			times:       2,
			subCommand:  NewWaitForResp(1 * time.Millisecond),
			expectedErr: assert.AnError,
			mockExecutionContext: func(t *testing.T) core.ExecutionContext {
				t.Helper()

				exCtx := core.NewMockExecutionContext(t)
				exCtx.EXPECT().WaitForResponse(1*time.Millisecond).Return(core.Message{}, assert.AnError)
				return exCtx
			},
		},
		{
			name:        "ZeroExecutions",
			times:       0,
			subCommand:  NewPrintMsg(core.Message{Type: core.Request, Data: "skip"}),
			expectedErr: nil,
			mockExecutionContext: func(t *testing.T) core.ExecutionContext {
				t.Helper()

				// Nothing should be called
				return core.NewMockExecutionContext(t)
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			exCtx := tt.mockExecutionContext(t)
			repeatCmd := NewRepeatCommand(tt.times, tt.subCommand)

			nextCmd, err := repeatCmd.Execute(exCtx)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, nextCmd)
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Nil(t, nextCmd)
			}
		})
	}
}

func TestSleep_Execute(t *testing.T) {
	c := NewSleepCommand(1 * time.Millisecond)

	start := time.Now()
	_, err := c.Execute(nil)

	elapsed := time.Since(start)

	assert.NoError(t, err)

	if elapsed < 1*time.Millisecond {
		t.Errorf("Sleep.Execute() elapsed = %v, want >= 1ms", elapsed)
	}
}

func TestEdit_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expectedErr      error
		expectedNextCmd  core.Executer
		mockExecutionCtx func(t *testing.T) core.ExecutionContext
		name             string
		mockContent      string
	}{
		{
			name:            "SuccessfulExecution",
			mockContent:     "test-content",
			expectedErr:     nil,
			expectedNextCmd: NewSend("test-response"),
			mockExecutionCtx: func(t *testing.T) core.ExecutionContext {
				t.Helper()

				exCtx := core.NewMockExecutionContext(t)
				exCtx.EXPECT().Print("->\n"+ShowCursor, color.FgGreen).Return(nil)
				exCtx.EXPECT().EditorMode("test-content").Return("test-response", nil)
				exCtx.EXPECT().Print(LineUp + LineClear + HideCursor).Return(nil)
				return exCtx
			},
		},
		{
			name:            "EditorModeError",
			mockContent:     "error-content",
			expectedErr:     assert.AnError,
			expectedNextCmd: nil,
			mockExecutionCtx: func(t *testing.T) core.ExecutionContext {
				t.Helper()

				exCtx := core.NewMockExecutionContext(t)
				exCtx.EXPECT().Print("->\n"+ShowCursor, color.FgGreen).Return(nil)
				exCtx.EXPECT().EditorMode("error-content").Return("", assert.AnError)
				return exCtx
			},
		},
		{
			name:            "PrintError",
			mockContent:     "print-error-content",
			expectedErr:     assert.AnError,
			expectedNextCmd: nil,
			mockExecutionCtx: func(t *testing.T) core.ExecutionContext {
				t.Helper()

				exCtx := core.NewMockExecutionContext(t)
				exCtx.EXPECT().Print("->\n"+ShowCursor, color.FgGreen).Return(assert.AnError)
				return exCtx
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			exCtx := tt.mockExecutionCtx(t)
			cmd := NewEdit(tt.mockContent)

			nextCmd, err := cmd.Execute(exCtx)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, nextCmd)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedNextCmd, nextCmd)
			}
		})
	}
}

func TestSend_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expectedErr      error
		expectedNextCmd  core.Executer
		mockExecutionCtx func(t *testing.T, mockRequest string) core.ExecutionContext
		name             string
		mockRequest      string
	}{
		{
			name:        "SuccessfulExecution",
			mockRequest: "test-request",
			expectedErr: nil,
			expectedNextCmd: NewPrintMsg(core.Message{
				Type: core.Request, Data: "test-request",
			}),
			mockExecutionCtx: func(t *testing.T, mockRequest string) core.ExecutionContext {
				t.Helper()

				exCtx := core.NewMockExecutionContext(t)
				exCtx.EXPECT().SendRequest(mockRequest).Return(nil)
				return exCtx
			},
		},
		{
			name:            "SendRequestError",
			mockRequest:     "error-request",
			expectedErr:     assert.AnError,
			expectedNextCmd: nil,
			mockExecutionCtx: func(t *testing.T, mockRequest string) core.ExecutionContext {
				t.Helper()

				exCtx := core.NewMockExecutionContext(t)
				exCtx.EXPECT().SendRequest(mockRequest).Return(assert.AnError)
				return exCtx
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			exCtx := tt.mockExecutionCtx(t, tt.mockRequest)
			cmd := NewSend(tt.mockRequest)

			nextCmd, err := cmd.Execute(exCtx)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, nextCmd)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedNextCmd, nextCmd)
			}
		})
	}
}

func TestInputFileCommand_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expectedNextCmd core.Executer
		mockCreateCmd   func(cmd string) (core.Executer, error)
		prepareFile     func(t *testing.T, filePath, content string)
		cleanupFile     func(filePath string)
		name            string
		filePath        string
		fileContent     string
		expectedErr     bool
	}{
		{
			name:        "SuccessfulFileReadAndCommandExecution",
			filePath:    "test-file.yaml",
			fileContent: "- print-msg-1\n- print-msg-2\n",
			mockCreateCmd: func(cmd string) (core.Executer, error) {
				return NewPrintMsg(core.Message{Type: core.Request, Data: cmd}), nil
			},
			expectedErr: false,
			expectedNextCmd: NewSequence([]core.Executer{
				NewPrintMsg(core.Message{Type: core.Request, Data: "print-msg-1"}),
				NewPrintMsg(core.Message{Type: core.Request, Data: "print-msg-2"}),
			}),
			prepareFile: func(t *testing.T, filePath string, content string) {
				t.Helper()

				err := os.WriteFile(filePath, []byte(content), 0o600)
				assert.NoError(t, err)
			},
			cleanupFile: func(filePath string) {
				_ = os.Remove(filePath)
			},
		},
		{
			name:            "InvalidFilePath",
			filePath:        "invalid-file.yaml",
			fileContent:     "",
			mockCreateCmd:   nil,
			expectedErr:     true,
			expectedNextCmd: nil,
			prepareFile:     func(_ *testing.T, _ string, _ string) {}, // No file preparation
			cleanupFile:     func(_ string) {},                         // No cleanup needed
		},
		{
			name:        "InvalidYAMLContent",
			filePath:    "invalid-yaml-file.yaml",
			fileContent: "not-a-valid-yaml",
			mockCreateCmd: func(_ string) (core.Executer, error) {
				return nil, nil
			},
			expectedErr:     true,
			expectedNextCmd: nil,
			prepareFile: func(t *testing.T, filePath string, content string) {
				t.Helper()

				err := os.WriteFile(filePath, []byte(content), 0o600)
				assert.NoError(t, err)
			},
			cleanupFile: func(filePath string) {
				_ = os.Remove(filePath)
			},
		},
		{
			name:        "CommandCreationError",
			filePath:    "commands.yaml",
			fileContent: "- valid-command\n- invalid-command\n",
			mockCreateCmd: func(cmd string) (core.Executer, error) {
				if cmd == "valid-command" {
					return NewPrintMsg(core.Message{Type: core.Request, Data: cmd}), nil
				}
				return nil, assert.AnError
			},
			expectedErr:     true,
			expectedNextCmd: nil,
			prepareFile: func(t *testing.T, filePath string, content string) {
				t.Helper()

				err := os.WriteFile(filePath, []byte(content), 0o600)
				assert.NoError(t, err)
			},
			cleanupFile: func(filePath string) {
				_ = os.Remove(filePath)
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Prepare environment
			if tt.prepareFile != nil {
				tt.prepareFile(t, tt.filePath, tt.fileContent)
			}

			if tt.cleanupFile != nil {
				defer tt.cleanupFile(tt.filePath)
			}

			// Mock execution context
			exCtx := core.NewMockExecutionContext(t)
			if tt.mockCreateCmd != nil {
				exCtx.EXPECT().CreateCommand(mock.Anything).RunAndReturn(tt.mockCreateCmd).Maybe()
			}

			// Execute InputFileCommand
			cmd := NewInputFileCommand(tt.filePath)
			nextCmd, err := cmd.Execute(exCtx)

			// Assertions
			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, nextCmd)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedNextCmd, nextCmd)
			}
		})
	}
}

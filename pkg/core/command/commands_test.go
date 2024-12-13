package command

import (
	"errors"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/ksysoev/wsget/pkg/core"
	"github.com/stretchr/testify/assert"
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
	expectedMsg := core.Message{Type: core.Request, Data: "test"}

	exCtx := core.NewMockExecutionContext(t)
	exCtx.EXPECT().PrintMessage(expectedMsg).Return(nil)

	c := NewPrintMsg(expectedMsg)
	_, err := c.Execute(exCtx)

	if err != nil {
		t.Errorf("PrintMsg.Execute() error = %v, wantErr %v", err, nil)
	}
}

func TestCmdEdit_Execute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		mockPrintError   error
		mockCommandError error
		mockCreateCmd    core.Executer
		mockCreateCmdErr error
		expectedNextCmd  core.Executer
		expectedErr      error
		mockRawCommand   string
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			exCtx := core.NewMockExecutionContext(t)
			exCtx.EXPECT().Print(":\x1b[?25h").Return(tt.mockPrintError).Maybe()
			exCtx.EXPECT().CommandMode("").Return(tt.mockRawCommand, tt.mockCommandError).Maybe()
			exCtx.EXPECT().Print(LineClear + "\r" + HideCursor).Return(nil).Maybe()
			exCtx.EXPECT().CreateCommand(tt.mockRawCommand).Return(tt.mockCreateCmd, tt.mockCreateCmdErr).Maybe()

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
		name        string
		timeout     time.Duration
		expectedErr error
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
		name             string
		subCommands      []core.Executer
		expectedNextCmd  core.Executer
		expectedErr      error
		mockExecutionCtx func(t *testing.T) core.ExecutionContext
	}{
		{
			name:            "AllSubCommandsExecuteSuccessfully",
			subCommands:     []core.Executer{NewPrintMsg(core.Message{Type: core.Request, Data: "test1"}), NewPrintMsg(core.Message{Type: core.Request, Data: "test2"})},
			expectedNextCmd: nil,
			expectedErr:     nil,
			mockExecutionCtx: func(t *testing.T) core.ExecutionContext {
				exCtx := core.NewMockExecutionContext(t)
				exCtx.EXPECT().PrintMessage(core.Message{Type: core.Request, Data: "test1"}).Return(nil).Maybe()
				exCtx.EXPECT().PrintMessage(core.Message{Type: core.Request, Data: "test2"}).Return(nil).Maybe()
				return exCtx
			},
		},
		{
			name: "SubCommandReturnsAnotherCommand",
			subCommands: []core.Executer{
				NewPrintMsg(core.Message{Type: core.Request, Data: "test"}),
				NewWaitForResp(5 * time.Second),
			},
			expectedNextCmd: nil,
			expectedErr:     nil,
			mockExecutionCtx: func(t *testing.T) core.ExecutionContext {
				exCtx := core.NewMockExecutionContext(t)
				exCtx.EXPECT().PrintMessage(core.Message{Type: core.Request, Data: "test"}).Return(nil).Maybe()
				exCtx.EXPECT().PrintMessage(core.Message{Type: core.Response, Data: "response"}).Return(nil).Maybe()
				expectedResponse := core.Message{Type: core.Response, Data: "response"}
				exCtx.EXPECT().WaitForResponse(5*time.Second).Return(expectedResponse, nil).Maybe()
				return exCtx
			},
		},
		{
			name: "SubCommandFailsWithError",
			subCommands: []core.Executer{
				NewPrintMsg(core.Message{Type: core.Request, Data: "test"}),
				NewExit(),
			},
			expectedNextCmd: nil,
			expectedErr:     errors.New("mock error"),
			mockExecutionCtx: func(t *testing.T) core.ExecutionContext {
				exCtx := core.NewMockExecutionContext(t)
				exCtx.EXPECT().PrintMessage(core.Message{Type: core.Request, Data: "test"}).Return(nil).Maybe()
				return exCtx
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			exCtx := tt.mockExecutionCtx(t)
			seq := NewSequence(tt.subCommands)
			nextCmd, err := seq.Execute(exCtx)

			if tt.expectedErr != nil {
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
	tests := []struct {
		name                 string
		times                int
		subCommand           core.Executer
		expectedErr          error
		mockExecutionContext func(t *testing.T) core.ExecutionContext
	}{
		{
			name:        "SubCommandExecutesOnce",
			times:       1,
			subCommand:  NewPrintMsg(core.Message{Type: core.Request, Data: "test"}),
			expectedErr: nil,
			mockExecutionContext: func(t *testing.T) core.ExecutionContext {
				exCtx := core.NewMockExecutionContext(t)
				exCtx.EXPECT().PrintMessage(core.Message{Type: core.Request, Data: "test"}).Return(nil).Maybe()
				return exCtx
			},
		},
		{
			name:        "SubCommandExecutesMultipleTimes",
			times:       3,
			subCommand:  NewPrintMsg(core.Message{Type: core.Request, Data: "repeat"}),
			expectedErr: nil,
			mockExecutionContext: func(t *testing.T) core.ExecutionContext {
				exCtx := core.NewMockExecutionContext(t)
				exCtx.EXPECT().PrintMessage(core.Message{Type: core.Request, Data: "repeat"}).Return(nil).Times(3)
				return exCtx
			},
		},
		{
			name:        "SubCommandFails",
			times:       2,
			subCommand:  NewPrintMsg(core.Message{Type: core.Request, Data: "fail"}),
			expectedErr: errors.New("mock error"),
			mockExecutionContext: func(t *testing.T) core.ExecutionContext {
				exCtx := core.NewMockExecutionContext(t)
				exCtx.EXPECT().PrintMessage(core.Message{Type: core.Request, Data: "fail"}).Return(errors.New("mock error")).Times(1)
				return exCtx
			},
		},
		{
			name:        "ZeroExecutions",
			times:       0,
			subCommand:  NewPrintMsg(core.Message{Type: core.Request, Data: "skip"}),
			expectedErr: nil,
			mockExecutionContext: func(t *testing.T) core.ExecutionContext {
				exCtx := core.NewMockExecutionContext(t) // Nothing should be called
				return exCtx
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
		name             string
		mockContent      string
		expectedErr      error
		expectedNextCmd  core.Executer
		mockExecutionCtx func(t *testing.T) core.ExecutionContext
	}{
		{
			name:            "SuccessfulExecution",
			mockContent:     "test-content",
			expectedErr:     nil,
			expectedNextCmd: NewSend("test-response"),
			mockExecutionCtx: func(t *testing.T) core.ExecutionContext {
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

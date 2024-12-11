package command

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/ksysoev/wsget/pkg/core"
	"github.com/ksysoev/wsget/pkg/core/formater"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCommand struct {
	err error
}

func (c *mockCommand) Execute(_ core.ExecutionContext) (core.Executer, error) {
	return nil, c.err
}

func TestSequence_Execute(t *testing.T) {
	tests := []struct {
		exCtx       core.ExecutionContext
		name        string
		subCommands []core.Executer
		wantErr     bool
	}{
		{
			name:        "empty command sequence",
			subCommands: []core.Executer{},
			exCtx:       core.NewMockExecutionContext(t),
			wantErr:     false,
		},
		{
			name: "command sequence with one command",
			subCommands: []core.Executer{
				&mockCommand{},
			},
			exCtx:   core.NewMockExecutionContext(t),
			wantErr: false,
		},
		{
			name: "command sequence with multiple commands",
			subCommands: []core.Executer{
				&mockCommand{},
				&mockCommand{},
				&mockCommand{},
			},
			exCtx:   core.NewMockExecutionContext(t),
			wantErr: false,
		},
		{
			name: "command sequence with error",
			subCommands: []core.Executer{
				&mockCommand{},
				&mockCommand{err: fmt.Errorf("error")},
				&mockCommand{},
			},
			exCtx:   core.NewMockExecutionContext(t),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &Sequence{
				subCommands: tt.subCommands,
			}

			_, err := cs.Execute(tt.exCtx)

			if (err != nil) != tt.wantErr {
				t.Errorf("Sequence.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExit_Execute(t *testing.T) {
	c := &Exit{}
	_, err := c.Execute(nil)

	if err == nil {
		t.Errorf("Exit.Execute() error = %v, wantErr %v", err, true)
	}

	if !errors.Is(err, core.ErrInterrupted) {
		t.Errorf("Exit.Execute() error = %v, wantErr interrupted", err)
	}
}

func TestPrintMsg_Execute(t *testing.T) {
	tests := []struct {
		name        string
		expectedOut string
		msg         core.Message
		wantErr     bool
	}{
		{
			name: "request message",
			msg: core.Message{
				Type: core.Request,
				Data: "some request data",
			},
			wantErr:     false,
			expectedOut: "->\n" + "some request data\n",
		},
		{
			name: "response message",
			msg: core.Message{
				Type: core.Response,
				Data: "some response data",
			},
			wantErr:     false,
			expectedOut: "<-\n" + "some response data\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &PrintMsg{
				msg: tt.msg,
			}

			exCtx := core.NewMockExecutionContext(t)
			exCtx.EXPECT().Formater().Return(formater.NewFormat())
			exCtx.EXPECT().Output().Return(&bytes.Buffer{})
			exCtx.EXPECT().OutputFile().Return(nil)

			_, err := c.Execute(exCtx)

			if (err != nil) != tt.wantErr {
				t.Errorf("PrintMsg.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}
		})
	}
}

func TestCmdEdit_Execute(t *testing.T) {
	exCtx := core.NewMockExecutionContext(t)
	output := &bytes.Buffer{}
	input := make(chan core.KeyEvent)
	editor := core.NewMockEditor(t)
	editor.EXPECT().CommandMode(mock.Anything, "").Return("", nil)

	exCtx.EXPECT().Output().Return(output)
	exCtx.EXPECT().Editor().Return(editor)
	exCtx.EXPECT().Factory().Return(NewFactory(nil))

	c := &CmdEdit{}

	go func() {
		input <- core.KeyEvent{}
		close(input)
	}()

	cmd, err := c.Execute(exCtx)
	assert.NoError(t, err)
	assert.Nil(t, cmd)
}

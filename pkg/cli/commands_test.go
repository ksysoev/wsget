package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ksysoev/wsget/pkg/formater"
	"github.com/ksysoev/wsget/pkg/ws"
)

func TestCommandFactory(t *testing.T) {
	tests := []struct {
		macro   *Macro
		want    Executer
		name    string
		raw     string
		wantErr bool
	}{
		{
			name:    "empty command",
			raw:     "",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "exit command",
			raw:     "exit",
			macro:   nil,
			want:    NewCommandExit(),
			wantErr: false,
		},
		{
			name:    "edit command with content",
			raw:     "edit some content",
			macro:   nil,
			want:    NewCommandEdit("some content"),
			wantErr: false,
		},
		{
			name:    "edit command without content",
			raw:     "edit",
			macro:   nil,
			want:    NewCommandEdit(""),
			wantErr: false,
		},
		{
			name:    "send command with request",
			raw:     "send some request",
			macro:   nil,
			want:    NewCommandSend("some request"),
			wantErr: false,
		},
		{
			name:    "send command without request",
			raw:     "send",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "wait command without timeout",
			raw:     "wait",
			macro:   nil,
			want:    NewCommandWaitForResp(time.Duration(0)),
			wantErr: false,
		},
		{
			name:    "wait command with timeout",
			raw:     "wait 5",
			macro:   nil,
			want:    NewCommandWaitForResp(time.Duration(5) * time.Second),
			wantErr: false,
		},
		{
			name:    "wait command with invalid timeout",
			raw:     "wait invalid",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "unknown command",
			raw:     "unknown",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "macro command",
			raw:     "macro",
			macro:   &Macro{},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CommandFactory(tt.raw, tt.macro)

			if (err != nil) != tt.wantErr {
				t.Errorf("CommandFactory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got == nil && !tt.wantErr {
				t.Errorf("CommandFactory() got = %v, want non-nil", got)
				return
			}

			if got != nil && !strings.Contains(fmt.Sprintf("%T", got), fmt.Sprintf("%T", tt.want)) {
				t.Errorf("CommandFactory() got = %T, want %T", got, tt.want)
			}

			if got != nil && tt.want != nil {
				switch gotType := got.(type) {
				case *CommandEdit:
					if got.(*CommandEdit).content != tt.want.(*CommandEdit).content {
						t.Errorf("CommandFactory() type %v,  got = %v, want %v", gotType, got, tt.want)
					}
				case *CommandSend:
					if got.(*CommandSend).request != tt.want.(*CommandSend).request {
						t.Errorf("CommandFactory() type %v, got = %v, want %v", gotType, got, tt.want)
					}
				case *CommandWaitForResp:
					if got.(*CommandWaitForResp).timeout != tt.want.(*CommandWaitForResp).timeout {
						t.Errorf("CommandFactory() type %v, got = %v, want %v", gotType, got, tt.want)
					}
				}
			}
		})
	}
}

type mockCommand struct {
	err error
}

func (c *mockCommand) Execute(_ *ExecutionContext) (Executer, error) {
	return nil, c.err
}

func TestCommandSequence_Execute(t *testing.T) {
	tests := []struct {
		exCtx       *ExecutionContext
		name        string
		subCommands []Executer
		wantErr     bool
	}{
		{
			name:        "empty command sequence",
			subCommands: []Executer{},
			exCtx:       &ExecutionContext{},
			wantErr:     false,
		},
		{
			name: "command sequence with one command",
			subCommands: []Executer{
				&mockCommand{},
			},
			exCtx:   &ExecutionContext{},
			wantErr: false,
		},
		{
			name: "command sequence with multiple commands",
			subCommands: []Executer{
				&mockCommand{},
				&mockCommand{},
				&mockCommand{},
			},
			exCtx:   &ExecutionContext{},
			wantErr: false,
		},
		{
			name: "command sequence with error",
			subCommands: []Executer{
				&mockCommand{},
				&mockCommand{err: fmt.Errorf("error")},
				&mockCommand{},
			},
			exCtx:   &ExecutionContext{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := &CommandSequence{
				subCommands: tt.subCommands,
			}

			_, err := cs.Execute(tt.exCtx)

			if (err != nil) != tt.wantErr {
				t.Errorf("CommandSequence.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommandExit_Execute(t *testing.T) {
	c := &CommandExit{}
	_, err := c.Execute(nil)

	if err == nil {
		t.Errorf("CommandExit.Execute() error = %v, wantErr %v", err, true)
	}

	if err.Error() != "interrupted" {
		t.Errorf("CommandExit.Execute() error = %v, wantErr %v", err, "interrupted")
	}
}

func TestCommandPrintMsg_Execute(t *testing.T) {
	tests := []struct {
		exCtx       *ExecutionContext
		name        string
		expectedOut string
		msg         ws.Message
		wantErr     bool
	}{
		{
			name: "request message",
			msg: ws.Message{
				Type: ws.Request,
				Data: "some request data",
			},
			exCtx: &ExecutionContext{
				cli: &CLI{
					formater: formater.NewFormatter(),
				},
			},
			wantErr:     false,
			expectedOut: "->\n" + "some request data\n",
		},
		{
			name: "response message",
			msg: ws.Message{
				Type: ws.Response,
				Data: "some response data",
			},
			exCtx: &ExecutionContext{
				cli: &CLI{
					formater: formater.NewFormatter(),
				},
			},
			wantErr:     false,
			expectedOut: "<-\n" + "some response data\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CommandPrintMsg{
				msg: tt.msg,
			}

			var buf bytes.Buffer
			exCtx := tt.exCtx
			exCtx.cli.output = &buf

			_, err := c.Execute(exCtx)

			if (err != nil) != tt.wantErr {
				t.Errorf("CommandPrintMsg.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if buf.String() != tt.expectedOut {
				t.Errorf("CommandPrintMsg.Execute() output = %q, want %q", buf.String(), tt.expectedOut)
			}
		})
	}
}

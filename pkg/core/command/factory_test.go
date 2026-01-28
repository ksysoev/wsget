package command

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ksysoev/wsget/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestFactory_Create(t *testing.T) {
	mockMacro := NewMockMacroRepo(t)
	mockMacro.EXPECT().Get("macro", "").Return(nil, assert.AnError).Maybe()

	tests := []struct {
		macro   MacroRepo
		want    core.Executer
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
			want:    NewExit(),
			wantErr: false,
		},
		{
			name:    "edit command with content",
			raw:     "edit some content",
			macro:   nil,
			want:    NewEdit("some content"),
			wantErr: false,
		},
		{
			name:    "edit command without content",
			raw:     "edit",
			macro:   nil,
			want:    NewEdit(""),
			wantErr: false,
		},
		{
			name:    "send command with request",
			raw:     "send some request",
			macro:   nil,
			want:    NewSend("some request"),
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
			want:    NewWaitForResp(time.Duration(0)),
			wantErr: false,
		},
		{
			name:    "wait command with timeout",
			raw:     "wait 5",
			macro:   nil,
			want:    NewWaitForResp(time.Duration(5) * time.Second),
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
			macro:   mockMacro,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "editcmd command",
			raw:     "editcmd",
			macro:   nil,
			want:    NewCmdEdit(),
			wantErr: false,
		},
		{
			name:    "sleep command",
			raw:     "sleep 3",
			macro:   nil,
			want:    NewSleepCommand(3 * time.Second),
			wantErr: false,
		},
		{
			name:    "sleep command without duration",
			raw:     "sleep",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "sleep command with invalid duration",
			raw:     "sleep invalid",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "sleep command with negative duration",
			raw:     "sleep -1",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "ping command",
			raw:     "ping",
			macro:   nil,
			want:    NewPingCommand(),
			wantErr: false,
		},
		{
			name:    "repeat command",
			raw:     "repeat 3 send test",
			macro:   nil,
			want:    NewRepeatCommand(3, NewSend("test")),
			wantErr: false,
		},
		{
			name:    "repeat command without times",
			raw:     "repeat",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "repeat command with invalid times",
			raw:     "repeat invalid send test",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "repeat command with zero times",
			raw:     "repeat 0 send test",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "repeat command with negative times",
			raw:     "repeat -1 send test",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "print command with Request type",
			raw:     "print Request test message",
			macro:   nil,
			want:    NewPrintMsg(core.Message{Type: core.Request, Data: "test message"}),
			wantErr: false,
		},
		{
			name:    "print command with Response type",
			raw:     "print Response test message",
			macro:   nil,
			want:    NewPrintMsg(core.Message{Type: core.Response, Data: "test message"}),
			wantErr: false,
		},
		{
			name:    "print command without message",
			raw:     "print",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "print command with invalid type",
			raw:     "print Invalid test",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "print command with not enough arguments",
			raw:     "print Request",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "wait command with negative timeout",
			raw:     "wait -5",
			macro:   nil,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFactory(tt.macro)

			got, err := f.Create(tt.raw)

			if (err != nil) != tt.wantErr {
				t.Errorf("Factory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got == nil && !tt.wantErr {
				t.Errorf("Factory() got = %v, want non-nil", got)
				return
			}

			if got != nil && !strings.Contains(fmt.Sprintf("%T", got), fmt.Sprintf("%T", tt.want)) {
				t.Errorf("Factory() got = %T, want %T", got, tt.want)
			}

			if got != nil && tt.want != nil {
				switch gotType := got.(type) {
				case *Edit:
					edit, ok := tt.want.(*Edit)
					if !ok {
						t.Errorf("Factory() type %v, got = %v, want %v", gotType, got, tt.want)
					}

					if gotType.content != edit.content {
						t.Errorf("Factory() type %v,  got = %v, want %v", gotType, got, tt.want)
					}
				case *Send:
					send, ok := tt.want.(*Send)
					if !ok {
						t.Errorf("Factory() type %v, got = %v, want %v", gotType, got, tt.want)
					}

					if gotType.request != send.request {
						t.Errorf("Factory() type %v, got = %v, want %v", gotType, got, tt.want)
					}
				case *WaitForResp:
					wait, ok := tt.want.(*WaitForResp)
					if !ok {
						t.Errorf("Factory() type %v, got = %v, want %v", gotType, got, tt.want)
					}

					if gotType.timeout != wait.timeout {
						t.Errorf("Factory() type %v, got = %v, want %v", gotType, got, tt.want)
					}
				}
			}
		})
	}
}

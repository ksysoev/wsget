package command

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ksysoev/wsget/pkg/core"
)

func TestFactory_Create(t *testing.T) {
	tests := []struct {
		macro   *Macro
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
			macro:   &Macro{},
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
					if got.(*Edit).content != tt.want.(*Edit).content {
						t.Errorf("Factory() type %v,  got = %v, want %v", gotType, got, tt.want)
					}
				case *Send:
					if got.(*Send).request != tt.want.(*Send).request {
						t.Errorf("Factory() type %v, got = %v, want %v", gotType, got, tt.want)
					}
				case *WaitForResp:
					if got.(*WaitForResp).timeout != tt.want.(*WaitForResp).timeout {
						t.Errorf("Factory() type %v, got = %v, want %v", gotType, got, tt.want)
					}
				}
			}
		})
	}
}

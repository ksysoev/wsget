package edit

import (
	"bytes"
	"errors"
	"github.com/ksysoev/wsget/pkg/repo"
	"os"
	"testing"

	"github.com/ksysoev/wsget/pkg/core"
)

func TestNewEditor(t *testing.T) {
	output := new(bytes.Buffer)
	history := repo.NewHistory("", 0)
	editor := NewEditor(output, history, false)

	if editor.content == nil {
		t.Error("Expected non-nil content")
	}

	if editor.output != output {
		t.Error("Expected output to be set")
	}

	if editor.History != history {
		t.Error("Expected history to be set")
	}

	if editor.buffer == nil {
		t.Error("Expected non-nil buffer")
	}

	if editor.pos != 0 {
		t.Error("Expected pos to be 0")
	}
}

func TestEdit(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_history")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())

	output := new(bytes.Buffer)
	history := repo.NewHistory(tmpfile.Name(), 5)
	editor := NewEditor(output, history, false)

	keyStream := make(chan core.KeyEvent)
	defer close(keyStream)

	go func() {
		for _, key := range "request" {
			keyStream <- core.KeyEvent{Rune: key}
		}

		keyStream <- core.KeyEvent{Key: core.KeyCtrlS}
	}()

	req, err := editor.Edit(keyStream, "")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if req != "request" {
		t.Errorf("Expected empty request, got %s", req)
	}
}

func TestEditInterrupted(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_history")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())

	output := new(bytes.Buffer)
	history := repo.NewHistory(tmpfile.Name(), 5)
	editor := NewEditor(output, history, false)

	keyStream := make(chan core.KeyEvent)
	defer close(keyStream)

	go func() {
		keyStream <- core.KeyEvent{Key: core.KeyCtrlC}
	}()

	req, err := editor.Edit(keyStream, "")

	if err == nil {
		t.Error("Expected error")
	}

	if req != "" {
		t.Errorf("Expected empty request, got %s", req)
	}

	go func() {
		keyStream <- core.KeyEvent{Key: core.KeyCtrlD}
	}()

	req, err = editor.Edit(keyStream, "")

	if !errors.Is(err, core.ErrInterrupted) {
		t.Error("Expected error")
	}

	if req != "" {
		t.Errorf("Expected empty request, got %s", req)
	}
}

func TestEditExitEditor(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_history")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())

	output := new(bytes.Buffer)
	history := repo.NewHistory(tmpfile.Name(), 5)
	editor := NewEditor(output, history, false)

	keyStream := make(chan core.KeyEvent)
	defer close(keyStream)

	go func() {
		keyStream <- core.KeyEvent{Key: core.KeyEsc}
	}()

	req, err := editor.Edit(keyStream, "")

	if err != nil {
		t.Error("Expected no error")
	}

	if req != "" {
		t.Errorf("Expected empty request, got %s", req)
	}
}

func TestEditClosingKeyboard(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_history")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())

	output := new(bytes.Buffer)
	history := repo.NewHistory(tmpfile.Name(), 5)
	editor := NewEditor(output, history, false)

	keyStream := make(chan core.KeyEvent)
	close(keyStream)

	req, err := editor.Edit(keyStream, "")

	if err == nil {
		t.Error("Expected error")
	}

	if req != "" {
		t.Errorf("Expected empty request, got %s", req)
	}
}

func TestEditSpecialKeys(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_history")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())

	output := new(bytes.Buffer)
	history := repo.NewHistory(tmpfile.Name(), 5)
	editor := NewEditor(output, history, false)

	keyStream := make(chan core.KeyEvent)

	go func() {
		for _, key := range []core.Key{
			core.KeySpace,
			core.KeyCtrlU,
			core.KeyEsc,
		} {
			keyStream <- core.KeyEvent{Key: key}
		}
	}()

	req, err := editor.Edit(keyStream, "")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if req != "" {
		t.Errorf("Expected empty request, got %s", req)
	}

	outputStr := output.String()

	if outputStr != " "+LineClear+"\r" {
		t.Errorf("Unexpected output: %q", outputStr)
	}

	go func() {
		for _, key := range []core.Key{
			core.KeySpace,
			core.KeyCtrlU,
			core.KeyEsc,
		} {
			keyStream <- core.KeyEvent{Key: key}
		}
	}()
}
func TestHandleEscKey(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		keyEvent core.KeyEvent
		handled  bool
	}{
		{
			name:     "Alt + Left",
			keyEvent: core.KeyEvent{Rune: 98},
			expected: "",
			handled:  true,
		},
		{
			name:     "Alt + Right",
			keyEvent: core.KeyEvent{Rune: 102},
			expected: "",
			handled:  true,
		},
		{
			name:     "Alt + Delete",
			keyEvent: core.KeyEvent{Rune: 100},
			expected: "",
			handled:  true,
		},
		{
			name:     "Esc",
			keyEvent: core.KeyEvent{Rune: 0},
			expected: "",
			handled:  false,
		},
		{
			name:     "Esc + any other key",
			keyEvent: core.KeyEvent{Rune: 1},
			expected: "",
			handled:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := new(bytes.Buffer)
			editor := NewEditor(output, nil, false)

			handled := handleEscKey(tt.keyEvent, editor)

			if handled != tt.handled {
				t.Errorf("Expected handled to be %v, got %v", tt.handled, handled)
			}

			outputStr := output.String()
			if tt.expected != "" && !bytes.Contains(output.Bytes(), []byte(tt.expected)) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, outputStr)
			}
		})
	}
}

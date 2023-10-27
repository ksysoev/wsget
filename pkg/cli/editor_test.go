package cli

import (
	"os"
	"testing"

	"github.com/eiannone/keyboard"
)

func TestNewEditor(t *testing.T) {
	output := os.Stdout
	history := NewHistory("", 0)
	editor := NewEditor(output, history)

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

func TestEditRequest(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_history")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())

	output := os.Stdout
	history := NewHistory(tmpfile.Name(), 5)
	editor := NewEditor(output, history)

	keyStream := make(chan keyboard.KeyEvent)
	defer close(keyStream)

	go func() {
		for _, key := range "request" {
			keyStream <- keyboard.KeyEvent{Rune: key}
		}

		keyStream <- keyboard.KeyEvent{Key: keyboard.KeyCtrlS}
	}()

	req, err := editor.EditRequest(keyStream, "")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if req != "request" {
		t.Errorf("Expected empty request, got %s", req)
	}
}

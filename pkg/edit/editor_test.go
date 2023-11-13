package edit

import (
	"bytes"
	"os"
	"testing"

	"github.com/eiannone/keyboard"
)

func TestNewEditor(t *testing.T) {
	output := new(bytes.Buffer)
	history := NewHistory("", 0)
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
	history := NewHistory(tmpfile.Name(), 5)
	editor := NewEditor(output, history, false)

	keyStream := make(chan keyboard.KeyEvent)
	defer close(keyStream)

	go func() {
		for _, key := range "request" {
			keyStream <- keyboard.KeyEvent{Rune: key}
		}

		keyStream <- keyboard.KeyEvent{Key: keyboard.KeyCtrlS}
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
	history := NewHistory(tmpfile.Name(), 5)
	editor := NewEditor(output, history, false)

	keyStream := make(chan keyboard.KeyEvent)
	defer close(keyStream)

	go func() {
		keyStream <- keyboard.KeyEvent{Key: keyboard.KeyCtrlC}
	}()

	req, err := editor.Edit(keyStream, "")

	if err == nil {
		t.Error("Expected error")
	}

	if req != "" {
		t.Errorf("Expected empty request, got %s", req)
	}

	go func() {
		keyStream <- keyboard.KeyEvent{Key: keyboard.KeyCtrlD}
	}()

	req, err = editor.Edit(keyStream, "")

	if err == nil {
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
	history := NewHistory(tmpfile.Name(), 5)
	editor := NewEditor(output, history, false)

	keyStream := make(chan keyboard.KeyEvent)
	defer close(keyStream)

	go func() {
		keyStream <- keyboard.KeyEvent{Key: keyboard.KeyEsc}
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
	history := NewHistory(tmpfile.Name(), 5)
	editor := NewEditor(output, history, false)

	keyStream := make(chan keyboard.KeyEvent)
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
	history := NewHistory(tmpfile.Name(), 5)
	editor := NewEditor(output, history, false)

	keyStream := make(chan keyboard.KeyEvent)

	go func() {
		for _, key := range []keyboard.Key{
			keyboard.KeySpace,
			keyboard.KeyCtrlU,
			keyboard.KeyEsc,
		} {
			keyStream <- keyboard.KeyEvent{Key: key}
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
		for _, key := range []keyboard.Key{
			keyboard.KeySpace,
			keyboard.KeyCtrlU,
			keyboard.KeyEsc,
		} {
			keyStream <- keyboard.KeyEvent{Key: key}
		}
	}()
}

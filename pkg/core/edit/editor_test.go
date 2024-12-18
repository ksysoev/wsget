package edit

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/ksysoev/wsget/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewEditor(t *testing.T) {
	output := new(bytes.Buffer)
	history := NewMockHistoryRepo(t)
	editor := NewEditor(output, history, false)

	if editor.content == nil {
		t.Error("Expected non-nil content")
	}

	if editor.output != output {
		t.Error("Expected output to be set")
	}

	if editor.history != history {
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
	output := new(bytes.Buffer)

	history := NewMockHistoryRepo(t)
	history.EXPECT().ResetPosition()
	history.EXPECT().AddRequest("request")

	editor := NewEditor(output, history, false)

	keyStream := make(chan core.KeyEvent)
	defer close(keyStream)

	editor.SetInput(keyStream)

	go func() {
		for _, key := range "request" {
			keyStream <- core.KeyEvent{Rune: key}
		}

		keyStream <- core.KeyEvent{Key: core.KeyCtrlS}
	}()

	req, err := editor.Edit(context.Background(), "")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if req != "request" {
		t.Errorf("Expected empty request, got %s", req)
	}
}

func TestEditInterrupted(t *testing.T) {
	history := NewMockHistoryRepo(t)
	history.EXPECT().ResetPosition()

	editor := NewEditor(io.Discard, history, false)

	keyStream := make(chan core.KeyEvent)
	defer close(keyStream)

	editor.SetInput(keyStream)

	go func() {
		keyStream <- core.KeyEvent{Key: core.KeyCtrlC}
	}()

	req, err := editor.Edit(context.Background(), "")

	if err == nil {
		t.Error("Expected error")
	}

	if req != "" {
		t.Errorf("Expected empty request, got %s", req)
	}

	go func() {
		keyStream <- core.KeyEvent{Key: core.KeyCtrlD}
	}()

	req, err = editor.Edit(context.Background(), "")

	if !errors.Is(err, core.ErrInterrupted) {
		t.Error("Expected error")
	}

	if req != "" {
		t.Errorf("Expected empty request, got %s", req)
	}
}

func TestEdit_CancelledContext(t *testing.T) {
	history := NewMockHistoryRepo(t)
	history.EXPECT().ResetPosition()

	editor := NewEditor(io.Discard, history, false)

	keyStream := make(chan core.KeyEvent)
	defer close(keyStream)

	editor.SetInput(keyStream)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req, err := editor.Edit(ctx, "")

	if !errors.Is(err, core.ErrInterrupted) {
		t.Error("Expected error")
	}

	if req != "" {
		t.Errorf("Expected empty request, got %s", req)
	}
}

func TestEdit_NoInput(t *testing.T) {
	history := NewMockHistoryRepo(t)
	history.EXPECT().ResetPosition()

	editor := NewEditor(io.Discard, history, false)

	req, err := editor.Edit(context.Background(), "")

	assert.EqualError(t, err, "input stream is not set")
	assert.Empty(t, req)
}

func TestEditExitEditor(t *testing.T) {
	history := NewMockHistoryRepo(t)
	history.EXPECT().ResetPosition()

	editor := NewEditor(io.Discard, history, false)

	keyStream := make(chan core.KeyEvent)
	defer close(keyStream)

	editor.SetInput(keyStream)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req, err := editor.Edit(ctx, "")

	assert.ErrorIs(t, err, core.ErrInterrupted)
	assert.Empty(t, req)
}

func TestEditClosingKeyboard(t *testing.T) {
	output := new(bytes.Buffer)

	history := NewMockHistoryRepo(t)
	history.EXPECT().ResetPosition()

	editor := NewEditor(output, history, false)

	keyStream := make(chan core.KeyEvent)
	close(keyStream)

	editor.SetInput(keyStream)

	req, err := editor.Edit(context.Background(), "")

	if err == nil {
		t.Error("Expected error")
	}

	if req != "" {
		t.Errorf("Expected empty request, got %s", req)
	}
}

func TestEditSpecialKeys(t *testing.T) {
	output := new(bytes.Buffer)

	history := NewMockHistoryRepo(t)
	history.EXPECT().ResetPosition()

	editor := NewEditor(output, history, false)

	keyStream := make(chan core.KeyEvent)
	defer close(keyStream)

	editor.SetInput(keyStream)

	go func() {
		for _, key := range []core.Key{
			core.KeySpace,
			core.KeyCtrlU,
			core.KeyEsc,
		} {
			keyStream <- core.KeyEvent{Key: key}
		}
	}()

	req, err := editor.Edit(context.Background(), "")

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

			assert.Equal(t, tt.handled, handled, "Expected handled to be %v, got %v", tt.handled, handled)

			outputStr := output.String()
			if tt.expected != "" {
				assert.Contains(t, outputStr, tt.expected, "Expected output to contain %q", tt.expected)
			}
		})
	}
}

func TestNewLineOrDone(t *testing.T) {
	tests := []struct {
		name           string
		expectedOutput string
		prevSymbol     rune
		isSingleLine   bool
		isPasting      bool
		expectedIsDone bool
	}{
		{
			name:           "Single line editor",
			isSingleLine:   true,
			prevSymbol:     'a',
			isPasting:      false,
			expectedIsDone: true,
			expectedOutput: "",
		},
		{
			name:           "Multi-line, previous is backslash",
			isSingleLine:   false,
			prevSymbol:     '\\',
			isPasting:      false,
			expectedIsDone: false,
			expectedOutput: "\b \b\n",
		},
		{
			name:           "Multi-line, previous is backslash, pasting",
			isSingleLine:   false,
			prevSymbol:     '\\',
			isPasting:      true,
			expectedIsDone: false,
			expectedOutput: "\n",
		},
		{
			name:           "Multi-line, no backslash, not pasting",
			isSingleLine:   false,
			prevSymbol:     'a',
			isPasting:      false,
			expectedIsDone: true,
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := new(bytes.Buffer)
			content := NewContent()
			content.InsertSymbol(tt.prevSymbol)

			editor := &Editor{
				output:       output,
				content:      content,
				isSingleLine: tt.isSingleLine,
			}

			isDone := editor.newLineOrDone(tt.isPasting)

			assert.Equal(t, tt.expectedIsDone, isDone)
			assert.Equal(t, tt.expectedOutput, output.String())
		})
	}
}

func TestEditorHandleKey(t *testing.T) {
	tests := []struct {
		expectedErr    error
		name           string
		expectedRes    string
		expectedOutput string
		keyEvent       core.KeyEvent
		expectedNext   bool
	}{
		{
			name:         "Empty Event",
			keyEvent:     core.KeyEvent{},
			expectedNext: true,
			expectedRes:  "",
			expectedErr:  nil,
		},
		{
			name:         "Ctrl+S for Done",
			keyEvent:     core.KeyEvent{Key: core.KeyCtrlS},
			expectedNext: false,
			expectedRes:  "",
			expectedErr:  nil,
		},
		{
			name:         "Ctrl+C Interrupt",
			keyEvent:     core.KeyEvent{Key: core.KeyCtrlC},
			expectedNext: false,
			expectedRes:  "",
			expectedErr:  core.ErrInterrupted,
		},
		{
			name:         "Ctrl+D Interrupt",
			keyEvent:     core.KeyEvent{Key: core.KeyCtrlD},
			expectedNext: false,
			expectedRes:  "",
			expectedErr:  core.ErrInterrupted,
		},
		{
			name:         "Alt+Backspace",
			keyEvent:     core.KeyEvent{Key: core.KeyAltBackspace},
			expectedNext: true,
			expectedRes:  "",
			expectedErr:  nil,
		},
		{
			name:           "Space Key",
			keyEvent:       core.KeyEvent{Key: core.KeySpace},
			expectedNext:   true,
			expectedRes:    "",
			expectedErr:    nil,
			expectedOutput: " ",
		},
		{
			name:         "Backspace Key",
			keyEvent:     core.KeyEvent{Key: core.KeyBackspace},
			expectedNext: true,
			expectedRes:  "",
			expectedErr:  nil,
		},
		{
			name:         "Delete Key",
			keyEvent:     core.KeyEvent{Key: core.KeyDelete},
			expectedNext: true,
			expectedRes:  "",
			expectedErr:  nil,
		},
		{
			name:         "Arrow Left Key",
			keyEvent:     core.KeyEvent{Key: core.KeyArrowLeft},
			expectedNext: true,
			expectedRes:  "",
			expectedErr:  nil,
		},
		{
			name:         "Arrow Right Key",
			keyEvent:     core.KeyEvent{Key: core.KeyArrowRight},
			expectedNext: true,
			expectedRes:  "",
			expectedErr:  nil,
		},
		{
			name:         "Tab with No Dictionary",
			keyEvent:     core.KeyEvent{Key: core.KeyTab},
			expectedNext: true,
			expectedRes:  "",
			expectedErr:  nil,
		},
		{
			name:         "Home Key",
			keyEvent:     core.KeyEvent{Key: core.KeyHome},
			expectedNext: true,
			expectedRes:  "",
			expectedErr:  nil,
		},
		{
			name:         "End Key",
			keyEvent:     core.KeyEvent{Key: core.KeyEnd},
			expectedNext: true,
			expectedRes:  "",
			expectedErr:  nil,
		},
		{
			name:         "Arrow Up Key",
			keyEvent:     core.KeyEvent{Key: core.KeyArrowUp},
			expectedNext: true,
			expectedRes:  "",
			expectedErr:  nil,
		},
		{
			name:         "Arrow Down Key",
			keyEvent:     core.KeyEvent{Key: core.KeyArrowDown},
			expectedNext: true,
			expectedRes:  "",
			expectedErr:  nil,
		},
		{
			name:         "Esc Key",
			keyEvent:     core.KeyEvent{Key: core.KeyEsc},
			expectedNext: false,
			expectedRes:  "",
			expectedErr:  nil,
		},
		{
			name:         "Alt + Left Key",
			keyEvent:     core.KeyEvent{Key: core.KeyEsc, Rune: 98},
			expectedNext: true,
			expectedRes:  "",
			expectedErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := new(bytes.Buffer)
			editor := NewEditor(output, nil, false)

			mockHistory := NewMockHistoryRepo(t)
			mockHistory.EXPECT().PrevRequest().Return("").Maybe()
			mockHistory.EXPECT().NextRequest().Return("").Maybe()
			mockHistory.EXPECT().Search(mock.Anything).Return("").Maybe()

			editor.history = mockHistory

			next, res, err := editor.handleKey(tt.keyEvent)

			assert.Equal(t, tt.expectedNext, next, "Expected next to be %v, got %v", tt.expectedNext, next)
			assert.Equal(t, tt.expectedRes, res, "Expected res to be %q, got %q", tt.expectedRes, res)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr, "Expected error %v, got %v", tt.expectedErr, err)
			} else {
				assert.NoError(t, err, "Expected no error, got %v", err)
			}

			if tt.expectedOutput != "" {
				assert.Contains(t, output.String(), tt.expectedOutput, "Expected output to contain %q", tt.expectedOutput)
			}
		})
	}
}

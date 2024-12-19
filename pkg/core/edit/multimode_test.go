package edit

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/ksysoev/wsget/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestNewMultiMode(t *testing.T) {
	output := io.Discard
	reqHistory := NewMockHistoryRepo(t)
	cmdHistory := NewMockHistoryRepo(t)

	multiMode := NewMultiMode(output, reqHistory, cmdHistory)
	assert.NotNil(t, multiMode)
	assert.NotNil(t, multiMode.commandMode)
	assert.NotNil(t, multiMode.editMode)
}

func TestMultiMode_CommandMode(t *testing.T) {
	history := NewMockHistoryRepo(t)
	history.EXPECT().ResetPosition()
	history.EXPECT().AddRequest("initial")

	multiMode := &MultiMode{
		editMode:    NewEditor(io.Discard, history, true),
		commandMode: NewEditor(io.Discard, history, true),
	}
	keyStream := make(chan core.KeyEvent, 1)

	defer close(keyStream)

	keyStream <- core.KeyEvent{Key: core.KeyEnter}

	multiMode.SetInput(keyStream)

	result, err := multiMode.CommandMode(context.Background(), "initial")
	assert.NoError(t, err)
	assert.Equal(t, "initial", result)
}

func TestMultiMode_Edit(t *testing.T) {
	history := NewMockHistoryRepo(t)
	history.EXPECT().ResetPosition()
	history.EXPECT().AddRequest("edit")

	multiMode := &MultiMode{
		commandMode: NewEditor(io.Discard, history, true),
		editMode:    NewEditor(io.Discard, history, true),
	}

	keyStream := make(chan core.KeyEvent, 1)

	defer close(keyStream)

	keyStream <- core.KeyEvent{Key: core.KeyEnter}

	multiMode.SetInput(keyStream)

	result, err := multiMode.Edit(context.Background(), "edit")
	assert.NoError(t, err)
	assert.Equal(t, "edit", result)
}

type failingWriter struct{}

func (f failingWriter) Write(_ []byte) (n int, err error) {
	return 0, errors.New("failed to write")
}

func TestEditorOpenHook(t *testing.T) {
	tests := []struct {
		writer         io.Writer
		expectedError  error
		name           string
		expectedOutput string
	}{
		{
			name:           "Success with valid writer",
			writer:         &strings.Builder{},
			expectedOutput: "->\n\x1b[?25h", // Output contains the "->" and the ANSI escape ShowCursor
			expectedError:  nil,
		},
		{
			name:           "Error on colored write",
			writer:         failingWriter{},
			expectedOutput: "",
			expectedError:  errors.New("failed to write"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a writer for capturing output
			builder, ok := tt.writer.(*strings.Builder)

			// Execute the function
			err := editorOpenHook(tt.writer)

			// Assert expected outcomes
			assert.Equal(t, tt.expectedError, err)

			// Check writer contents for successful cases
			if ok {
				assert.Equal(t, tt.expectedOutput, builder.String())
			}
		})
	}
}

func TestEditorCloseHook(t *testing.T) {
	tests := []struct {
		writer         io.Writer
		expectedError  error
		name           string
		expectedOutput string
	}{
		{
			name:           "Success with valid writer",
			writer:         &strings.Builder{},
			expectedOutput: LineUp + LineClear + HideCursor, // Output contains the ANSI escape sequences for cursor movement and hiding
			expectedError:  nil,
		},
		{
			name:           "Error with failing writer",
			writer:         failingWriter{},
			expectedOutput: "",
			expectedError:  errors.New("failed to write"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a writer for capturing output
			builder, ok := tt.writer.(*strings.Builder)

			// Execute the function
			err := editorCloseHook(tt.writer)

			// Assert expected outcomes
			assert.Equal(t, tt.expectedError, err)

			// Check writer contents for successful cases
			if ok {
				assert.Equal(t, tt.expectedOutput, builder.String())
			}
		})
	}
}

func TestCmdEditorOpenHook(t *testing.T) {
	tests := []struct {
		writer         io.Writer
		expectedError  error
		name           string
		expectedOutput string
	}{
		{
			name:           "Success with valid writer",
			writer:         &strings.Builder{},
			expectedOutput: ":" + ShowCursor, // ':' followed by ShowCursor
			expectedError:  nil,
		},
		{
			name:           "Error with failing writer",
			writer:         failingWriter{},
			expectedOutput: "",
			expectedError:  errors.New("failed to write"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a writer for capturing output
			builder, ok := tt.writer.(*strings.Builder)

			// Execute the function
			err := cmdEditorOpenHook(tt.writer)

			// Assert expected outcomes
			assert.Equal(t, tt.expectedError, err)

			// Check writer contents for successful cases
			if ok {
				assert.Equal(t, tt.expectedOutput, builder.String())
			}
		})
	}
}

func TestCmdEditorCloseHook(t *testing.T) {
	tests := []struct {
		writer         io.Writer
		expectedError  error
		name           string
		expectedOutput string
	}{
		{
			name:           "Success with valid writer",
			writer:         &strings.Builder{},
			expectedOutput: LineClear + "\r" + HideCursor, // Verify correct output with valid writer
			expectedError:  nil,
		},
		{
			name:           "Error with failing writer",
			writer:         failingWriter{},
			expectedOutput: "",
			expectedError:  errors.New("failed to write"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a writer for capturing output
			builder, ok := tt.writer.(*strings.Builder)

			// Execute the function
			err := cmdEditorCloseHook(tt.writer)

			// Assert expected outcomes
			assert.Equal(t, tt.expectedError, err)

			// Check writer contents for successful cases
			if ok {
				assert.Equal(t, tt.expectedOutput, builder.String())
			}
		})
	}
}

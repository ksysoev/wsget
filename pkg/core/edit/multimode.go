package edit

import (
	"context"
	"fmt"
	"io"

	"github.com/ksysoev/wsget/pkg/core"
)

const (
	HideCursor = "\x1b[?25l"
	ShowCursor = "\x1b[?25h"
)

type MultiMode struct {
	commandMode *Editor
	editMode    *Editor
}

// NewMultiMode initializes a new MultiMode structure with separate editors for command and standard input modes.
// It takes an io.Writer, two HistoryRepo instances for request and command histories, and an optional Dictionary.
// It returns a pointer to the created MultiMode, setting up command and edit modes appropriately.
func NewMultiMode(output io.Writer, reqHistory, cmdHistory HistoryRepo) *MultiMode {
	commandMode := NewEditor(
		output,
		cmdHistory,
		true,
		WithOpenHook(func(w io.Writer) error {
			_, err := fmt.Fprintf(w, ":"+ShowCursor)
			return err
		}),
		WithCloseHook(func(w io.Writer) error {
			_, err := fmt.Fprintf(w, LineClear+"\r"+HideCursor)
			return err
		}),
	)

	return &MultiMode{
		commandMode: commandMode,
		editMode:    NewEditor(output, reqHistory, false),
	}
}

// CommandMode activates the command mode, reading user input from keyStream with an initial buffer initBuffer.
// It returns the resulting command string or an error if any issue occurs.
func (m *MultiMode) CommandMode(ctx context.Context, initBuffer string) (string, error) {
	return m.commandMode.Edit(ctx, initBuffer)
}

// Edit switches the editor to edit mode, processing user input from keyStream with an initial buffer.
// It returns the final string after editing or an error if an issue occurs.
func (m *MultiMode) Edit(ctx context.Context, initBuffer string) (string, error) {
	return m.editMode.Edit(ctx, initBuffer)
}

// SetInput sets the input channel for both command and edit modes.
func (m *MultiMode) SetInput(input <-chan core.KeyEvent) {
	m.commandMode.SetInput(input)
	m.editMode.SetInput(input)
}

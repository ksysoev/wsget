package edit

import (
	"context"
	"fmt"
	"io"

	"github.com/fatih/color"
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
		WithOpenHook(cmdEditorOpenHook),
		WithCloseHook(cmdEditorCloseHook),
	)

	editMode := NewEditor(
		output,
		reqHistory,
		false,
		WithOpenHook(editorOpenHook),
		WithCloseHook(editorCloseHook),
	)

	return &MultiMode{
		commandMode: commandMode,
		editMode:    editMode,
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

// editorOpenHook prepares the editor's environment when it opens.
// It takes w of type io.Writer to write initialization sequences.
// It returns an error if writing to the provided io.Writer fails.
func editorOpenHook(w io.Writer) error {
	if _, err := color.New(color.FgGreen).Fprint(w, "->"); err != nil {
		return err
	}

	_, err := fmt.Fprint(w, "\n"+ShowCursor)

	return err
}

// editorCloseHook restores the editor's environment when it closes.
// It takes w of type io.Writer to write cleanup sequences.
// It returns an error if writing to the provided io.Writer fails.
func editorCloseHook(w io.Writer) error {
	_, err := fmt.Fprint(w, LineUp+LineClear+HideCursor)

	return err
}

// cmdEditorOpenHook prepares the command editor's environment when it opens.
// It takes w of type io.Writer to write initialization sequences.
// It returns an error if writing to the provided io.Writer fails.
func cmdEditorOpenHook(w io.Writer) error {
	_, err := fmt.Fprint(w, ":"+ShowCursor)
	return err
}

// cmdEditorCloseHook cleans up the command editor's environment when it closes.
// It takes w of type io.Writer to write terminal reset sequences.
// It returns an error if writing to the provided io.Writer fails.
func cmdEditorCloseHook(w io.Writer) error {
	_, err := fmt.Fprint(w, LineClear+"\r"+HideCursor)
	return err
}

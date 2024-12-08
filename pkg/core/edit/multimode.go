package edit

import (
	"io"

	"github.com/ksysoev/wsget/pkg/core"
)

type MultiMode struct {
	commandMode *Editor
	editMode    *Editor
}

// NewMultiMode initializes a new MultiMode structure with separate editors for command and standard input modes.
// It takes an io.Writer, two HistoryRepo instances for request and command histories, and an optional Dictionary.
// It returns a pointer to the created MultiMode, setting up command and edit modes appropriately.
func NewMultiMode(output io.Writer, reqHistory, cmdHistory HistoryRepo, cmdDict *Dictionary) *MultiMode {
	commandMode := NewEditor(output, cmdHistory, true)
	if cmdDict != nil {
		commandMode.Dictionary = cmdDict
	}

	return &MultiMode{
		commandMode: commandMode,
		editMode:    NewEditor(output, reqHistory, false),
	}
}

// CommandMode activates the command mode, reading user input from keyStream with an initial buffer initBuffer.
// It returns the resulting command string or an error if any issue occurs.
func (m *MultiMode) CommandMode(keyStream <-chan core.KeyEvent, initBuffer string) (string, error) {
	return m.commandMode.Edit(keyStream, initBuffer)
}

// Edit switches the editor to edit mode, processing user input from keyStream with an initial buffer.
// It returns the final string after editing or an error if an issue occurs.
func (m *MultiMode) Edit(keyStream <-chan core.KeyEvent, initBuffer string) (string, error) {
	return m.editMode.Edit(keyStream, initBuffer)
}

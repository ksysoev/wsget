package edit

import (
	"io"

	"github.com/ksysoev/wsget/pkg/core"
)

type MultiMode struct {
	commandMode *Editor
	editMode    *Editor
}

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

func (m *MultiMode) CommandMode(keyStream <-chan core.KeyEvent, initBuffer string) (string, error) {
	return m.commandMode.Edit(keyStream, initBuffer)
}

func (m *MultiMode) Edit(keyStream <-chan core.KeyEvent, initBuffer string) (string, error) {
	return m.editMode.Edit(keyStream, initBuffer)
}

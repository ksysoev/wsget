package edit

import (
	"io"
	"testing"

	"github.com/ksysoev/wsget/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestNewMultiMode(t *testing.T) {
	output := io.Discard
	reqHistory := NewMockHistoryRepo(t)
	cmdHistory := NewMockHistoryRepo(t)
	cmdDict := NewDictionary([]string{})

	multiMode := NewMultiMode(output, reqHistory, cmdHistory, cmdDict)
	assert.NotNil(t, multiMode)
	assert.NotNil(t, multiMode.commandMode)
	assert.NotNil(t, multiMode.editMode)
}

func TestMultiMode_CommandMode(t *testing.T) {
	history := NewMockHistoryRepo(t)
	history.EXPECT().ResetPosition()
	history.EXPECT().AddRequest("initial")

	multiMode := &MultiMode{
		commandMode: NewEditor(io.Discard, history, true),
	}
	keyStream := make(chan core.KeyEvent, 1)

	defer close(keyStream)

	keyStream <- core.KeyEvent{Key: core.KeyEnter}

	result, err := multiMode.CommandMode(keyStream, "initial")
	assert.NoError(t, err)
	assert.Equal(t, "initial", result)
}

func TestMultiMode_Edit(t *testing.T) {
	history := NewMockHistoryRepo(t)
	history.EXPECT().ResetPosition()
	history.EXPECT().AddRequest("edit")

	multiMode := &MultiMode{
		editMode: NewEditor(io.Discard, history, true),
	}

	keyStream := make(chan core.KeyEvent, 1)

	defer close(keyStream)

	keyStream <- core.KeyEvent{Key: core.KeyEnter}

	result, err := multiMode.Edit(keyStream, "edit")
	assert.NoError(t, err)
	assert.Equal(t, "edit", result)
}

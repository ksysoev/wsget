package edit

import (
	"context"
	"io"
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

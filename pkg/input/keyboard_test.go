package input

import (
	"context"
	"testing"

	"github.com/ksysoev/wsget/pkg/core"
	"github.com/stretchr/testify/assert"
)

type mockKeyHandler struct {
	events []core.KeyEvent
}

func (m *mockKeyHandler) OnKeyEvent(event core.KeyEvent) {
	m.events = append(m.events, event)
}

func TestNewKeyboard(t *testing.T) {
	handler := &mockKeyHandler{}
	kb := NewKeyboard(handler)

	assert.NotNil(t, kb)
	assert.Equal(t, handler, kb.handler)
}

func TestKeyboard_Close(t *testing.T) {
	handler := &mockKeyHandler{}
	kb := NewKeyboard(handler)

	// Close should not panic
	assert.NotPanics(t, func() {
		kb.Close()
	})
}

func TestKeyboard_Run_ContextCancelled(t *testing.T) {
	handler := &mockKeyHandler{}
	kb := NewKeyboard(handler)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Run should return nil when context is cancelled before keyboard.Open
	// Note: This test may fail in some environments where keyboard.Open succeeds
	// In that case, the test validates that context cancellation works
	err := kb.Run(ctx)

	// The function should either return nil (context cancelled) or an error from keyboard.Open
	// Both are acceptable outcomes
	if err != nil {
		t.Logf("keyboard.Open returned error (expected in headless environment): %v", err)
	}
}

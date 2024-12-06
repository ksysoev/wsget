package input

import (
	"context"

	"github.com/eiannone/keyboard"
	"github.com/ksysoev/wsget/pkg/core"
)

const (
	KeyboardBufferSize = 10
)

type KeyHandler interface {
	OnKeyEvent(event core.KeyEvent)
}

type Keyboard struct {
	handler KeyHandler
}

func NewKeyboard(handler KeyHandler) *Keyboard {
	return &Keyboard{
		handler: handler,
	}
}

func (k *Keyboard) Run(ctx context.Context) error {
	if err := keyboard.Open(); err != nil {
		return err
	}

	keysEvents, err := keyboard.GetKeys(KeyboardBufferSize)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-keysEvents:
			if e.Err != nil {
				return e.Err
			}

			event := core.KeyEvent{Key: core.Key(e.Key), Rune: e.Rune}
			k.handler.OnKeyEvent(event)
		}
	}
}

func (k *Keyboard) Close() {
	keyboard.Close()
}

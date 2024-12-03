package core

import "github.com/eiannone/keyboard"

const (
	KeyboardBufferSize = 10
)

type Keyboard struct{}

func NewKeyboard() *Keyboard {
	return &Keyboard{}
}

func (k *Keyboard) GetKeys() (<-chan keyboard.KeyEvent, error) {
	if err := keyboard.Open(); err != nil {
		return nil, err
	}

	keysEvents, err := keyboard.GetKeys(KeyboardBufferSize)
	if err != nil {
		return nil, err
	}

	return keysEvents, nil
}

func (k *Keyboard) Close() {
	keyboard.Close()
}

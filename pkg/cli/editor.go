package cli

import (
	"fmt"

	"github.com/eiannone/keyboard"
)

type Editor struct {
	History *History
	content *Content
	buffer  []rune
	pos     int
}

func NewEditor(history *History) *Editor {
	return &Editor{
		History: history,
		content: NewContent(),
		buffer:  make([]rune, 0),
		pos:     0,
	}
}

func (ed *Editor) EditRequest(keyStream <-chan keyboard.KeyEvent, initBuffer string) (string, error) {
	ed.History.ResetPosition()
	fmt.Print(ed.content.ReplaceText(initBuffer))

	for e := range keyStream {
		if e.Err != nil {
			return "", e.Err
		}

		switch e.Key {
		case keyboard.KeyCtrlC, keyboard.KeyCtrlD:
			return "", fmt.Errorf("interrupted")
		case keyboard.KeyCtrlS:
			req := ed.content.ToRequest()

			fmt.Print(ed.content.Clear())

			if req == "" {
				return req, fmt.Errorf("cannot send empty request")
			}

			ed.History.AddRequest(req)

			return req, nil
		case keyboard.KeyEsc:
			return "", nil
		case keyboard.KeyCtrlU:
			fmt.Print(ed.content.Clear())
		case keyboard.KeySpace:
			fmt.Print(ed.content.InsertSymbol(' '))
		case keyboard.KeyEnter:
			fmt.Print(ed.content.InsertSymbol('\n'))
		case keyboard.KeyBackspace, keyboard.KeyDelete, MacOSDeleteKey:
			fmt.Print(ed.content.RemoveSymbol())
		case keyboard.KeyArrowLeft:
			fmt.Print(ed.content.MovePositionLeft())
		case keyboard.KeyArrowRight:
			fmt.Print(ed.content.MovePositionRight())
		case keyboard.KeyArrowUp:
			req := ed.History.PrevRequst()

			if req == "" {
				fmt.Print(Bell)
				continue
			}

			fmt.Print(ed.content.ReplaceText(req))
		case keyboard.KeyArrowDown:
			req := ed.History.NextRequst()

			if req == "" {
				fmt.Print(Bell)
				continue
			}

			fmt.Print(ed.content.ReplaceText(req))
		default:
			if e.Key > 0 {
				continue
			}

			fmt.Print(ed.content.InsertSymbol(e.Rune))
		}
	}

	return "", fmt.Errorf("keyboard stream was unexpectably closed")
}

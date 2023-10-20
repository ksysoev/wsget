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
	historyIndex := 0

	fmt.Print(ed.content.ReplaceText(initBuffer))

	for e := range keyStream {
		if e.Err != nil {
			return "", e.Err
		}

		switch e.Key {
		case keyboard.KeyCtrlC, keyboard.KeyCtrlD:
			return "", fmt.Errorf("interrupted")
		case keyboard.KeyCtrlS:
			fmt.Print(ed.content.MoveToEnd() + "\n")

			req := ed.content.ToRequest()
			if req == "" {
				return req, fmt.Errorf("cannot send empty request")
			}

			ed.History.AddRequest(req)

			return req, nil
		case keyboard.KeyEsc:
			return "", nil

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
			historyIndex++
			req := ed.History.GetRequst(historyIndex)

			if req == "" {
				historyIndex--
				continue
			}

			fmt.Print(ed.content.ReplaceText(req))
		case keyboard.KeyArrowDown:
			historyIndex--
			req := ed.History.GetRequst(historyIndex)

			if req == "" {
				historyIndex++
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

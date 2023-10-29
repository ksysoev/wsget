package cli

import (
	"fmt"
	"io"

	"github.com/eiannone/keyboard"
)

type Editor struct {
	History *History
	content *Content
	output  io.Writer
	buffer  []rune
	pos     int
}

func NewEditor(output io.Writer, history *History) *Editor {
	return &Editor{
		History: history,
		content: NewContent(),
		buffer:  make([]rune, 0),
		pos:     0,
		output:  output,
	}
}

func (ed *Editor) EditRequest(keyStream <-chan keyboard.KeyEvent, initBuffer string) (string, error) {
	ed.History.ResetPosition()
	fmt.Fprint(ed.output, ed.content.ReplaceText(initBuffer))

	for e := range keyStream {
		if e.Err != nil {
			return "", e.Err
		}

		switch e.Key {
		case keyboard.KeyCtrlV:
			fmt.Println("Ctrl+V")
		case keyboard.KeyCtrlC, keyboard.KeyCtrlD:
			return "", fmt.Errorf("interrupted")
		case keyboard.KeyCtrlS:
			return ed.done()
		case keyboard.KeyEsc:
			fmt.Fprint(ed.output, ed.content.Clear())
			return "", nil
		case keyboard.KeyCtrlU:
			fmt.Fprint(ed.output, ed.content.Clear())
		case keyboard.KeySpace:
			fmt.Fprint(ed.output, ed.content.InsertSymbol(' '))
		case keyboard.KeyEnter:
			if isDone := ed.newLineOrDone(); isDone {
				return ed.done()
			}
		case keyboard.KeyBackspace, keyboard.KeyDelete, MacOSDeleteKey:
			fmt.Fprint(ed.output, ed.content.RemoveSymbol())
		case keyboard.KeyArrowLeft:
			fmt.Fprint(ed.output, ed.content.MovePositionLeft())
		case keyboard.KeyArrowRight:
			fmt.Fprint(ed.output, ed.content.MovePositionRight())
		case keyboard.KeyArrowUp:
			ed.prevFromHistory()
		case keyboard.KeyArrowDown:
			ed.nextFromHistory()
		default:
			if e.Key > 0 {
				fmt.Fprintf(ed.output, ">%q<", e.Key)
				continue
			}

			fmt.Fprint(ed.output, ed.content.InsertSymbol(e.Rune))
		}
	}

	return "", fmt.Errorf("keyboard stream was unexpectably closed")
}

func (ed *Editor) done() (string, error) {
	req := ed.content.ToRequest()

	fmt.Fprint(ed.output, ed.content.Clear())

	if req == "" {
		return req, fmt.Errorf("empty request")
	}

	ed.History.AddRequest(req)

	return req, nil
}

func (ed *Editor) prevFromHistory() {
	req := ed.History.PrevRequst()

	if req == "" {
		fmt.Fprint(ed.output, Bell)
		return
	}

	fmt.Fprint(ed.output, ed.content.ReplaceText(req))
}

func (ed *Editor) nextFromHistory() {
	req := ed.History.NextRequst()

	if req == "" {
		fmt.Fprint(ed.output, Bell)
		return
	}

	fmt.Fprint(ed.output, ed.content.ReplaceText(req))
}

func (ed *Editor) newLineOrDone() (isDone bool) {
	prev := ed.content.PrevSymbol()

	isDone = prev != '\\'
	if !isDone {
		fmt.Fprint(ed.output, ed.content.RemoveSymbol())
		fmt.Fprint(ed.output, ed.content.InsertSymbol('\n'))

		return isDone
	}

	return isDone
}

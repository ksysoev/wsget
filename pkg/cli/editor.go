package cli

import (
	"fmt"
	"strings"

	"github.com/eiannone/keyboard"
)

type Editor struct {
	History *History
	buffer  []rune
	pos     int
}

func NewEditor(history *History) *Editor {
	return &Editor{
		History: history,
		buffer:  make([]rune, 0),
		pos:     0,
	}
}

func (ed *Editor) EditRequest(keyStream <-chan keyboard.KeyEvent, initBuffer string) (string, error) {
	historyIndex := 0
	ed.buffer = []rune(initBuffer)

	for e := range keyStream {
		if e.Err != nil {
			return "", e.Err
		}

		switch e.Key {
		case keyboard.KeyCtrlC, keyboard.KeyCtrlD:
			return "", fmt.Errorf("interrupted")
		case keyboard.KeyCtrlS:
			stringBuffer := string(ed.buffer)
			requet := strings.TrimSpace(stringBuffer)

			if requet == "" {
				return requet, fmt.Errorf("cannot send empty request")
			}

			ed.History.AddRequest(requet)

			return requet, nil
		case keyboard.KeyEsc:
			return "", nil

		case keyboard.KeySpace:
			fmt.Print(" ")

			ed.buffer = append(ed.buffer, ' ')
			ed.pos++
		case keyboard.KeyEnter:
			fmt.Print("\n")

			ed.buffer = append(ed.buffer, '\n')
			ed.pos++
		case keyboard.KeyBackspace, keyboard.KeyDelete, MacOSDeleteKey:
			ed.removeSymbol()
		case keyboard.KeyArrowUp:
			historyIndex++
			req := ed.History.GetRequst(historyIndex)

			if req == "" {
				historyIndex--
				continue
			}

			ed.clearInput()

			fmt.Print(req)
			ed.buffer = []rune(req)
			ed.pos = len(ed.buffer) - 1
		case keyboard.KeyArrowDown:
			historyIndex--
			req := ed.History.GetRequst(historyIndex)

			if req == "" {
				historyIndex++
				continue
			}

			ed.clearInput()

			fmt.Print(req)
			ed.buffer = []rune(req)
			ed.pos = len(ed.buffer) - 1
		default:
			if e.Key > 0 {
				continue
			}

			fmt.Print(string(e.Rune))

			ed.buffer = append(ed.buffer, e.Rune)
			ed.pos++
		}
	}

	return "", fmt.Errorf("keyboard stream was unexpectably closed")
}

func (ed *Editor) clearInput() {
	for i := 0; i < len(ed.buffer); i++ {
		if ed.buffer[i] == '\n' {
			fmt.Print(LineUp)
			fmt.Print(LineClear)
		} else {
			fmt.Print("\b \b")
		}
	}
}

func (ed *Editor) removeSymbol() {
	if ed.pos < 1 || ed.pos > len(ed.buffer) {
		return
	}

	ed.pos--
	symbol := ed.buffer[ed.pos]
	buffer := ed.buffer[:ed.pos]
	if ed.pos < (len(ed.buffer) - 1) {
		buffer = append(buffer, ed.buffer[ed.pos+1:]...)
	}
	ed.buffer = buffer

	if symbol == '\n' {

		fmt.Print(LineUp)

		startPrevLine := LastIndexOf(ed.buffer, '\n')
		if startPrevLine == -1 {
			startPrevLine = 0
		} else {
			startPrevLine++
		}

		fmt.Print(string(ed.buffer[startPrevLine:]))
	} else {
		fmt.Print("\b \b")
	}
}

func LastIndexOf(buffer []rune, search rune) int {
	for i := len(buffer) - 1; i >= 0; i-- {
		if buffer[i] == search {
			return i
		}
	}

	return -1
}

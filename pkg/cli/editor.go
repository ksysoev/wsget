package cli

import (
	"fmt"
	"strings"

	"github.com/eiannone/keyboard"
)

type Editor struct {
	History *History
	buffer  []rune
}

func NewEditor(history *History) *Editor {
	return &Editor{
		History: history,
		buffer:  make([]rune, 0),
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
		case keyboard.KeyEnter:
			fmt.Print("\n")

			ed.buffer = append(ed.buffer, '\n')
		case keyboard.KeyBackspace, keyboard.KeyDelete, MacOSDeleteKey:
			if len(ed.buffer) == 0 {
				continue
			}

			if ed.buffer[len(ed.buffer)-1] == '\n' {
				ed.buffer = ed.buffer[:len(ed.buffer)-1]

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
				ed.buffer = ed.buffer[:len(ed.buffer)-1]
			}
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
		default:
			if e.Key > 0 {
				continue
			}

			fmt.Print(string(e.Rune))

			ed.buffer = append(ed.buffer, e.Rune)
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

func LastIndexOf(buffer []rune, search rune) int {
	for i := len(buffer) - 1; i >= 0; i-- {
		if buffer[i] == search {
			return i
		}
	}

	return -1
}

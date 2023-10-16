package cli

import (
	"fmt"
	"strings"

	"github.com/eiannone/keyboard"
)

type Editor struct {
	History *History
}

func NewEditor(history *History) *Editor {
	return &Editor{
		History: history,
	}
}

func (ed *Editor) EditRequest(keyStream <-chan keyboard.KeyEvent, buffer string) (string, error) {
	historyIndex := 0

	for e := range keyStream {
		if e.Err != nil {
			return buffer, e.Err
		}

		switch e.Key {
		case keyboard.KeyCtrlC, keyboard.KeyCtrlD:
			return buffer, fmt.Errorf("interrupted")
		case keyboard.KeyCtrlS:
			if buffer == "" {
				return buffer, fmt.Errorf("cannot send empty request")
			}

			requet := strings.TrimSpace(buffer)

			ed.History.AddRequest(requet)

			return requet, nil
		case keyboard.KeyEsc:
			return "", nil

		case keyboard.KeySpace:
			fmt.Print(" ")

			buffer += " "
		case keyboard.KeyEnter:
			fmt.Print("\n")

			buffer += "\n"
		case keyboard.KeyBackspace, keyboard.KeyDelete, MacOSDeleteKey:
			if buffer == "" {
				continue
			}

			if buffer[len(buffer)-1] == '\n' {
				buffer = buffer[:len(buffer)-1]

				fmt.Print(LineUp)

				startPrevLine := strings.LastIndex(buffer, "\n")
				if startPrevLine == -1 {
					startPrevLine = 0
				} else {
					startPrevLine++
				}

				fmt.Print(buffer[startPrevLine:])
			} else {
				fmt.Print("\b \b")
				buffer = buffer[:len(buffer)-1]
			}
		case keyboard.KeyArrowUp:
			historyIndex++
			req := ed.History.GetRequst(historyIndex)

			if req == "" {
				historyIndex--
				continue
			}

			ed.clearInput(buffer)

			fmt.Print(req)
			buffer = req
		case keyboard.KeyArrowDown:
			historyIndex--
			req := ed.History.GetRequst(historyIndex)

			if req == "" {
				historyIndex++
				continue
			}

			ed.clearInput(buffer)

			fmt.Print(req)
			buffer = req
		default:
			if e.Key > 0 {
				continue
			}

			fmt.Print(string(e.Rune))

			buffer += string(e.Rune)
		}
	}

	return buffer, fmt.Errorf("keyboard stream was unexpectably closed")
}

func (ed *Editor) clearInput(buffer string) {
	for i := 0; i < len(buffer); i++ {
		if buffer[i] == '\n' {
			fmt.Print(LineUp)
			fmt.Print(LineClear)
		} else {
			fmt.Print("\b \b")
		}
	}
}

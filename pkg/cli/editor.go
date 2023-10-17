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

func (ed *Editor) EditRequest(keyStream <-chan keyboard.KeyEvent, initBuffer string) (string, error) {
	historyIndex := 0
	runeBuffer := []rune(initBuffer)

	for e := range keyStream {
		if e.Err != nil {
			return "", e.Err
		}

		switch e.Key {
		case keyboard.KeyCtrlC, keyboard.KeyCtrlD:
			return "", fmt.Errorf("interrupted")
		case keyboard.KeyCtrlS:
			stringBuffer := string(runeBuffer)
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

			runeBuffer = append(runeBuffer, ' ')
		case keyboard.KeyEnter:
			fmt.Print("\n")

			runeBuffer = append(runeBuffer, '\n')
		case keyboard.KeyBackspace, keyboard.KeyDelete, MacOSDeleteKey:
			if len(runeBuffer) == 0 {
				continue
			}

			if runeBuffer[len(runeBuffer)-1] == '\n' {
				runeBuffer = runeBuffer[:len(runeBuffer)-1]

				fmt.Print(LineUp)

				startPrevLine := LastIndexOf(runeBuffer, '\n')
				if startPrevLine == -1 {
					startPrevLine = 0
				} else {
					startPrevLine++
				}

				fmt.Print(string(runeBuffer[startPrevLine:]))
			} else {
				fmt.Print("\b \b")
				runeBuffer = runeBuffer[:len(runeBuffer)-1]
			}
		case keyboard.KeyArrowUp:
			historyIndex++
			req := ed.History.GetRequst(historyIndex)

			if req == "" {
				historyIndex--
				continue
			}

			ed.clearInput(runeBuffer)

			fmt.Print(req)
			runeBuffer = []rune(req)
		case keyboard.KeyArrowDown:
			historyIndex--
			req := ed.History.GetRequst(historyIndex)

			if req == "" {
				historyIndex++
				continue
			}

			ed.clearInput(runeBuffer)

			fmt.Print(req)
			runeBuffer = []rune(req)
		default:
			if e.Key > 0 {
				continue
			}

			fmt.Print(string(e.Rune))

			runeBuffer = append(runeBuffer, e.Rune)
		}
	}

	return "", fmt.Errorf("keyboard stream was unexpectably closed")
}

func (ed *Editor) clearInput(runeBuffer []rune) {
	for i := 0; i < len(runeBuffer); i++ {
		if runeBuffer[i] == '\n' {
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

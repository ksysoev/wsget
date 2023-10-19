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
	ed.pos = len(ed.buffer)

	for e := range keyStream {
		if e.Err != nil {
			return "", e.Err
		}

		switch e.Key {
		case keyboard.KeyCtrlC, keyboard.KeyCtrlD:
			return "", fmt.Errorf("interrupted")
		case keyboard.KeyCtrlS:
			if ed.pos < len(ed.buffer) {
				fmt.Print(string(ed.buffer[ed.pos:]))
			}
			fmt.Print("\n")
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
			ed.InsertSymbol(' ')
		case keyboard.KeyEnter:
			ed.InsertSymbol('\n')
		case keyboard.KeyBackspace, keyboard.KeyDelete, MacOSDeleteKey:
			ed.removeSymbol()
		case keyboard.KeyArrowLeft:
			if ed.pos > 0 {
				ed.pos--
				if ed.buffer[ed.pos] == '\n' {
					fmt.Print(LineUp)

					startPrevLine := LastIndexOf(ed.buffer, ed.pos-1, '\n')
					if startPrevLine == -1 {
						startPrevLine = 0
					} else {
						startPrevLine++
					}

					fmt.Print(string(ed.buffer[startPrevLine:ed.pos]))
				} else {
					fmt.Print("\b")
				}
			}
		case keyboard.KeyArrowRight:
			if ed.pos < len(ed.buffer) {
				fmt.Print(string(ed.buffer[ed.pos]))
				ed.pos++
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
			ed.pos = len(ed.buffer)
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
			ed.pos = len(ed.buffer)
		default:
			if e.Key > 0 {
				continue
			}

			ed.InsertSymbol(e.Rune)
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

		startPrevLine := LastIndexOf(ed.buffer, ed.pos, '\n')
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

func (ed *Editor) InsertSymbol(symbol rune) {
	buffer := make([]rune, ed.pos, len(ed.buffer)+1)
	copy(buffer, ed.buffer[:ed.pos])
	buffer = append(buffer, symbol)
	endOfStr := ""

	if ed.pos < len(ed.buffer) {
		buffer = append(buffer, ed.buffer[ed.pos:]...)
		moveCursor := ""

		for i := ed.pos; i < len(ed.buffer); i++ {
			if ed.buffer[i] != '\n' {
				endOfStr += string(ed.buffer[i])
				moveCursor += "\b"
			}
		}

		endOfStr += moveCursor
	}

	ed.buffer = buffer
	ed.pos++
	fmt.Print(string(symbol) + endOfStr)
}

func LastIndexOf(buffer []rune, pos int, search rune) int {
	for i := pos; i >= 0; i-- {
		if buffer[i] == search {
			return i
		}
	}

	return -1
}

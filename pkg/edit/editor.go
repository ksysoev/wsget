package edit

import (
	"fmt"
	"io"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/ksysoev/wsget/pkg/clierrors"
)

const (
	PastingTimingThresholdInMicrosec = 250
	MacOSDeleteKey                   = 127
	Bell                             = "\a"
)

type Editor struct {
	History         *History
	content         *Content
	Dictionary      *Dictionary
	output          io.Writer
	prevPressedTime time.Time
	buffer          []rune
	pos             int
	isSingleLine    bool
}

// NewEditor creates a new instance of Editor struct.
// It takes an io.Writer to output the editor content, a *History to store the command history,
// a boolean value to indicate whether the editor should be single line or not.
// It returns a pointer to the created Editor struct.
func NewEditor(output io.Writer, history *History, isSingleLine bool) *Editor {
	return &Editor{
		History:         history,
		content:         NewContent(),
		buffer:          make([]rune, 0),
		pos:             0,
		output:          output,
		prevPressedTime: time.Now(),
		isSingleLine:    isSingleLine,
	}
}

// Edit reads input from the user via a keyboard stream and returns the resulting string.
// It takes a channel of keyboard events and an initial buffer string as input.
// It returns the resulting string and an error if any.
func (ed *Editor) Edit(keyStream <-chan keyboard.KeyEvent, initBuffer string) (string, error) {
	ed.History.ResetPosition()
	fmt.Fprint(ed.output, ed.content.ReplaceText(initBuffer))

	for e := range keyStream {
		isPasting := ed.isPasting()

		if e.Err != nil {
			return "", e.Err
		}

		if keyboard.KeyEsc == e.Key && e.Rune == 98 {
			// Alt + Left
			fmt.Fprint(ed.output, ed.content.MoveToPrevWord())
			continue
		}

		if keyboard.KeyEsc == e.Key && e.Rune == 102 {
			// Alt + Right
			fmt.Fprint(ed.output, ed.content.MoveToNextWord())
			continue
		}

		if keyboard.KeyCtrlW == e.Key {
			// Alt + Backspace
			fmt.Fprint(ed.output, ed.content.DeleteToPrevWord())
			continue
		}

		if keyboard.KeyEsc == e.Key && e.Rune == 100 {
			// Alt + Delete
			fmt.Fprint(ed.output, ed.content.DeleteToNextWord())
			continue
		}

		switch e.Key {
		case keyboard.KeyCtrlC, keyboard.KeyCtrlD:
			return "", clierrors.Interrupted{}
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
			if ed.newLineOrDone(isPasting) {
				return ed.done()
			}
		case keyboard.KeyBackspace, MacOSDeleteKey:
			fmt.Fprint(ed.output, ed.content.RemovePrevSymbol())
		case keyboard.KeyDelete:
			fmt.Fprint(ed.output, ed.content.RemoveNextSymbol())
		case keyboard.KeyArrowLeft:
			fmt.Fprint(ed.output, ed.content.MovePositionLeft())
		case keyboard.KeyArrowRight:
			fmt.Fprint(ed.output, ed.content.MovePositionRight())
		case keyboard.KeyArrowUp:
			ed.prevFromHistory()
		case keyboard.KeyArrowDown:
			ed.nextFromHistory()
		case keyboard.KeyTab:
			content := ed.content.String()
			if ed.Dictionary == nil || content == "" {
				continue
			}

			match := ed.Dictionary.Search(content)
			if match == "" || match == content {
				continue
			}

			diff := match[len(content):]

			for _, r := range diff {
				fmt.Fprint(ed.output, ed.content.InsertSymbol(r))
			}
		case keyboard.KeyHome:
			fmt.Fprint(ed.output, ed.content.MoveToRowStart())
		case keyboard.KeyEnd:
			fmt.Fprint(ed.output, ed.content.MoveToRowEnd())
		default:
			if e.Key > 0 {
				continue
			}

			if ed.isSingleLine && e.Rune == '\n' {
				continue
			}

			fmt.Fprint(ed.output, ed.content.InsertSymbol(e.Rune))
		}
	}

	return "", fmt.Errorf("keyboard stream was unexpectably closed")
}

// done returns the current request and clears the editor content.
// If the editor content is empty, it returns an empty string.
// It also adds the request to the editor's history.
func (ed *Editor) done() (string, error) {
	req := ed.content.ToRequest()

	fmt.Fprint(ed.output, ed.content.Clear())

	if req == "" {
		return req, nil
	}

	ed.History.AddRequest(req)

	return req, nil
}

// prevFromHistory retrieves the previous request from the history and replaces the current content with it.
// If there is no previous request, it prints a bell character and returns.
func (ed *Editor) prevFromHistory() {
	req := ed.History.PrevRequst()

	if req == "" {
		fmt.Fprint(ed.output, Bell)
		return
	}

	fmt.Fprint(ed.output, ed.content.ReplaceText(req))
}

// nextFromHistory retrieves the next request from the history and replaces the current content with it.
// If there are no more requests in the history, it prints a bell character and returns.
func (ed *Editor) nextFromHistory() {
	req := ed.History.NextRequst()

	if req == "" {
		fmt.Fprint(ed.output, Bell)
		return
	}

	fmt.Fprint(ed.output, ed.content.ReplaceText(req))
}

// newLineOrDone returns a boolean indicating whether the editor is done or not.
// It takes a boolean isPasting as an argument which indicates whether the editor is currently pasting or not.
// If isSingleLine is true, it returns true. If the previous symbol is not a backslash, it returns true.
// If isPasting is true, it returns false. Otherwise, it returns true.
func (ed *Editor) newLineOrDone(isPasting bool) (isDone bool) {
	prev := ed.content.PrevSymbol()

	if ed.isSingleLine {
		return true
	}

	isDone = prev != '\\'
	if !isDone {
		fmt.Fprint(ed.output, ed.content.RemovePrevSymbol())
		fmt.Fprint(ed.output, ed.content.InsertSymbol('\n'))

		return isDone
	}

	if isPasting {
		fmt.Fprint(ed.output, ed.content.InsertSymbol('\n'))
		return false
	}

	return isDone
}

// isPasting returns true if the time elapsed since the last key press is less than the pasting timing threshold.
func (ed *Editor) isPasting() bool {
	elapsed := time.Since(ed.prevPressedTime)
	ed.prevPressedTime = time.Now()

	return elapsed.Microseconds() < PastingTimingThresholdInMicrosec
}

// Close saves the history to the history file.
func (ed *Editor) Close() error {
	return ed.History.SaveToFile()
}

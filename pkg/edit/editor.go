package edit

import (
	"fmt"
	"io"
	"time"

	"github.com/ksysoev/wsget/pkg/core"
)

type HistoryRepo interface {
	AddRequest(req string)
	PrevRequest() string
	NextRequest() string
	ResetPosition()
}

const (
	PastingTimingThresholdInMicrosec = 250
	MacOSDeleteKey                   = 127
	Bell                             = "\a"
)

type Editor struct {
	History         HistoryRepo
	content         *Content
	Dictionary      *Dictionary
	output          io.Writer
	prevPressedTime time.Time
	buffer          []rune
	pos             int
	isSingleLine    bool
}

// NewEditor creates a new instance of Editor struct.
// It takes an io.Writer to output the editor content, a *HistoryRepo to store the command history,
// a boolean value to indicate whether the editor should be single line or not.
// It returns a pointer to the created Editor struct.
func NewEditor(output io.Writer, history HistoryRepo, isSingleLine bool) *Editor {
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
func (ed *Editor) Edit(keyStream <-chan core.KeyEvent, initBuffer string) (string, error) {
	ed.History.ResetPosition()
	fmt.Fprint(ed.output, ed.content.ReplaceText(initBuffer))

	for e := range keyStream {
		isPasting := ed.isPasting()

		switch e.Key {
		case core.KeyAltBackspace:
			fmt.Fprint(ed.output, ed.content.DeleteToPrevWord())
			continue
		case core.KeyCtrlC, core.KeyCtrlD:
			return "", core.ErrInterrupted
		case core.KeyCtrlS:
			return ed.done()
		case core.KeyEsc:
			if handleEscKey(e, ed) {
				continue
			}

			return "", nil
		case core.KeyCtrlU:
			fmt.Fprint(ed.output, ed.content.Clear())
		case core.KeySpace:
			fmt.Fprint(ed.output, ed.content.InsertSymbol(' '))
		case core.KeyEnter:
			if ed.newLineOrDone(isPasting) {
				return ed.done()
			}
		case core.KeyBackspace, MacOSDeleteKey:
			fmt.Fprint(ed.output, ed.content.RemovePrevSymbol())
		case core.KeyDelete:
			fmt.Fprint(ed.output, ed.content.RemoveNextSymbol())
		case core.KeyArrowLeft:
			fmt.Fprint(ed.output, ed.content.MovePositionLeft())
		case core.KeyArrowRight:
			fmt.Fprint(ed.output, ed.content.MovePositionRight())
		case core.KeyArrowUp:
			ed.prevFromHistory()
		case core.KeyArrowDown:
			ed.nextFromHistory()
		case core.KeyTab:
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
		case core.KeyHome:
			fmt.Fprint(ed.output, ed.content.MoveToRowStart())
		case core.KeyEnd:
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

// handleEscKey processes keyboard events involving the Escape key and updates the editor state accordingly.
// It takes a keyboard.KeyEvent `e` and an `Editor` pointer `ed`.
// It returns a boolean indicating whether the event was handled (`true`) or not (`false`).
//
// The function handles the following key combinations:
// - Alt + Left (`e.Rune == 98`): Moves the cursor to the previous word.
// - Alt + Right (`e.Rune == 102`): Moves the cursor to the next word.
// - Alt + Delete (`e.Rune == 100`): Deletes text up to the next word.
// - Esc (`e.Rune == 0`): Clears the editor content and stops further processing.
//
// Any other key combination with the Escape key is ignored, and the function returns `true`.
func handleEscKey(e core.KeyEvent, ed *Editor) bool {
	switch e.Rune {
	case 98: //nolint:mnd // Alt + Left
		fmt.Fprint(ed.output, ed.content.MoveToPrevWord())
		return true
	case 102: //nolint:mnd // Alt + Right
		fmt.Fprint(ed.output, ed.content.MoveToNextWord())
		return true
	case 100: //nolint:mnd // Alt + Delete
		fmt.Fprint(ed.output, ed.content.DeleteToNextWord())
		return true
	case 0:
		// Esc
		fmt.Fprint(ed.output, ed.content.Clear())
		return false
	default:
		// Esc + any other key is ignored
		return true
	}
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
	req := ed.History.PrevRequest()

	if req == "" {
		fmt.Fprint(ed.output, Bell)
		return
	}

	fmt.Fprint(ed.output, ed.content.ReplaceText(req))
}

// nextFromHistory retrieves the next request from the history and replaces the current content with it.
// If there are no more requests in the history, it prints a bell character and returns.
func (ed *Editor) nextFromHistory() {
	req := ed.History.NextRequest()

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

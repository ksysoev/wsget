package edit

import (
	"context"
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
	Search(prefix string) string
}

const (
	PastingTimingThresholdInMicrosec = 250
	MacOSDeleteKey                   = 127
	Bell                             = "\a"
)

type Option func(*Editor)

type Editor struct {
	prevPressedTime time.Time
	history         HistoryRepo
	output          io.Writer
	input           <-chan core.KeyEvent
	content         *Content
	onOpen          func(io.Writer) error
	onClose         func(io.Writer) error
	buffer          []rune
	pos             int
	isSingleLine    bool
}

// NewEditor initializes a new instance of Editor for text editing tasks.
// It takes output of type io.Writer for writing, history of type HistoryRepo for request history,
// and isSingleLine of type bool to specify single-line mode.
// It returns a pointer to an initialized Editor structure.
func NewEditor(output io.Writer, history HistoryRepo, isSingleLine bool, opts ...Option) *Editor {
	e := &Editor{
		history:         history,
		content:         NewContent(),
		buffer:          make([]rune, 0),
		pos:             0,
		output:          output,
		prevPressedTime: time.Now(),
		isSingleLine:    isSingleLine,
		onOpen:          func(_ io.Writer) error { return nil },
		onClose:         func(_ io.Writer) error { return nil },
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// SetInput sets the input channel for the Editor instance to process keyboard events.
// It takes a single parameter input of type <-chan core.KeyEvent.
// This method does not return any value.
func (ed *Editor) SetInput(input <-chan core.KeyEvent) {
	ed.input = input
}

// Edit processes keyboard input to manipulate and return the edited content.
// It takes a context ctx of type context.Context for cancellation and an initial buffer initBuffer of type string.
// It returns the final edited string content or an error if input is unavailable, keyboard stream is closed, or an interrupt occurs.
func (ed *Editor) Edit(ctx context.Context, initBuffer string) (res string, err error) {
	if err := ed.onOpen(ed.output); err != nil {
		return "", fmt.Errorf("failed to execute open hook: %w", err)
	}

	defer func() {
		if closeErr := ed.onClose(ed.output); err == nil {
			err = closeErr
		}
	}()

	ed.history.ResetPosition()

	if _, err := fmt.Fprint(ed.output, ed.content.ReplaceText(initBuffer)); err != nil {
		return "", fmt.Errorf("failed to write initial buffer: %w", err)
	}

	if ed.input == nil {
		return "", fmt.Errorf("input stream is not set")
	}

	for {
		select {
		case <-ctx.Done():
			return "", core.ErrInterrupted
		case e, ok := <-ed.input:
			if !ok {
				return "", fmt.Errorf("keyboard stream was unexpectedly closed")
			}

			next, s, err := ed.handleKey(e)

			switch {
			case err != nil:
				return "", err
			case next:
				continue
			default:
				return s, nil
			}
		}
	}
}

// handleKey processes a single keyboard event to modify the editor's content or control its behavior.
// It takes e of type core.KeyEvent, representing the pressed key and associated rune.
// It returns a boolean next indicating whether to continue processing, a string res for the result, and an error if any.
// It returns an error if the operation is interrupted or invalid input occurs.
func (ed *Editor) handleKey(e core.KeyEvent) (next bool, res string, err error) {
	isPasting := ed.isPasting()

	switch e.Key {
	case core.KeyAltBackspace:
		_, _ = fmt.Fprint(ed.output, ed.content.DeleteToPrevWord())

		return true, "", nil
	case core.KeyCtrlC, core.KeyCtrlD:
		return false, "", core.ErrInterrupted
	case core.KeyCtrlS:
		return false, ed.done(), nil
	case core.KeyEsc:
		if handleEscKey(e, ed) {
			return true, "", nil
		}

		return false, "", nil
	case core.KeyCtrlU:
		_, _ = fmt.Fprint(ed.output, ed.content.Clear())
	case core.KeySpace:
		_, _ = fmt.Fprint(ed.output, ed.content.InsertSymbol(' '))
	case core.KeyEnter:
		if ed.newLineOrDone(isPasting) {
			return false, ed.done(), nil
		}
	case core.KeyBackspace, MacOSDeleteKey:
		_, _ = fmt.Fprint(ed.output, ed.content.RemovePrevSymbol())
	case core.KeyDelete:
		_, _ = fmt.Fprint(ed.output, ed.content.RemoveNextSymbol())
	case core.KeyArrowLeft:
		_, _ = fmt.Fprint(ed.output, ed.content.MovePositionLeft())
	case core.KeyArrowRight:
		_, _ = fmt.Fprint(ed.output, ed.content.MovePositionRight())
	case core.KeyArrowUp:
		ed.prevFromHistory()
	case core.KeyArrowDown:
		ed.nextFromHistory()
	case core.KeyTab:
		curWord := ed.content.GetCurrentWord()

		match := ed.history.Search(curWord)
		if match == "" || match == curWord {
			return true, "", nil
		}

		diff := match[len(curWord):]

		for _, r := range diff {
			_, _ = fmt.Fprint(ed.output, ed.content.InsertSymbol(r))
		}
	case core.KeyHome:
		_, _ = fmt.Fprint(ed.output, ed.content.MoveToRowStart())
	case core.KeyEnd:
		_, _ = fmt.Fprint(ed.output, ed.content.MoveToRowEnd())
	default:
		if e.Key > 0 {
			return true, "", nil
		}

		if ed.isSingleLine && e.Rune == '\n' {
			return true, "", nil
		}

		_, _ = fmt.Fprint(ed.output, ed.content.InsertSymbol(e.Rune))
	}

	return true, "", nil
}

// handleEscKey handles specific key events triggered by pressing the Esc key or its combinations.
// It takes e of type core.KeyEvent, representing the key and associated rune, and ed of type *Editor for handling content operations.
// It returns a boolean indicating whether further handling of the key event is required.
// The function executes specific actions for combinations like Alt + Left, Alt + Right, and Alt + Delete.
// It returns false when Esc is pressed alone, clearing the editor content. Esc with other keys is ignored.
func handleEscKey(e core.KeyEvent, ed *Editor) bool {
	switch e.Rune {
	case 98: //nolint:mnd // Alt + Left
		_, _ = fmt.Fprint(ed.output, ed.content.MoveToPrevWord())
		return true
	case 102: //nolint:mnd // Alt + Right
		_, _ = fmt.Fprint(ed.output, ed.content.MoveToNextWord())
		return true
	case 100: //nolint:mnd // Alt + Delete
		_, _ = fmt.Fprint(ed.output, ed.content.DeleteToNextWord())
		return true
	case 0:
		// Esc
		_, _ = fmt.Fprint(ed.output, ed.content.Clear())
		return false
	default:
		// Esc + any other key is ignored
		return true
	}
}

// done finalizes the editing process and clears the editor's content.
// It converts the current content to a request, clears it from the editor, and adds it to the history if non-empty.
// It returns the final request string. If the content is empty, it returns an empty string.
func (ed *Editor) done() string {
	req := ed.content.ToRequest()

	_, _ = fmt.Fprint(ed.output, ed.content.Clear())

	if req == "" {
		return req
	}

	ed.history.AddRequest(req)

	return req
}

// prevFromHistory retrieves the previous request from the history and replaces the current content with it.
// It does not take any parameters.
// It prints a bell character if no previous history exists.
func (ed *Editor) prevFromHistory() {
	req := ed.history.PrevRequest()

	if req == "" {
		_, _ = fmt.Fprint(ed.output, Bell)
		return
	}

	_, _ = fmt.Fprint(ed.output, ed.content.ReplaceText(req))
}

// nextFromHistory retrieves the next request from the history and replaces the current content with it.
// It does not take any parameters.
// It prints a bell character if no next history exists.
func (ed *Editor) nextFromHistory() {
	req := ed.history.NextRequest()

	if req == "" {
		_, _ = fmt.Fprint(ed.output, Bell)
		return
	}

	_, _ = fmt.Fprint(ed.output, ed.content.ReplaceText(req))
}

// newLineOrDone inserts a newline or marks the editing process as done based on input and editor state.
// It takes isPasting of type bool, indicating whether the input is a pasted sequence.
// It returns a boolean isDone, which is true if the editing process is complete.
// It returns true immediately if the editor is in single-line mode. When not pasting, it checks the last symbol to determine the result.
func (ed *Editor) newLineOrDone(isPasting bool) (isDone bool) {
	prev := ed.content.PrevSymbol()

	if ed.isSingleLine {
		return true
	}

	if isPasting {
		_, _ = fmt.Fprint(ed.output, ed.content.InsertSymbol('\n'))
		return false
	}

	isDone = prev != '\\'
	if !isDone {
		_, _ = fmt.Fprint(ed.output, ed.content.RemovePrevSymbol())
		_, _ = fmt.Fprint(ed.output, ed.content.InsertSymbol('\n'))

		return isDone
	}

	return isDone
}

// isPasting determines whether the user is pasting content based on the time interval between key presses.
// It does not take any parameters.
// It returns a boolean: true if the key press interval is less than the defined threshold, otherwise false.
func (ed *Editor) isPasting() bool {
	elapsed := time.Since(ed.prevPressedTime)
	ed.prevPressedTime = time.Now()

	return elapsed.Microseconds() < PastingTimingThresholdInMicrosec
}

// WithOpenHook sets the onOpen function for the Editor instance.
// It takes a function hook of type func(io.Writer) error.
// It returns an Option function to set the onOpen function.
func WithOpenHook(hook func(io.Writer) error) Option {
	return func(ed *Editor) {
		ed.onOpen = hook
	}
}

// WithCloseHook sets the onClose function for the Editor instance.
// It takes a function hook of type func(io.Writer) error.
// It returns an Option function to set the onClose function.
func WithCloseHook(hook func(io.Writer) error) Option {
	return func(ed *Editor) {
		ed.onClose = hook
	}
}

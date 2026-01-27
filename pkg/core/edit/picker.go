package edit

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/ksysoev/wsget/pkg/core"
	"github.com/ksysoev/wsget/pkg/repo/history"
)

const (
	maxDisplayLines = 10
	pickerPrompt    = "fuzzy> "
)

// FuzzyPicker provides an interactive fuzzy search interface for history.
type FuzzyPicker struct {
	output  io.Writer
	input   <-chan core.KeyEvent
	history HistoryRepo
	content *Content
}

// NewFuzzyPicker creates a new FuzzyPicker instance.
func NewFuzzyPicker(output io.Writer, hist HistoryRepo) *FuzzyPicker {
	return &FuzzyPicker{
		output:  output,
		history: hist,
		content: NewContent(),
	}
}

// SetInput sets the input channel for the picker.
func (fp *FuzzyPicker) SetInput(input <-chan core.KeyEvent) {
	fp.input = input
}

// Pick displays an interactive fuzzy search interface and returns the selected request.
// Returns the selected request string or an error if interrupted or input unavailable.
func (fp *FuzzyPicker) Pick(ctx context.Context) (string, error) {
	if fp.input == nil {
		return "", fmt.Errorf("input stream is not set")
	}

	// Clear screen and show initial state
	if err := fp.render(""); err != nil {
		return "", err
	}

	state := &pickerState{
		query:       "",
		selectedIdx: 0,
	}

	for {
		select {
		case <-ctx.Done():
			return "", core.ErrInterrupted
		case e, ok := <-fp.input:
			if !ok {
				return "", fmt.Errorf("keyboard stream was unexpectedly closed")
			}

			result, done, err := fp.handlePickerKey(e, state)
			if err != nil {
				return "", err
			}

			if done {
				return result, nil
			}
		}
	}
}

// pickerState holds the current state of the picker.
type pickerState struct {
	query       string
	selectedIdx int
}

// handlePickerKey processes a single key event in the picker.
// Returns the selected result, whether picking is done, and any error.
func (fp *FuzzyPicker) handlePickerKey(e core.KeyEvent, state *pickerState) (result string, done bool, err error) {
	switch e.Key {
	case core.KeyCtrlC, core.KeyCtrlD, core.KeyEsc:
		_ = fp.clearPicker()
		return "", true, core.ErrInterrupted

	case core.KeyEnter:
		return fp.handleEnter(state)

	case core.KeyArrowUp:
		return fp.handleArrowUp(state)

	case core.KeyArrowDown:
		return fp.handleArrowDown(state)

	case core.KeyBackspace, MacOSDeleteKey:
		return fp.handleBackspace(state)

	case core.KeyCtrlU:
		return fp.handleClearQuery(state)

	default:
		return fp.handleCharInput(e, state)
	}
}

// handleEnter processes the Enter key to select the current item.
func (fp *FuzzyPicker) handleEnter(state *pickerState) (result string, done bool, err error) {
	_ = fp.clearPicker()
	matches := fp.history.FuzzySearch(state.query)

	if len(matches) > 0 && state.selectedIdx < len(matches) {
		return matches[state.selectedIdx].Request, true, nil
	}

	return "", true, nil
}

// handleArrowUp moves selection up in the list.
func (fp *FuzzyPicker) handleArrowUp(state *pickerState) (result string, done bool, err error) {
	if state.selectedIdx > 0 {
		state.selectedIdx--

		if err := fp.render(state.query); err != nil {
			return "", true, err
		}

		if err := fp.highlightSelection(state.selectedIdx, fp.history.FuzzySearch(state.query)); err != nil {
			return "", true, err
		}
	}

	return "", false, nil
}

// handleArrowDown moves selection down in the list.
func (fp *FuzzyPicker) handleArrowDown(state *pickerState) (result string, done bool, err error) {
	matches := fp.history.FuzzySearch(state.query)
	maxIdx := len(matches) - 1

	if maxIdx > maxDisplayLines-1 {
		maxIdx = maxDisplayLines - 1
	}

	if state.selectedIdx < maxIdx {
		state.selectedIdx++

		if err := fp.render(state.query); err != nil {
			return "", true, err
		}

		if err := fp.highlightSelection(state.selectedIdx, matches); err != nil {
			return "", true, err
		}
	}

	return "", false, nil
}

// handleBackspace removes the last character from the query.
func (fp *FuzzyPicker) handleBackspace(state *pickerState) (result string, done bool, err error) {
	if state.query != "" {
		state.query = state.query[:len(state.query)-1]
		state.selectedIdx = 0

		if err := fp.render(state.query); err != nil {
			return "", true, err
		}

		if err := fp.highlightSelection(state.selectedIdx, fp.history.FuzzySearch(state.query)); err != nil {
			return "", true, err
		}
	}

	return "", false, nil
}

// handleClearQuery clears the entire query.
func (fp *FuzzyPicker) handleClearQuery(state *pickerState) (result string, done bool, err error) {
	state.query = ""
	state.selectedIdx = 0

	if err := fp.render(state.query); err != nil {
		return "", true, err
	}

	if err := fp.highlightSelection(state.selectedIdx, fp.history.FuzzySearch(state.query)); err != nil {
		return "", true, err
	}

	return "", false, nil
}

// handleCharInput adds a character to the query.
func (fp *FuzzyPicker) handleCharInput(e core.KeyEvent, state *pickerState) (result string, done bool, err error) {
	if e.Key == 0 && e.Rune != 0 && e.Rune != '\n' {
		state.query += string(e.Rune)
		state.selectedIdx = 0

		if err := fp.render(state.query); err != nil {
			return "", true, err
		}

		if err := fp.highlightSelection(state.selectedIdx, fp.history.FuzzySearch(state.query)); err != nil {
			return "", true, err
		}
	}

	return "", false, nil
}

// render displays the fuzzy picker interface with current query and matches.
func (fp *FuzzyPicker) render(query string) error {
	// Move cursor to beginning and clear from cursor to end of screen
	output := "\r\033[K"

	// Show prompt and query
	output += pickerPrompt + query + "\n"

	// Get matches
	matches := fp.history.FuzzySearch(query)

	// Display up to maxDisplayLines matches
	displayCount := len(matches)
	if displayCount > maxDisplayLines {
		displayCount = maxDisplayLines
	}

	for i := 0; i < displayCount; i++ {
		match := matches[i]
		display := formatMatchLine(match, i == 0)
		output += display + "\n"
	}

	// Show count if there are more matches
	if len(matches) > maxDisplayLines {
		output += fmt.Sprintf("\033[90m... and %d more\033[0m\n", len(matches)-maxDisplayLines)
	} else if len(matches) == 0 {
		output += "\033[90mNo matches\033[0m\n"
	}

	// Move cursor back to query line
	linesToMove := displayCount + 1
	if len(matches) > maxDisplayLines || len(matches) == 0 {
		linesToMove++
	}

	output += fmt.Sprintf("\033[%dA", linesToMove)
	output += fmt.Sprintf("\r%s%s", pickerPrompt, query)

	_, err := fmt.Fprint(fp.output, output)

	return err
}

// highlightSelection updates the display to show the selected item.
func (fp *FuzzyPicker) highlightSelection(selectedIdx int, matches []history.FuzzyMatch) error {
	// Move to first match line
	output := "\r\033[1B\033[K"

	displayCount := len(matches)
	if displayCount > maxDisplayLines {
		displayCount = maxDisplayLines
	}

	for i := 0; i < displayCount; i++ {
		if i > 0 {
			output += "\r\033[K"
		}

		match := matches[i]
		display := formatMatchLine(match, i == selectedIdx)
		output += display

		if i < displayCount-1 {
			output += "\n"
		}
	}

	// Move cursor back to query line
	output += fmt.Sprintf("\033[%dA", displayCount)
	output += "\r" + pickerPrompt

	_, err := fmt.Fprint(fp.output, output)

	return err
}

// formatMatchLine formats a match for display with highlighting.
func formatMatchLine(match history.FuzzyMatch, isSelected bool) string {
	const maxLineLength = 100

	display := match.Request
	// Truncate long lines
	if len(display) > maxLineLength {
		display = display[:maxLineLength-3] + "..."
	}

	// Replace newlines with spaces for display
	display = strings.ReplaceAll(display, "\n", " ")

	if isSelected {
		return fmt.Sprintf("\033[7m> %s\033[0m", display) // Reverse video for selection
	}

	return fmt.Sprintf("  %s", display)
}

// clearPicker clears the picker interface from the screen.
func (fp *FuzzyPicker) clearPicker() error {
	// Clear current line and all lines below
	output := "\r\033[K\033[J"
	_, err := fmt.Fprint(fp.output, output)

	return err
}

package edit

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

const (
	LineUp         = "\x1b[1A"
	LineClear      = "\x1b[2K"
	NewLine        = '\n'
	ReturnCarriege = "\r"
	Backspace      = "\b"
)

type Content struct {
	text []rune
	pos  int
}

// NewContent returns a new instance of Content with empty text and position set to 0.
func NewContent() *Content {
	return &Content{
		text: []rune{},
		pos:  0,
	}
}

// ReplaceText replaces the current text with the given text and returns the resulting string.
func (c *Content) ReplaceText(text string) string {
	output := c.Clear()
	c.text = []rune(text)
	c.pos = len(c.text)

	return output + text
}

// String returns the string representation of the Content.
func (c *Content) String() string {
	return string(c.text)
}

// ToRequest returns the content as a trimmed string.
func (c *Content) ToRequest() string {
	return strings.TrimSpace(c.String())
}

// MovePositionLeft moves the cursor position one character to the left and returns the ANSI escape sequence
// required to move the cursor to the new position. If the cursor is already at the beginning of the line, it moves
// the cursor to the end of the previous line. If the cursor is already at the beginning of the content, it returns an empty string.
func (c *Content) MovePositionLeft() string {
	if c.pos <= 0 {
		return ""
	}

	c.pos--
	if c.text[c.pos] != NewLine {
		return Backspace
	}

	startPrevLine := lastIndexOf(c.text, c.pos-1, NewLine)
	if startPrevLine == -1 {
		startPrevLine = 0
	} else {
		startPrevLine++
	}

	return LineUp + string(c.text[startPrevLine:c.pos])
}

// MovePositionRight moves the position of the cursor to the right by one character in the content.
// If the position is already at the end of the content, it returns an empty string.
func (c *Content) MovePositionRight() string {
	if c.pos >= len(c.text) {
		return ""
	}

	c.pos++

	return string(c.text[c.pos-1])
}

// MoveToNextWord moves the cursor to the beginning of the next word in the content and returns the word.
func (c *Content) MoveToNextWord() string {
	if c.pos >= len(c.text) {
		return ""
	}

	pos := c.pos

	// move to the end of the current word
	for pos < len(c.text) && isWord(c.text[pos]) {
		pos++
	}

	// move to the beginning of the next word
	for pos < len(c.text) && !isWord(c.text[pos]) {
		pos++
	}

	if pos == c.pos {
		return ""
	}

	output := string(c.text[c.pos:pos])
	c.pos = pos

	return output
}

// MoveToPrevWord moves the cursor to the beginning of the previous word in the content and returns the word.
func (c *Content) MoveToPrevWord() string {
	if c.pos <= 0 {
		return ""
	}

	pos := c.pos - 1

	// Handle case where case in the beginning of the word
	if pos > 0 && isWord(c.text[pos]) {
		pos--
	}

	// Move to the end of previous word
	for pos > 0 && !isWord(c.text[pos]) {
		pos--
	}

	// Move to the beginning of the previous word
	for pos > 0 && isWord(c.text[pos-1]) {
		pos--
	}

	buf := bytes.NewBufferString("")
	for i := c.pos - 1; pos <= i; i-- {
		fmt.Fprint(buf, c.MovePositionLeft())
	}

	return buf.String()
}

// Clear clears the content and returns the string representation of the cleared content.
// If the content is already empty, it returns an empty string.
func (c *Content) Clear() string {
	if len(c.text) == 0 {
		return ""
	}

	output := c.MoveToEnd()
	output += LineClear + ReturnCarriege

	for i := 0; i < len(c.text); i++ {
		if c.text[i] == NewLine {
			output += LineUp + LineClear + ReturnCarriege
		}
	}

	c.text = []rune{}
	c.pos = 0

	return output
}

// RemoveSymbol removes the symbol at the current position of the Content object and returns the string representation of the changes made.
// If the current position is out of bounds, an empty string is returned.
// If the removed symbol is not a newline character, the function returns the string representation of the changes made, including clearing the current line and moving the cursor to the beginning of the line.
// If the removed symbol is a newline character, the function returns the string representation of the changes made, including moving the cursor up one line and clearing the current line.
func (c *Content) RemoveSymbol() string {
	if c.pos < 1 || c.pos > len(c.text) {
		return ""
	}

	c.pos--
	symbol := c.text[c.pos]

	startCurrentLine, lines := c.GetLinesAfterPosition(c.pos)

	buffer := c.text[:c.pos]

	if c.pos < (len(c.text) - 1) {
		buffer = append(buffer, c.text[c.pos+1:]...)
	}

	c.text = buffer

	if c.pos == len(c.text) && symbol != NewLine {
		return Backspace + " " + Backspace
	}

	if symbol != NewLine {
		endCurrentLine := startCurrentLine + len(lines[0])
		return LineClear + ReturnCarriege + string(c.text[startCurrentLine:endCurrentLine-1]) + ReturnCarriege + string(c.text[startCurrentLine:c.pos])
	}

	output := LineUp + LineClear + ReturnCarriege + lines[0]
	moveUp := ""

	for i := 1; i < len(lines); i++ {
		output += lines[i] + string(NewLine) + LineClear + ReturnCarriege
		moveUp += LineUp
	}

	if moveUp != "" {
		output += moveUp
		output += ReturnCarriege + lines[0]
	}

	return output
}

// InsertSymbol inserts a rune at the current position of the Content object.
// If the position is invalid, it returns an empty string.
// If the inserted symbol is not a newline and the next character is a newline, it returns the inserted symbol.
// If the inserted symbol is a newline, it returns the content of the lines affected by the insertion, with the cursor moved to the beginning of the next line.
func (c *Content) InsertSymbol(symbol rune) string {
	if c.pos < 0 || c.pos > len(c.text) {
		return ""
	}

	if c.pos == len(c.text) {
		c.text = append(c.text, symbol)
		c.pos++

		return string(symbol)
	}

	buffer := make([]rune, c.pos, len(c.text)+1)
	copy(buffer, c.text[:c.pos])
	buffer = append(buffer, symbol)
	buffer = append(buffer, c.text[c.pos:]...)
	c.pos++
	c.text = buffer

	if symbol != NewLine && c.text[c.pos] == NewLine {
		return string(symbol)
	}

	startCurrentLine, lines := c.GetLinesAfterPosition(c.pos - 1)

	if symbol != NewLine {
		// here probably i have a room for optimization
		endCurrentLine := startCurrentLine + len(lines[0])
		return LineClear + ReturnCarriege + string(c.text[startCurrentLine:endCurrentLine]) + ReturnCarriege + string(c.text[startCurrentLine:c.pos])
	}

	output := ""
	for i := 0; i < len(lines); i++ {
		output += LineClear + ReturnCarriege + lines[i]
		if i < len(lines)-1 {
			output += string(NewLine)
		}
	}

	moveUp := ""
	for i := 1; i < len(lines)-1; i++ {
		moveUp += LineUp
	}

	output += moveUp + ReturnCarriege

	return output
}

// MoveToEnd moves the cursor to the end of the content and returns the remaining text.
// If the cursor is already at the end, it returns an empty string.
func (c *Content) MoveToEnd() string {
	if c.pos >= len(c.text) {
		return ""
	}

	output := string(c.text[c.pos:])
	c.pos = len(c.text)

	return output
}

// GetLinesAfterPosition returns the start index of the line containing the given position
// and a slice of strings representing the lines after the given position.
// If the position is before the first line, the start index is 0.
func (c *Content) GetLinesAfterPosition(pos int) (startOfLine int, lines []string) {
	startOfLine = lastIndexOf(c.text, pos-1, NewLine)
	if startOfLine == -1 {
		startOfLine = 0
	} else {
		startOfLine++
	}

	return startOfLine, strings.Split(string(c.text[startOfLine:]), string(NewLine))
}

// PrevSymbol returns the symbol before the current position in the content text.
// If the current position is at the beginning of the text, it returns 0.
func (c *Content) PrevSymbol() rune {
	if c.pos <= 0 {
		return 0
	}

	return c.text[c.pos-1]
}

// lastIndexOf returns the index of the last occurrence of the given rune in the buffer, starting from the given position.
// If the rune is not found, it returns -1.
func lastIndexOf(buffer []rune, pos int, search rune) int {
	for i := pos; i >= 0; i-- {
		if buffer[i] == search {
			return i
		}
	}

	return -1
}

// isWord returns true if the given rune is a digit or a letter.
func isWord(r rune) bool {
	return unicode.IsDigit(r) || unicode.IsLetter(r)
}

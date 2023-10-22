package cli

import (
	"strings"
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

func NewContent() *Content {
	return &Content{
		text: []rune{},
		pos:  0,
	}
}

func (c *Content) ReplaceText(text string) string {
	output := c.Clear()
	c.text = []rune(text)
	c.pos = len(c.text)

	return output + text
}

func (c *Content) String() string {
	return string(c.text)
}

func (c *Content) ToRequest() string {
	return strings.TrimSpace(c.String())
}

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

func (c *Content) MovePositionRight() string {
	if c.pos >= len(c.text) {
		return ""
	}

	c.pos++

	return string(c.text[c.pos-1])
}

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

	return output
}

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

func (c *Content) MoveToEnd() string {
	if c.pos >= len(c.text) {
		return ""
	}

	output := string(c.text[c.pos:])
	c.pos = len(c.text)

	return output
}

func (c *Content) GetLinesAfterPosition(pos int) (startOfLine int, lines []string) {
	startOfLine = lastIndexOf(c.text, pos-1, NewLine)
	if startOfLine == -1 {
		startOfLine = 0
	} else {
		startOfLine++
	}

	return startOfLine, strings.Split(string(c.text[startOfLine:]), string(NewLine))
}

func lastIndexOf(buffer []rune, pos int, search rune) int {
	for i := pos; i >= 0; i-- {
		if buffer[i] == search {
			return i
		}
	}

	return -1
}

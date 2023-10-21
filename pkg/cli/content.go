package cli

import (
	"strings"
)

const (
	LineUp    = "\x1b[1A"
	LineClear = "\x1b[2K"
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
	if c.text[c.pos] != '\n' {
		return "\b"
	}

	startPrevLine := lastIndexOf(c.text, c.pos-1, '\n')
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
	output += LineClear + "\r"

	for i := 0; i < len(c.text); i++ {
		if c.text[i] == '\n' {
			output += LineUp + LineClear + "\r"
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

	if c.pos == len(c.text) && symbol != '\n' {
		return "\b \b"
	}

	if symbol != '\n' {
		endCurrentLine := startCurrentLine + len(lines[0])
		return LineClear + "\r" + string(c.text[startCurrentLine:endCurrentLine-1]) + "\r" + string(c.text[startCurrentLine:c.pos])
	}

	output := LineUp + LineClear + "\r" + lines[0]

	for i := 1; i < len(lines); i++ {
		output += lines[i] + "\n" + LineClear + "\r"
	}

	// Move cursor back to position
	for i := 1; i < len(lines); i++ {
		output += LineUp
	}

	output += "\r" + lines[0]

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

	if symbol != '\n' && c.text[c.pos] == '\n' {
		return string(symbol)
	}

	startCurrentLine, lines := c.GetLinesAfterPosition(c.pos - 1)

	if symbol != '\n' {
		// here probably i have a room for optimization
		endCurrentLine := startCurrentLine + len(lines[0])
		return LineClear + "\r" + string(c.text[startCurrentLine:endCurrentLine]) + "\r" + string(c.text[startCurrentLine:c.pos])
	}

	output := ""

	for i := 0; i < len(lines); i++ {
		output += LineClear + "\r" + lines[i]
		if i < len(lines)-1 {
			output += "\n"
		}
	}

	// Move cursor back to position
	for i := 2; i < len(lines); i++ {
		output += LineUp
	}

	output += "\r"

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
	if pos < 0 || pos > len(c.text) {
		return 0, []string{}
	}

	startOfLine = lastIndexOf(c.text, pos, '\t')
	if startOfLine == -1 {
		startOfLine = 0
	} else {
		startOfLine++
	}

	return startOfLine, strings.Split(string(c.text[startOfLine:]), "\n")
}

func lastIndexOf(buffer []rune, pos int, search rune) int {
	for i := pos; i >= 0; i-- {
		if buffer[i] == search {
			return i
		}
	}

	return -1
}

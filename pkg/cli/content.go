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

	output := LineClear + "\r"

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

	lines := c.GetLinesAfterPosition(c.pos)

	buffer := c.text[:c.pos]

	if c.pos < (len(c.text) - 1) {
		buffer = append(buffer, c.text[c.pos+1:]...)
	}

	c.text = buffer

	if symbol != '\n' {
		return "\b \b"
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
	buffer := make([]rune, c.pos, len(c.text)+1)
	copy(buffer, c.text[:c.pos])
	buffer = append(buffer, symbol)
	output := ""

	if symbol == '\n' && c.pos < len(c.text) {
		endOfLine := lastIndexOf(c.text, c.pos, '\n')
		if endOfLine == -1 {
			endOfLine = len(c.text)
		}

		for i := c.pos; i <= endOfLine; i++ {
			output += string(' ')
		}
	}

	output += string(symbol)

	if c.pos < len(c.text) {
		buffer = append(buffer, c.text[c.pos:]...)
		moveCursor := ""

		for i := c.pos; i < len(c.text); i++ {
			if c.text[i] != '\n' {
				output += string(c.text[i])
				moveCursor += "\b"
			} else {
				break
			}
		}

		output += moveCursor
	}

	c.text = buffer
	c.pos++

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

func (c *Content) GetLinesAfterPosition(pos int) []string {
	if pos < 0 || pos > len(c.text) {
		return []string{}
	}

	startOfLine := lastIndexOf(c.text, pos, '\t')
	if startOfLine == -1 {
		startOfLine = 0
	} else {
		startOfLine++
	}

	return strings.Split(string(c.text[startOfLine:]), "\n")
}

func lastIndexOf(buffer []rune, pos int, search rune) int {
	for i := pos; i >= 0; i-- {
		if buffer[i] == search {
			return i
		}
	}

	return -1
}

// BEGIN: yz9d8f7g6h5j
package cli

import (
	"testing"
)

func TestNewContent(t *testing.T) {
	c := NewContent()

	if c == nil {
		t.Error("NewContent() returned nil")
		return
	}

	if c.pos != 0 {
		t.Errorf("NewContent() returned Content with pos %d, expected 0", c.pos)
	}

	if len(c.text) != 0 {
		t.Errorf("NewContent() returned Content with text %v, expected empty slice", c.text)
	}
}

func TestReplaceEmptyText(t *testing.T) {
	c := NewContent()
	text := "TestReplaceEmptyText"
	output := c.ReplaceText(text)

	if c.String() != text {
		t.Errorf("ReplaceText() did not set Content text correctly, expected %s, got %s", text, c.String())
	}

	if c.pos != len(text) {
		t.Errorf("ReplaceText() did not set Content pos correctly, expected %d, got %d", len(text), c.pos)
	}

	if output != text {
		t.Errorf("ReplaceText() did not return correct output, expected %s, got %s", c.Clear()+text, output)
	}
}

func TestReplaceNonEmptyText(t *testing.T) {
	c := NewContent()
	c.ReplaceText("buy")

	text := "hello"
	output := c.ReplaceText(text)

	if c.String() != text {
		t.Errorf("ReplaceText() did not set Content text correctly, expected %s, got %s", text, c.String())
	}

	if c.pos != len(text) {
		t.Errorf("ReplaceText() did not set Content pos correctly, expected %d, got %d", len(text), c.pos)
	}

	expected := "\x1b[2K\r" + text
	if output != expected {
		t.Errorf("ReplaceText() did not return correct output, expected %q, got %q", expected, output)
	}
}

func TestReplaceMultiLineText(t *testing.T) {
	c := NewContent()
	c.ReplaceText("buy\nsome milk")

	text := "hello"
	output := c.ReplaceText(text)

	if c.String() != text {
		t.Errorf("ReplaceText() did not set Content text correctly, expected %s, got %s", text, c.String())
	}

	if c.pos != len(text) {
		t.Errorf("ReplaceText() did not set Content pos correctly, expected %d, got %d", len(text), c.pos)
	}

	expected := "\x1b[2K\r\033[1A\x1b[2K\r" + text
	if output != expected {
		t.Errorf("ReplaceText() did not return correct output, expected %q, got %q", expected, output)
	}
}

func TestString(t *testing.T) {
	c := NewContent()
	text := "hello world"
	c.ReplaceText(text)

	if c.String() != text {
		t.Errorf("String() did not return correct string, expected %s, got %s", text, c.String())
	}
}

func TestToRequest(t *testing.T) {
	c := NewContent()
	text := " hello world\n"
	c.ReplaceText(text)

	if c.ToRequest() != "hello world" {
		t.Errorf("ToRequest() did not return correct string, expected 'hello world', got '%s'", c.ToRequest())
	}
}

func TestContent_Clear(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty content",
			input:    "",
			expected: "",
		},
		{
			name:     "single line content",
			input:    "hello world",
			expected: "\x1b[2K\r",
		},
		{
			name:     "multi-line content",
			input:    "hello\nworld",
			expected: "\x1b[2K\r\x1b[1A\x1b[2K\r",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Content{
				text: []rune(tt.input),
				pos:  len(tt.input),
			}

			actual := c.Clear()
			if actual != tt.expected {
				t.Errorf("unexpected output: got %q, want %q", actual, tt.expected)
			}
		})
	}
}

func TestContent_MovePositionLeft(t *testing.T) {
	tests := []struct {
		name        string
		content     *Content
		expected    string
		expectedPos int
	}{
		{
			name: "move left when position is at the beginning",
			content: &Content{
				text: []rune("hello"),
				pos:  0,
			},
			expected:    "",
			expectedPos: 0,
		},
		{
			name: "move left when position is not at the beginning and previous character is not a newline",
			content: &Content{
				text: []rune("hello"),
				pos:  3,
			},
			expected:    "\b",
			expectedPos: 2,
		},
		{
			name: "move left when position is not at the beginning and previous character is a newline",
			content: &Content{
				text: []rune("hello\nworld"),
				pos:  6,
			},
			expected:    LineUp + "hello",
			expectedPos: 5,
		},
		{
			name: "move left when position is not at the beginning and previous character is a newline and there is no previous line",
			content: &Content{
				text: []rune("\nhello"),
				pos:  1,
			},
			expected:    LineUp,
			expectedPos: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.content.MovePositionLeft()
			if actual != tt.expected {
				t.Errorf("expected %q, but got %q", tt.expected, actual)
			}
			if tt.content.pos != tt.expectedPos {
				t.Errorf("expected position %d, but got %d", tt.expectedPos, tt.content.pos)
			}
		})
	}
}

func TestContent_MovePositionRight(t *testing.T) {
	tests := []struct {
		name        string
		content     *Content
		expected    string
		expectedPos int
	}{
		{
			name:     "empty content",
			content:  NewContent(),
			expected: "",
		},
		{
			name: "single character",
			content: &Content{
				text: []rune{'a'},
				pos:  0,
			},
			expected:    "a",
			expectedPos: 1,
		},
		{
			name: "multiple characters",
			content: &Content{
				text: []rune{'a', 'b', 'c'},
				pos:  1,
			},
			expected:    "b",
			expectedPos: 2,
		},
		{
			name: "at end of content",
			content: &Content{
				text: []rune{'a', 'b', 'c'},
				pos:  3,
			},
			expected:    "",
			expectedPos: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.content.MovePositionRight()
			if actual != tt.expected {
				t.Errorf("expected %q, but got %q", tt.expected, actual)
			}
			if tt.content.pos != tt.expectedPos {
				t.Errorf("expected position %d, but got %d", tt.expectedPos, tt.content.pos)
			}
		})
	}
}

func TestContent_RemoveSymbol(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		output       string
		contentAfter string
		pos          int
	}{
		{
			name:         "remove symbol in the middle of the text",
			input:        "hello world",
			pos:          6,
			output:       "\b \b",
			contentAfter: "helloworld",
		},
		{
			name:         "remove symbol at the beginning of the text",
			input:        "hello world",
			pos:          0,
			output:       "",
			contentAfter: "hello world",
		},
		{
			name:         "remove symbol at the end of the text",
			input:        "hello world",
			pos:          11,
			output:       "\b \b",
			contentAfter: "hello worl",
		},
		{
			// This test is not really correct, but it is how it works now
			// TODO: fix this test
			name:         "remove newline symbol",
			input:        "hello\nworld",
			pos:          6,
			output:       "\x1b[1Ahelloworld",
			contentAfter: "helloworld",
		},
		{
			name:         "remove symbol when pos is out of range",
			input:        "hello world",
			pos:          20,
			output:       "",
			contentAfter: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := &Content{
				text: []rune(tt.input),
				pos:  tt.pos,
			}

			output := content.RemoveSymbol()

			if output != tt.output {
				t.Errorf("expected output %q, but got %q", tt.output, output)
			}

			if tt.output != "" && content.pos != tt.pos-1 {
				t.Errorf("expected position to be '%d', but got '%d'", tt.pos, content.pos)
			}

			if string(content.text) != tt.contentAfter {
				t.Errorf("expected text to be '%s', but got '%s'", tt.contentAfter, string(content.text))
			}
		})
	}
}

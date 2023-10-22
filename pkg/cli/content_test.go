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
			output:       "\x1b[2K\rhelloworld\rhello",
			contentAfter: "helloworld",
		},
		{
			name:         "remove symbol in the middle of the multiline text",
			input:        "hello\nworld",
			pos:          4,
			output:       "\x1b[2K\rhelo\rhel",
			contentAfter: "helo\nworld",
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
			name:         "remove newline symbol in the middle of the text",
			input:        "hello\nworld",
			pos:          6,
			output:       "\x1b[1A\x1b[2K\rhelloworld\n\x1b[2K\r\x1b[1A\rhello",
			contentAfter: "helloworld",
		},
		{
			name:         "remove newline symbol at the beginning of the text",
			input:        "\nhello\nworld",
			pos:          1,
			output:       "\x1b[1A\x1b[2K\rhello\n\x1b[2K\rworld\n\x1b[2K\r\x1b[1A\x1b[1A\r",
			contentAfter: "hello\nworld",
		},
		{
			name:         "remove newline symbol at the end of the text",
			input:        "hello\nworld\n",
			pos:          12,
			output:       "\x1b[1A\x1b[2K\rworld\n\x1b[2K\r\x1b[1A\rworld",
			contentAfter: "hello\nworld",
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
				t.Errorf("expected text to be '%q', but got '%q'", tt.contentAfter, string(content.text))
			}
		})
	}
}

func TestContent_InsertSymbol(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		output       string
		contentAfter string
		symbol       rune
		pos          int
		posAfter     int
	}{
		{
			name:         "insert at the end",
			input:        "hello world",
			symbol:       '!',
			pos:          11,
			output:       "!",
			contentAfter: "hello world!",
			posAfter:     12,
		},
		{
			name:         "insert at the beginning",
			input:        "hello world",
			symbol:       '>',
			pos:          0,
			output:       "\x1b[2K\r>hello world\r>",
			contentAfter: ">hello world",
			posAfter:     1,
		},
		{
			name:         "insert in the middle",
			input:        "hello world",
			symbol:       ',',
			pos:          5,
			output:       "\x1b[2K\rhello, world\rhello,",
			contentAfter: "hello, world",
			posAfter:     6,
		},
		{
			name:         "insert newline in the middle",
			input:        "hello\nworld",
			symbol:       '\n',
			pos:          4,
			output:       "\x1b[2K\rhell\n\x1b[2K\ro\n\x1b[2K\rworld\x1b[1A\r",
			contentAfter: "hell\no\nworld",
			posAfter:     5,
		},
		{
			name:         "insert newline at the end",
			input:        "hello world",
			symbol:       '\n',
			pos:          11,
			output:       "\n",
			contentAfter: "hello world\n",
			posAfter:     12,
		},
		{
			name:         "insert newline at the begginning",
			input:        "hello\nworld",
			symbol:       '\n',
			pos:          0,
			output:       "\x1b[2K\r\n\x1b[2K\rhello\n\x1b[2K\rworld\x1b[1A\r",
			contentAfter: "\nhello\nworld",
			posAfter:     1,
		},
		{
			name:         "insert newline in a row",
			input:        "h\n\nello\nworld",
			symbol:       '\n',
			pos:          2,
			output:       "\x1b[2K\r\n\x1b[2K\r\n\x1b[2K\rello\n\x1b[2K\rworld\x1b[1A\x1b[1A\r",
			contentAfter: "h\n\n\nello\nworld",
			posAfter:     3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := &Content{
				text: []rune(tt.input),
				pos:  tt.pos,
			}

			output := content.InsertSymbol(tt.symbol)

			if output != tt.output {
				t.Errorf("expected output %q, but got %q", tt.output, output)
			}

			if tt.output != "" && content.pos != tt.pos+1 {
				t.Errorf("expected position to be '%d', but got '%d'", tt.pos, content.pos)
			}

			if string(content.text) != tt.contentAfter {
				t.Errorf("expected text to be '%q', but got '%q'", tt.contentAfter, string(content.text))
			}

			if content.pos != tt.posAfter {
				t.Errorf("expected position to be '%d', but got '%d'", tt.posAfter, content.pos)
			}
		})
	}
}

func TestContent_MoveToEnd(t *testing.T) {
	cases := []struct {
		name     string
		content  *Content
		expected string
	}{
		{
			name:     "empty content",
			content:  &Content{},
			expected: "",
		},
		{
			name: "cursor at the end",
			content: &Content{
				text: []rune("hello world"),
				pos:  11,
			},
			expected: "",
		},
		{
			name: "cursor at the beginning",
			content: &Content{
				text: []rune("hello world"),
				pos:  0,
			},
			expected: "hello world",
		},
		{
			name: "cursor in the middle",
			content: &Content{
				text: []rune("hello world"),
				pos:  6,
			},
			expected: "world",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.content.MoveToEnd()
			if actual != tc.expected {
				t.Errorf("expected %q but got %q", tc.expected, actual)
			}
			if tc.content.pos != len(tc.content.text) {
				t.Errorf("expected cursor to be at the end of the text")
			}
		})
	}
}

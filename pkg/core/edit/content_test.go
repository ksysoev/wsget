package edit

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

			output := content.RemovePrevSymbol()

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

func TestContent_MoveToNextWord(t *testing.T) {
	tests := []struct {
		name        string
		content     *Content
		expected    string
		expectedPos int
	}{
		{
			name: "move to next word when cursor is at the beginning",
			content: &Content{
				text: []rune("hello world"),
				pos:  0,
			},
			expected:    "hello ",
			expectedPos: 6,
		},
		{
			name: "move to next word when cursor is in the middle of a word",
			content: &Content{
				text: []rune("hello world"),
				pos:  2,
			},
			expected:    "llo ",
			expectedPos: 6,
		},
		{
			name: "move to next word when cursor is at the end of a word",
			content: &Content{
				text: []rune("hello world"),
				pos:  5,
			},
			expected:    " ",
			expectedPos: 6,
		},
		{
			name: "move to next word when cursor is at the beginning of a word",
			content: &Content{
				text: []rune("hello world"),
				pos:  6,
			},
			expected:    "world",
			expectedPos: 11,
		},
		{
			name: "move to next word when cursor is at the end of the content",
			content: &Content{
				text: []rune("hello world"),
				pos:  11,
			},
			expected:    "",
			expectedPos: 11,
		},
		{
			name: "move to next word when there are multiple spaces",
			content: &Content{
				text: []rune("hello   world"),
				pos:  5,
			},
			expected:    "   ",
			expectedPos: 8,
		},
		{
			name: "move to next word when there are no more words",
			content: &Content{
				text: []rune("hello"),
				pos:  5,
			},
			expected:    "",
			expectedPos: 5,
		},
		{
			name: "move to next word when cursor is at the beginning of an empty content",
			content: &Content{
				text: []rune(""),
				pos:  0,
			},
			expected:    "",
			expectedPos: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.content.MoveToNextWord()
			if actual != tt.expected {
				t.Errorf("expected %q, but got %q", tt.expected, actual)
			}

			if tt.content.pos != tt.expectedPos {
				t.Errorf("expected position %d, but got %d", tt.expectedPos, tt.content.pos)
			}
		})
	}
}

func TestContent_MoveToPrevWord(t *testing.T) {
	tests := []struct {
		name        string
		content     *Content
		expected    string
		expectedPos int
	}{
		{
			name: "move to previous word when cursor is at the end of a word",
			content: &Content{
				text: []rune("hello world"),
				pos:  11,
			},
			expectedPos: 6,
		},
		{
			name: "move to previous word when cursor is in the middle of a word",
			content: &Content{
				text: []rune("hello world"),
				pos:  8,
			},
			expectedPos: 6,
		},
		{
			name: "move to previous word when cursor is at the beginning of a word",
			content: &Content{
				text: []rune("hello world"),
				pos:  6,
			},
			expectedPos: 0,
		},
		{
			name: "move to previous word when cursor is at the beginning of the content",
			content: &Content{
				text: []rune("hello world"),
				pos:  0,
			},
			expectedPos: 0,
		},
		{
			name: "move to previous word when there are multiple spaces",
			content: &Content{
				text: []rune("hello   world"),
				pos:  11,
			},
			expectedPos: 8,
		},
		{
			name: "move to previous word when there are no more words",
			content: &Content{
				text: []rune("hello"),
				pos:  5,
			},
			expectedPos: 0,
		},
		{
			name: "move to previous word when cursor is at the beginning of an empty content",
			content: &Content{
				text: []rune(""),
				pos:  0,
			},
			expectedPos: 0,
		},
		{
			name: "move to previous word when cursor is at the end of a single word",
			content: &Content{
				text: []rune("word"),
				pos:  4,
			},
			expectedPos: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.content.MoveToPrevWord()

			if tt.content.pos != tt.expectedPos {
				t.Errorf("expected position %d, but got %d", tt.expectedPos, tt.content.pos)
			}
		})
	}
}

func TestContent_MoveToRowStart(t *testing.T) {
	tests := []struct {
		name        string
		content     *Content
		expected    string
		expectedPos int
	}{
		{
			name: "cursor at the beginning of the content",
			content: &Content{
				text: []rune("hello world"),
				pos:  0,
			},
			expected:    "",
			expectedPos: 0,
		},
		{
			name: "cursor at the beginning of a line",
			content: &Content{
				text: []rune("hello\nworld"),
				pos:  6,
			},
			expected:    "",
			expectedPos: 6,
		},
		{
			name: "cursor in the middle of a line",
			content: &Content{
				text: []rune("hello\nworld"),
				pos:  8,
			},
			expected:    "\b\b",
			expectedPos: 6,
		},
		{
			name: "cursor at the end of a line",
			content: &Content{
				text: []rune("hello world"),
				pos:  11,
			},
			expected:    "\b\b\b\b\b\b\b\b\b\b\b",
			expectedPos: 0,
		},
		{
			name: "cursor in the middle of a multiline content",
			content: &Content{
				text: []rune("hello\nworld\nfoo"),
				pos:  11,
			},
			expected:    "\b\b\b\b\b",
			expectedPos: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.content.MoveToRowStart()
			if actual != tt.expected {
				t.Errorf("expected %q, but got %q", tt.expected, actual)
			}

			if tt.content.pos != tt.expectedPos {
				t.Errorf("expected position %d, but got %d", tt.expectedPos, tt.content.pos)
			}
		})
	}
}
func TestContent_MoveToRowEnd(t *testing.T) {
	tests := []struct {
		name        string
		content     *Content
		expected    string
		expectedPos int
	}{
		{
			name: "cursor at the beginning of a single line",
			content: &Content{
				text: []rune("hello world"),
				pos:  0,
			},
			expected:    "hello world",
			expectedPos: 11,
		},
		{
			name: "cursor in the middle of a single line",
			content: &Content{
				text: []rune("hello world"),
				pos:  6,
			},
			expected:    "world",
			expectedPos: 11,
		},
		{
			name: "cursor at the end of a single line",
			content: &Content{
				text: []rune("hello world"),
				pos:  11,
			},
			expected:    "",
			expectedPos: 11,
		},
		{
			name: "cursor at the beginning of a multi-line content",
			content: &Content{
				text: []rune("hello\nworld"),
				pos:  0,
			},
			expected:    "hello",
			expectedPos: 5,
		},
		{
			name: "cursor in the middle of a multi-line content",
			content: &Content{
				text: []rune("hello\nworld"),
				pos:  3,
			},
			expected:    "lo",
			expectedPos: 5,
		},
		{
			name: "cursor at the end of a line in multi-line content",
			content: &Content{
				text: []rune("hello\nworld"),
				pos:  5,
			},
			expected:    "",
			expectedPos: 5,
		},
		{
			name: "cursor at the beginning of the second line in multi-line content",
			content: &Content{
				text: []rune("hello\nworld"),
				pos:  6,
			},
			expected:    "world",
			expectedPos: 11,
		},
		{
			name: "cursor in the middle of the second line in multi-line content",
			content: &Content{
				text: []rune("hello\nworld"),
				pos:  8,
			},
			expected:    "rld",
			expectedPos: 11,
		},
		{
			name: "cursor at the end of the second line in multi-line content",
			content: &Content{
				text: []rune("hello\nworld"),
				pos:  11,
			},
			expected:    "",
			expectedPos: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.content.MoveToRowEnd()
			if actual != tt.expected {
				t.Errorf("expected %q, but got %q", tt.expected, actual)
			}

			if tt.content.pos != tt.expectedPos {
				t.Errorf("expected position %d, but got %d", tt.expectedPos, tt.content.pos)
			}
		})
	}
}

func TestContent_RemoveNextSymbol(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		output       string
		expectedText string
		pos          int
		expectedPos  int
	}{
		{
			name:         "remove symbol in the middle of the text",
			input:        "hello world",
			pos:          5,
			output:       "\x1b[2K\rhelloworld\rhello",
			expectedText: "helloworld",
			expectedPos:  5,
		},
		{
			name:         "remove symbol at the beginning of the text",
			input:        "hello world",
			pos:          0,
			output:       "\x1b[2K\rello world\r",
			expectedText: "ello world",
			expectedPos:  0,
		},
		{
			name:         "remove symbol at the end of the text",
			input:        "hello world",
			pos:          11,
			output:       "",
			expectedText: "hello world",
			expectedPos:  11,
		},
		{
			name:         "remove newline character in the middle of the text",
			input:        "hello\nworld",
			pos:          5,
			output:       "world\n\x1b[2K\r\x1b[1A\rhello",
			expectedText: "helloworld",
			expectedPos:  5,
		},
		{
			name:         "remove symbol when pos is out of bounds (negative)",
			input:        "hello world",
			pos:          -1,
			output:       "",
			expectedText: "hello world",
			expectedPos:  -1,
		},
		{
			name:         "remove symbol when pos is out of bounds (exceeds length)",
			input:        "hello world",
			pos:          12,
			output:       "",
			expectedText: "hello world",
			expectedPos:  12,
		},
		{
			name:         "remove symbol from empty content",
			input:        "",
			pos:          0,
			output:       "",
			expectedText: "",
			expectedPos:  0,
		},
		{
			name:         "remove newline at the end of the text",
			input:        "hello world\n",
			pos:          11,
			output:       "\n\x1b[2K\r\x1b[1A\rhello world",
			expectedText: "hello world",
			expectedPos:  11,
		},
		{
			name:         "remove symbol at the end when cursor is before newline",
			input:        "hello\n",
			pos:          5,
			output:       "\n\x1b[2K\r\x1b[1A\rhello",
			expectedText: "hello",
			expectedPos:  5,
		},
		{
			name:         "remove symbol when next symbol is newline",
			input:        "hello\nworld",
			pos:          5,
			output:       "world\n\x1b[2K\r\x1b[1A\rhello",
			expectedText: "helloworld",
			expectedPos:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Content{
				text: []rune(tt.input),
				pos:  tt.pos,
			}

			output := c.RemoveNextSymbol()

			if output != tt.output {
				t.Errorf("expected output %q, but got %q", tt.output, output)
			}

			if string(c.text) != tt.expectedText {
				t.Errorf("expected text %q, but got %q", tt.expectedText, string(c.text))
			}

			if c.pos != tt.expectedPos {
				t.Errorf("expected position %d, but got %d", tt.expectedPos, c.pos)
			}
		})
	}
}

func TestContent_DeleteToPrevWord(t *testing.T) {
	tests := []struct {
		name           string
		inputText      string
		expectedOutput string
		expectedText   string
		inputPos       int
		expectedPos    int
	}{
		{
			name:           "Delete previous word when cursor is at the end",
			inputText:      "hello world",
			inputPos:       11,
			expectedOutput: "\b \b\b \b\b \b\b \b\b \b",
			expectedText:   "hello ",
			expectedPos:    6,
		},
		{
			name:           "Delete to previous word when cursor is in the middle of a word",
			inputText:      "hello world",
			inputPos:       8,
			expectedOutput: "\x1b[2K\rhello wrld\rhello w\x1b[2K\rhello rld\rhello ",
			expectedText:   "hello rld",
			expectedPos:    6,
		},
		{
			name:           "Delete to previous word when cursor is at the beginning of a word",
			inputText:      "hello world",
			inputPos:       6,
			expectedOutput: "\x1b[2K\rhelloworld\rhello\x1b[2K\rhellworld\rhell\x1b[2K\rhelworld\rhel\x1b[2K\rheworld\rhe\x1b[2K\rhworld\rh\x1b[2K\rworld\r",
			expectedText:   "world",
			expectedPos:    0,
		},
		{
			name:           "Delete to previous word when there is no previous word",
			inputText:      "hello world",
			inputPos:       0,
			expectedOutput: "",
			expectedText:   "hello world",
			expectedPos:    0,
		},
		{
			name:           "Delete to previous word when there are multiple spaces",
			inputText:      "hello   world",
			inputPos:       13,
			expectedOutput: "\b \b\b \b\b \b\b \b\b \b",
			expectedText:   "hello   ",
			expectedPos:    8,
		},
		{
			name:           "Delete to previous word when there is only one word",
			inputText:      "world",
			inputPos:       5,
			expectedOutput: "\b \b\b \b\b \b\b \b\b \b",
			expectedText:   "",
			expectedPos:    0,
		},
		{
			name:           "Delete to previous word in empty content",
			inputText:      "",
			inputPos:       0,
			expectedOutput: "",
			expectedText:   "",
			expectedPos:    0,
		},
		{
			name:           "Delete to previous word with punctuation",
			inputText:      "hello, world!",
			inputPos:       13,
			expectedOutput: "\b \b\b \b\b \b\b \b\b \b\b \b",
			expectedText:   "hello, ",
			expectedPos:    7,
		},
		{
			name:           "Delete to previous word with cursor in whitespace",
			inputText:      "hello    ",
			inputPos:       9,
			expectedOutput: "\b \b\b \b\b \b\b \b\b \b\b \b\b \b\b \b\b \b",
			expectedText:   "",
			expectedPos:    0,
		},
		{
			name:           "Delete to previous word with multiple non-word characters",
			inputText:      "hello---world",
			inputPos:       13,
			expectedOutput: "\b \b\b \b\b \b\b \b\b \b",
			expectedText:   "hello---",
			expectedPos:    8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Content{
				text: []rune(tt.inputText),
				pos:  tt.inputPos,
			}

			output := c.DeleteToPrevWord()

			if output != tt.expectedOutput {
				t.Errorf("expected output %q, but got %q", tt.expectedOutput, output)
			}

			if string(c.text) != tt.expectedText {
				t.Errorf("expected text %q, but got %q", tt.expectedText, string(c.text))
			}

			if c.pos != tt.expectedPos {
				t.Errorf("expected position %d, but got %d", tt.expectedPos, c.pos)
			}
		})
	}
}

func TestContent_DeleteToNextWord(t *testing.T) {
	tests := []struct {
		name           string
		inputText      string
		expectedOutput string
		expectedText   string
		inputPos       int
		expectedPos    int
	}{
		{
			name:           "Delete to next word when cursor is at the beginning of a word",
			inputText:      "hello world",
			inputPos:       0,
			expectedOutput: "\x1b[2K\rello world\r\x1b[2K\rllo world\r\x1b[2K\rlo world\r\x1b[2K\ro world\r\x1b[2K\r world\r\x1b[2K\rworld\r",
			expectedText:   "world",
			expectedPos:    0,
		},
		{
			name:           "Delete to next word when cursor is in the middle of a word",
			inputText:      "hello world",
			inputPos:       2,
			expectedOutput: "\x1b[2K\rhelo world\rhe\x1b[2K\rheo world\rhe\x1b[2K\rhe world\rhe\x1b[2K\rheworld\rhe",
			expectedText:   "heworld",
			expectedPos:    2,
		},
		{
			name:           "Delete to next word when cursor is at the end of a word",
			inputText:      "hello world",
			inputPos:       5,
			expectedOutput: "\x1b[2K\rhelloworld\rhello",
			expectedText:   "helloworld",
			expectedPos:    5,
		},
		{
			name:           "Delete to next word when cursor is at whitespace",
			inputText:      "hello   world",
			inputPos:       5,
			expectedOutput: "\x1b[2K\rhello  world\rhello\x1b[2K\rhello world\rhello\x1b[2K\rhelloworld\rhello",
			expectedText:   "helloworld",
			expectedPos:    5,
		},
		{
			name:           "Delete to next word when cursor is at the end of text",
			inputText:      "hello world",
			inputPos:       11,
			expectedOutput: "",
			expectedText:   "hello world",
			expectedPos:    11,
		},
		{
			name:           "Delete to next word when there are multiple spaces",
			inputText:      "hello   world",
			inputPos:       2,
			expectedOutput: "\x1b[2K\rhelo   world\rhe\x1b[2K\rheo   world\rhe\x1b[2K\rhe   world\rhe\x1b[2K\rhe  world\rhe\x1b[2K\rhe world\rhe\x1b[2K\rheworld\rhe",
			expectedText:   "heworld",
			expectedPos:    2,
		},
		{
			name:           "Delete to next word with punctuation",
			inputText:      "hello, world!",
			inputPos:       5,
			expectedOutput: "\x1b[2K\rhello world!\rhello\x1b[2K\rhelloworld!\rhello",
			expectedText:   "helloworld!",
			expectedPos:    5,
		},
		{
			name:           "Delete to next word when cursor is at the beginning of an empty content",
			inputText:      "",
			inputPos:       0,
			expectedOutput: "",
			expectedText:   "",
			expectedPos:    0,
		},
		{
			name:           "Delete to next word when cursor is in the middle of multiple spaces",
			inputText:      "hello     world",
			inputPos:       7,
			expectedOutput: "\x1b[2K\rhello    world\rhello  \x1b[2K\rhello   world\rhello  \x1b[2K\rhello  world\rhello  ",
			expectedText:   "hello  world",
			expectedPos:    7,
		},
		{
			name:           "Delete to next word in multi-line content",
			inputText:      "hello\nworld\nfoo",
			inputPos:       6,
			expectedOutput: "\x1b[2K\rorld\r\x1b[2K\rrld\r\x1b[2K\rld\r\x1b[2K\rd\r\x1b[2K\r\rfoo\n\x1b[2K\r\x1b[1A\r",
			expectedText:   "hello\nfoo",
			expectedPos:    6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Content{
				text: []rune(tt.inputText),
				pos:  tt.inputPos,
			}

			output := c.DeleteToNextWord()

			if output != tt.expectedOutput {
				t.Errorf("expected output %q, but got %q", tt.expectedOutput, output)
			}

			if string(c.text) != tt.expectedText {
				t.Errorf("expected text %q, but got %q", tt.expectedText, string(c.text))
			}

			if c.pos != tt.expectedPos {
				t.Errorf("expected position %d, but got %d", tt.expectedPos, c.pos)
			}
		})
	}
}

func TestContent_PrevSymbol(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pos      int
		expected rune
	}{
		{
			name:     "Cursor at the beginning",
			text:     "hello",
			pos:      0,
			expected: 0,
		},
		{
			name:     "Cursor in the middle",
			text:     "hello",
			pos:      2,
			expected: 'e',
		},
		{
			name:     "Cursor at the end",
			text:     "hello",
			pos:      5,
			expected: 'o',
		},
		{
			name:     "Cursor out of bounds (negative)",
			text:     "hello",
			pos:      -1,
			expected: 0,
		},
		{
			name:     "Cursor out of bounds (beyond length)",
			text:     "hello",
			pos:      6,
			expected: 'o',
		},
		{
			name:     "Empty text",
			text:     "",
			pos:      0,
			expected: 0,
		},
		{
			name:     "Cursor after newline character",
			text:     "hello\nworld",
			pos:      6,
			expected: '\n',
		},
		{
			name:     "Cursor at position 1",
			text:     "a",
			pos:      1,
			expected: 'a',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Content{
				text: []rune(tt.text),
				pos:  tt.pos,
			}

			if got := c.PrevSymbol(); got != tt.expected {
				t.Errorf("PrevSymbol() = %q, want %q", got, tt.expected)
			}
		})
	}
}

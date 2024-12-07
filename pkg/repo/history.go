package repo

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	HistoryFileRigths = 0o644
	DefaultLimit      = 100
)

// History maintains a record of command requests, enabling storage, retrieval, and navigation through past requests.
type History struct {
	fileName string
	requests []string
	limit    int
	pos      int
}

// NewHistory creates and returns a new History instance with default settings and an initial empty requests slice.
func NewHistory(fileName string) *History {
	return &History{
		fileName: fileName,
		limit:    DefaultLimit,
		requests: make([]string, 0, DefaultLimit),
		pos:      0,
	}
}

// LoadHistory loads the command history from the specified file.
// It opens the file with the given filename and reads its contents.
// Each line in the file represents a request and is added to a History instance.
// Newlines represented by "\n" within a request are preserved.
// Returns a pointer to a History instance or an error if the file cannot be opened.
func LoadHistory(fileName string) (*History, error) {
	fileHandler, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, HistoryFileRigths)
	if err != nil {
		return nil, fmt.Errorf("failed to open history file: %w", err)
	}

	defer fileHandler.Close()

	reader := bufio.NewReader(fileHandler)

	h := NewHistory(fileName)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		line = strings.ReplaceAll(line, "\\n", "\n")

		h.AddRequest(line)
	}

	return h, nil
}

// Close writes the recent requests to the file and closes the file. It returns an error if the operation fails.
func (h *History) Close() error {
	fileHandler, err := os.OpenFile(h.fileName, os.O_WRONLY|os.O_CREATE, HistoryFileRigths)
	if err != nil {
		return err
	}

	defer fileHandler.Close()

	writer := bufio.NewWriter(fileHandler)

	var pos int
	if len(h.requests) < h.limit {
		pos = 0
	} else {
		pos = len(h.requests) - h.limit
	}

	for _, request := range h.requests[pos:] {
		request = strings.TrimSpace(request)
		if request == "" {
			continue
		}

		request = strings.ReplaceAll(request, "\n", "\\n")

		_, err := writer.WriteString(request + "\n")
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

// AddRequest adds a new request to the history if it is not empty and not a duplicate of the last request.
func (h *History) AddRequest(request string) {
	if request == "" {
		return
	}

	if len(h.requests) > 0 {
		if h.requests[len(h.requests)-1] == request {
			return
		}
	}

	h.requests = append(h.requests, request)
	h.pos = len(h.requests)
}

// PrevRequest returns the previous request from the history. If at the beginning of the history, it returns an empty string.
func (h *History) PrevRequest() string {
	if h.pos <= 0 {
		return ""
	}

	h.pos--
	req := h.requests[h.pos]

	return req
}

// NextRequest returns the next request from the history. If at the end of the history, it returns an empty string.
func (h *History) NextRequest() string {
	if h.pos >= len(h.requests) {
		return ""
	}

	req := h.requests[h.pos]
	h.pos++

	return req
}

// ResetPosition resets the position in the history to the end, allowing traversal from the latest request again.
func (h *History) ResetPosition() {
	if len(h.requests) == 0 {
		return
	}

	h.pos = len(h.requests)
}

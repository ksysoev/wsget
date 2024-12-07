package repo

import (
	"bufio"
	"log"
	"os"
	"strings"
)

const (
	HistoryFileRigths = 0o644
)

type History struct {
	fileName string
	requests []string
	limit    int
	pos      int
}

// NewHistory creates a new History instance with the given file name and limit.
// The History instance stores a list of requests made by the user and loads them from the file if it exists.
// The limit parameter specifies the maximum number of requests to store in the history.
func NewHistory(fileName string, limit int) *History {
	h := &History{
		fileName: fileName,
		limit:    limit,
		requests: make([]string, 0),
		pos:      0,
	}

	_ = h.loadFromFile()

	return h
}

// loadFromFile reads the history file and loads the requests into the History struct.
// It returns an error if the file cannot be opened or read.
func (h *History) loadFromFile() error {
	fileHandler, err := os.OpenFile(h.fileName, os.O_RDONLY|os.O_CREATE, HistoryFileRigths)
	if err != nil {
		log.Println("Error opening history file:", err)
		return err
	}

	defer fileHandler.Close()

	reader := bufio.NewReader(fileHandler)

	h.requests = make([]string, 0)

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

		h.requests = append(h.requests, line)
	}

	h.pos = len(h.requests) - 1

	return nil
}

// SaveToFile saves the history to a file.
// It opens the file with the given filename and writes the history requests to it.
// If the file does not exist, it creates it.
// If the number of requests is greater than the limit, it writes only the last limit requests.
// It replaces newlines with the escape sequence "\\n".
// It returns an error if it fails to open the file or write to it.
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

// AddRequest adds a request to the history. If the request is empty or the same as the last request, it will not be added.
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

// PrevRequest returns the previous request in the history.
// If there are no more previous requests, it returns an empty string.
func (h *History) PrevRequest() string {
	if h.pos <= 0 {
		return ""
	}

	h.pos--
	req := h.requests[h.pos]

	return req
}

// NextRequest returns the next request in the history.
// If there are no more requests, it returns an empty string.
func (h *History) NextRequest() string {
	if h.pos >= len(h.requests)-1 {
		return ""
	}

	h.pos++
	req := h.requests[h.pos]

	return req
}

// ResetPosition resets the position of the history to the end.
// If the history is empty, it does nothing.
func (h *History) ResetPosition() {
	if len(h.requests) == 0 {
		return
	}

	h.pos = len(h.requests)
}

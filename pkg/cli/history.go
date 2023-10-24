package cli

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
	limit    uint
	pos      int
}

func NewHistory(fileName string, limit uint) *History {
	h := &History{
		fileName: fileName,
		limit:    limit,
		requests: make([]string, 0),
		pos:      0,
	}

	_ = h.loadFromFile()

	return h
}

func (h *History) loadFromFile() error {
	fileHandler, err := os.OpenFile(h.fileName, os.O_RDONLY|os.O_CREATE, HistoryFileRigths)
	if err != nil {
		log.Println("Error opening history file:", err)
		return err
	}

	defer fileHandler.Close()

	reader := bufio.NewReader(fileHandler)

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

func (h *History) SaveToFile() error {
	fileHandler, err := os.OpenFile(h.fileName, os.O_WRONLY|os.O_CREATE, HistoryFileRigths)
	if err != nil {
		return err
	}

	defer fileHandler.Close()

	writer := bufio.NewWriter(fileHandler)

	var pos int
	if uint(len(h.requests)) < h.limit {
		pos = 0
	} else {
		pos = len(h.requests) - int(h.limit)
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
	h.pos = len(h.requests) - 1
}

func (h *History) PrevRequst() string {
	if h.pos <= 0 {
		return ""
	}

	req := h.requests[h.pos]
	h.pos--

	return req
}

func (h *History) NextRequst() string {
	if h.pos >= len(h.requests)-1 {
		return ""
	}

	h.pos++
	req := h.requests[h.pos]

	return req
}

func (h *History) ResetPosition() {
	if len(h.requests) == 0 {
		return
	}

	h.pos = len(h.requests) - 1
}

package cli

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type History struct {
	file     string
	limit    uint
	requests []string
}

func NewHistory(file string, limit uint) *History {
	h := &History{
		file:     file,
		limit:    limit,
		requests: make([]string, 0),
	}

	h.loadFromFile()

	return h
}

func (h *History) loadFromFile() error {
	fileHandler, err := os.OpenFile(h.file, os.O_RDONLY|os.O_CREATE, 0644)
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

	return nil
}

func (h *History) SaveToFile() error {
	fileHandler, err := os.OpenFile(h.file, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Println("Error opening history file:", err)
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
}

func (h *History) GetRequst(pos int) string {
	if pos <= 0 {
		return ""
	}

	if pos > len(h.requests) {
		return ""
	}

	return h.requests[len(h.requests)-pos]
}

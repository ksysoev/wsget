package cli

type History struct {
	requests []string
}

func NewHistory() *History {
	return &History{
		requests: make([]string, 0),
	}
}

func (h *History) AddRequest(request string) {
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

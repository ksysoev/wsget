package repo

import (
	"os"
	"strings"
	"testing"
)

func TestHistory(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "test_history")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())

	h := NewHistory(tmpfile.Name(), 5)

	req := []string{
		"request1",
		"request2",
		"request3",
		"request4",
		"request5",
	}

	// Test AddRequest method
	for i := 0; i < len(req); i++ {
		h.AddRequest(req[i])
	}

	if len(h.requests) != 5 {
		t.Errorf("AddRequest failed, expected %d requests, got %d", 5, len(h.requests))
	}

	// Test PrevRequest method
	for i := len(req) - 1; i >= 0; i-- {
		if h.PrevRequest() != req[i] {
			t.Errorf("PrevRequest failed, expected %s, got %s", req[i], h.PrevRequest())
		}
	}

	if h.PrevRequest() != "" {
		t.Errorf("PrevRequest failed, expected %s, got %s", "", h.PrevRequest())
	}

	// Test NextRequest method
	for i := 1; i < len(req); i++ {
		if h.NextRequest() != req[i] {
			t.Errorf("NextRequest failed, expected %s, got %s", req[i], h.NextRequest())
		}
	}

	if h.NextRequest() != "" {
		t.Errorf("NextRequest failed, expected %s, got %s", "", h.NextRequest())
	}

	// Test ResetPosition method
	h.ResetPosition()

	if h.pos != 5 {
		t.Errorf("ResetPosition failed, expected %d, got %d", 4, h.pos)
	}

	// Test SaveToFile method
	if err = h.SaveToFile(); err != nil {
		t.Errorf("SaveToFile failed, expected to get no error, but got %s", err)
	}

	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	expectedData := strings.Join(req, "\n") + "\n"
	if string(data) != expectedData {
		t.Errorf("SaveToFile failed, expected %s, got %s", expectedData, string(data))
	}

	// Test loadFromFile method
	h2 := NewHistory(tmpfile.Name(), 5)

	if err = h2.loadFromFile(); err != nil {
		t.Errorf("loadFromFile failed, expected to get no error, but got %s", err)
	}

	if len(h2.requests) != 5 {
		t.Errorf("loadFromFile failed, expected %d requests, got %d", 5, len(h2.requests))
	}

	for i := 0; i < len(req); i++ {
		if h2.requests[i] != req[i] {
			t.Errorf("loadFromFile failed, expected %s, got %s", req[i], h2.requests[i])
		}
	}
}

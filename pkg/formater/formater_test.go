package formater

import (
	"testing"

	"github.com/ksysoev/wsget/pkg/core"
)

func TestFormat_FormatMessage(t *testing.T) {
	formater := NewFormat()

	// Test text message formatting
	textMsg := core.Message{
		Type: core.Request,
		Data: "TestFormat_FormatMessage",
	}

	formattedTextMsg, err := formater.FormatMessage(textMsg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedTextMsg := "TestFormat_FormatMessage"
	if formattedTextMsg != expectedTextMsg {
		t.Errorf("Unexpected formatted message: %v", formattedTextMsg)
	}

	// Test JSON message formatting
	jsonMsg := core.Message{
		Type: core.Response,
		Data: `{"status": 200, "body": "TestFormat_FormatMessage"}`,
	}

	formattedJSONMsg, err := formater.FormatMessage(jsonMsg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedJSONMsg := "{\n  \"body\": \"TestFormat_FormatMessage\",\n  \"status\": 200\n}"
	if formattedJSONMsg != expectedJSONMsg {
		t.Errorf("Unexpected formatted message: %v, wanted %v", formattedJSONMsg, expectedJSONMsg)
	}

	testString := `{"status": 200, "body": "TestFormat_FormatMessage"`
	// Test invalid JSON message formatting
	invalidJSONMsg := core.Message{
		Type: core.Response,
		Data: testString,
	}

	formattedInvalidJSONMsg, err := formater.FormatMessage(invalidJSONMsg)
	if err != nil {
		t.Errorf("Expected to get no error, but got %v", err)
	}

	if formattedInvalidJSONMsg != testString {
		t.Errorf("Expected formated plain string, but got %v", formattedInvalidJSONMsg)
	}

	// Test unknown message type
	unknownMsg := core.Message{
		Type: core.MessageType(0),
		Data: "unknown message type",
	}

	formattedUnknownMsg, err := formater.FormatMessage(unknownMsg)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	if formattedUnknownMsg != "" {
		t.Errorf("Expected empty string, but got %v", formattedUnknownMsg)
	}
}

func TestFormat_FormatForFile(t *testing.T) {
	formater := NewFormat()

	// Test text message formatting for file
	textMsg := core.Message{
		Type: core.Request,
		Data: "TestFormat_FormatForFile",
	}

	formattedTextMsg, err := formater.FormatForFile(textMsg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedTextMsg := "TestFormat_FormatForFile"
	if formattedTextMsg != expectedTextMsg {
		t.Errorf("Unexpected formatted message: %v", formattedTextMsg)
	}

	// Test JSON message formatting for file
	jsonMsg := core.Message{
		Type: core.Response,
		Data: `{"status": 200, "body": "TestFormat_FormatForFile"}`,
	}

	formattedJSONMsg, err := formater.FormatForFile(jsonMsg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedJSONMsg := `{"body":"TestFormat_FormatForFile","status":200}`
	if formattedJSONMsg != expectedJSONMsg {
		t.Errorf("Unexpected formatted message: %v", formattedJSONMsg)
	}

	// Test invalid JSON message formatting for file
	invalidJSONMsg := core.Message{
		Type: core.Response,
		Data: `{"status": 200, "body": "TestFormat_FormatForFile"`,
	}

	formattedInvalidJSONMsg, err := formater.FormatForFile(invalidJSONMsg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedInvalidJSONMsg := "{\"status\": 200, \"body\": \"TestFormat_FormatForFile\""
	if formattedInvalidJSONMsg != expectedInvalidJSONMsg {
		t.Errorf("Unexpected formatted message: %v", formattedInvalidJSONMsg)
	}
}

func TestFormat_formatTextMessage(t *testing.T) {
	formater := NewFormat()

	// Test request message formatting
	requestMsg := core.Request
	requestData := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"

	formattedRequestMsg, err := formater.formatTextMessage(requestMsg, requestData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedRequestMsg := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"
	if formattedRequestMsg != expectedRequestMsg {
		t.Errorf("Unexpected formatted message: %v", formattedRequestMsg)
	}

	// Test response message formatting
	responseMsg := core.Response
	responseData := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello, world!"

	formattedResponseMsg, err := formater.formatTextMessage(responseMsg, responseData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedResponseMsg := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello, world!"
	if formattedResponseMsg != expectedResponseMsg {
		t.Errorf("Unexpected formatted message: %v", formattedResponseMsg)
	}

	// Test unknown message type
	unknownMsg := core.MessageType(0)
	unknownData := "Hello, world!"

	formattedUnknownMsg, err := formater.formatTextMessage(unknownMsg, unknownData)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	if formattedUnknownMsg != "" {
		t.Errorf("Expected empty string, but got %v", formattedUnknownMsg)
	}
}

func TestFormat_formatJSONMessage(t *testing.T) {
	formater := NewFormat()

	// Test request message formatting as JSON
	requestMsg := core.Request
	requestData := map[string]interface{}{
		"status": "200",
		"body":   "Hello, world!",
	}

	formattedRequestMsg, err := formater.formatJSONMessage(requestMsg, requestData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedRequestMsg := "{\n  \"body\": \"Hello, world!\",\n  \"status\": \"200\"\n}"

	if formattedRequestMsg != expectedRequestMsg {
		t.Errorf("Unexpected formatted message: %v", formattedRequestMsg)
	}

	// Test response message formatting as JSON
	responseMsg := core.Response
	responseData := map[string]interface{}{
		"status": "200",
		"body":   "Hello, world!",
	}

	formattedResponseMsg, err := formater.formatJSONMessage(responseMsg, responseData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedResponseMsg := "{\n  \"body\": \"Hello, world!\",\n  \"status\": \"200\"\n}"
	if formattedResponseMsg != expectedResponseMsg {
		t.Errorf("Unexpected formatted message: %v", formattedResponseMsg)
	}

	// Test unknown message type
	unknownMsg := core.MessageType(0)
	unknownData := "Hello, world!"

	formattedUnknownMsg, err := formater.formatJSONMessage(unknownMsg, unknownData)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	if formattedUnknownMsg != "" {
		t.Errorf("Expected empty string, but got %v", formattedUnknownMsg)
	}
}

func TestFormat_parseJSON(t *testing.T) {
	formater := NewFormat()

	// Test valid JSON parsing
	validJSON := `{"status": 200, "body": "Hello, world!"}`

	parsedValidJSON, ok := formater.parseJSON(validJSON)
	if !ok {
		t.Errorf("Expected true, but got false")
	}

	expectedValidJSON := map[string]interface{}{
		"status": 200.0,
		"body":   "Hello, world!",
	}

	if parsedValidJSON.(map[string]interface{})["status"].(float64) != expectedValidJSON["status"] {
		t.Errorf("Unexpected parsed JSON: %v", parsedValidJSON)
	}

	if parsedValidJSON.(map[string]interface{})["body"].(string) != expectedValidJSON["body"] {
		t.Errorf("Unexpected parsed JSON: %v", parsedValidJSON)
	}

	// Test invalid JSON parsing
	invalidJSON := `{"status": 200, "body": "Hello, world!"`

	parsedInvalidJSON, ok := formater.parseJSON(invalidJSON)
	if ok {
		t.Errorf("Expected false, but got true")
	}

	if parsedInvalidJSON != nil {
		t.Errorf("Expected nil, but got %v", parsedInvalidJSON)
	}
}

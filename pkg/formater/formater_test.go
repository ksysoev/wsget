package formater

import (
	"testing"

	"github.com/ksysoev/wsget/pkg/ws"
)

func TestFormater_FormatMessage(t *testing.T) {
	formater := NewFormatter()

	// Test text message formatting
	textMsg := ws.Message{
		Type: ws.Request,
		Data: "TestFormater_FormatMessage",
	}

	formattedTextMsg, err := formater.FormatMessage(textMsg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedTextMsg := "TestFormater_FormatMessage"
	if formattedTextMsg != expectedTextMsg {
		t.Errorf("Unexpected formatted message: %v", formattedTextMsg)
	}

	// Test JSON message formatting
	jsonMsg := ws.Message{
		Type: ws.Response,
		Data: `{"status": 200, "body": "TestFormater_FormatMessage"}`,
	}

	formattedJSONMsg, err := formater.FormatMessage(jsonMsg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedJSONMsg := "{\n  \"body\": \"TestFormater_FormatMessage\",\n  \"status\": 200\n}"
	if formattedJSONMsg != expectedJSONMsg {
		t.Errorf("Unexpected formatted message: %v, wanted %v", formattedJSONMsg, expectedJSONMsg)
	}

	testString := `{"status": 200, "body": "TestFormater_FormatMessage"`
	// Test invalid JSON message formatting
	invalidJSONMsg := ws.Message{
		Type: ws.Response,
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
	unknownMsg := ws.Message{
		Type: ws.MessageType(0),
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

func TestFormater_FormatForFile(t *testing.T) {
	formater := NewFormatter()

	// Test text message formatting for file
	textMsg := ws.Message{
		Type: ws.Request,
		Data: "TestFormater_FormatForFile",
	}

	formattedTextMsg, err := formater.FormatForFile(textMsg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedTextMsg := "TestFormater_FormatForFile"
	if formattedTextMsg != expectedTextMsg {
		t.Errorf("Unexpected formatted message: %v", formattedTextMsg)
	}

	// Test JSON message formatting for file
	jsonMsg := ws.Message{
		Type: ws.Response,
		Data: `{"status": 200, "body": "TestFormater_FormatForFile"}`,
	}

	formattedJSONMsg, err := formater.FormatForFile(jsonMsg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedJSONMsg := `{"body":"TestFormater_FormatForFile","status":200}`
	if formattedJSONMsg != expectedJSONMsg {
		t.Errorf("Unexpected formatted message: %v", formattedJSONMsg)
	}

	// Test invalid JSON message formatting for file
	invalidJSONMsg := ws.Message{
		Type: ws.Response,
		Data: `{"status": 200, "body": "TestFormater_FormatForFile"`,
	}

	formattedInvalidJSONMsg, err := formater.FormatForFile(invalidJSONMsg)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expectedInvalidJSONMsg := "{\"status\": 200, \"body\": \"TestFormater_FormatForFile\""
	if formattedInvalidJSONMsg != expectedInvalidJSONMsg {
		t.Errorf("Unexpected formatted message: %v", formattedInvalidJSONMsg)
	}
}

func TestFormater_formatTextMessage(t *testing.T) {
	formater := NewFormatter()

	// Test request message formatting
	requestMsg := ws.Request
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
	responseMsg := ws.Response
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
	unknownMsg := ws.MessageType(0)
	unknownData := "Hello, world!"

	formattedUnknownMsg, err := formater.formatTextMessage(unknownMsg, unknownData)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	if formattedUnknownMsg != "" {
		t.Errorf("Expected empty string, but got %v", formattedUnknownMsg)
	}
}

func TestFormater_formatJSONMessage(t *testing.T) {
	formater := NewFormatter()

	// Test request message formatting as JSON
	requestMsg := ws.Request
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
	responseMsg := ws.Response
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
	unknownMsg := ws.MessageType(0)
	unknownData := "Hello, world!"

	formattedUnknownMsg, err := formater.formatJSONMessage(unknownMsg, unknownData)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	if formattedUnknownMsg != "" {
		t.Errorf("Expected empty string, but got %v", formattedUnknownMsg)
	}
}

func TestFormater_parseJSON(t *testing.T) {
	formater := NewFormatter()

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

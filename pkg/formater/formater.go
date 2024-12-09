package formater

import (
	"encoding/json"
	"fmt"
)

// Format is a struct that contains two formatters, one for text and one for JSON.
type Format struct {
	text *TextFormat
	json *JSONFormat
}

// NewFormat creates a new instance of Format struct.
func NewFormat() *Format {
	return &Format{
		text: NewTextFormat(),
		json: NewJSONFormat(),
	}
}

// FormatMessage formats the given WebSocket message based on its type and data.
// If the data is a valid JSON, it will be formatted using the JSON formatter.
// Otherwise, it will be formatted using the text formatter.
func (f *Format) FormatMessage(msgType string, msgData string) (string, error) {
	obj, ok := f.parseJSON(msgData)

	if !ok {
		return f.formatTextMessage(msgType, msgData)
	}

	return f.formatJSONMessage(msgType, obj)
}

// FormatForFile formats the given WebSocket message for a file.
// It first tries to parse the message data as JSON, and if successful, formats it as JSON.
// If parsing fails, it formats the message data as plain text.
func (f *Format) FormatForFile(msgType string, msgData string) (string, error) {
	obj, ok := f.parseJSON(msgData)

	if !ok {
		return f.text.FormatForFile(msgData)
	}

	return f.json.FormatForFile(obj)
}

// formatTextMessage formats the given WebSocket message data as text based on its type.
func (f *Format) formatTextMessage(msgType string, data string) (string, error) {
	switch msgType {
	case "Request":
		return f.text.FormatRequest(data)
	case "Response":
		return f.text.FormatResponse(data)
	case "NotDefined":
		return "", fmt.Errorf("unknown message type")
	default:
		panic("Unexpected message type: " + string(msgType))
	}
}

// formatJSONMessage formats the given WebSocket message data as JSON based on its type.
func (f *Format) formatJSONMessage(msgType string, data any) (string, error) {
	switch msgType {
	case "Request":
		return f.json.FormatRequest(data)
	case "Response":
		return f.json.FormatResponse(data)
	case "NotDefined":
		return "", fmt.Errorf("unknown message type")
	default:
		panic("Unexpected message type: " + string(msgType))
	}
}

// parseJSON parses the given string as JSON and returns the parsed object.
// If the string is not a valid JSON, it returns false as the second value.
func (f *Format) parseJSON(data string) (any, bool) {
	var obj any
	err := json.Unmarshal([]byte(data), &obj)

	if err != nil {
		return obj, false
	}

	return obj, true
}

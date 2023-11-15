package formater

import (
	"encoding/json"
	"fmt"

	"github.com/ksysoev/wsget/pkg/ws"
)

type Formater interface {
	FormatMessage(wsMsg ws.Message) (string, error)
	FormatForFile(wsMsg ws.Message) (string, error)
}

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
func (f *Format) FormatMessage(wsMsg ws.Message) (string, error) {
	wsMsgData := wsMsg.Data

	obj, ok := f.parseJSON(wsMsgData)

	if !ok {
		return f.formatTextMessage(wsMsg.Type, wsMsgData)
	}

	return f.formatJSONMessage(wsMsg.Type, obj)
}

// FormatForFile formats the given WebSocket message for a file.
// It first tries to parse the message data as JSON, and if successful, formats it as JSON.
// If parsing fails, it formats the message data as plain text.
func (f *Format) FormatForFile(wsMsg ws.Message) (string, error) {
	wsMsgData := wsMsg.Data

	obj, ok := f.parseJSON(wsMsgData)

	if !ok {
		return f.text.FormatForFile(wsMsgData)
	}

	return f.json.FormatForFile(obj)
}

// formatTextMessage formats the given WebSocket message data as text based on its type.
func (f *Format) formatTextMessage(msgType ws.MessageType, data string) (string, error) {
	switch msgType {
	case ws.Request:
		return f.text.FormatRequest(data)
	case ws.Response:
		return f.text.FormatResponse(data)
	case ws.NotDefined:
		return "", fmt.Errorf("unknown message type")
	default:
		panic("Unexpected message type: " + string(msgType))
	}
}

// formatJSONMessage formats the given WebSocket message data as JSON based on its type.
func (f *Format) formatJSONMessage(msgType ws.MessageType, data any) (string, error) {
	switch msgType {
	case ws.Request:
		return f.json.FormatRequest(data)
	case ws.Response:
		return f.json.FormatResponse(data)
	case ws.NotDefined:
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

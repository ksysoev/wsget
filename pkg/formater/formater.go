package formater

import (
	"encoding/json"
	"fmt"

	"github.com/ksysoev/wsget/pkg/ws"
)

// Formater is a struct that contains two formatters, one for text and one for JSON.
type Formater struct {
	text *TextFormater
	json *JsonFormater
}

// NewFormatter creates a new instance of Formater struct.
func NewFormatter() *Formater {
	return &Formater{
		text: NewTextFormater(),
		json: NewJsonFormater(),
	}
}

// FormatMessage formats the given WebSocket message based on its type and data.
// If the data is a valid JSON, it will be formatted using the JSON formatter.
// Otherwise, it will be formatted using the text formatter.
func (f *Formater) FormatMessage(wsMsg ws.Message) (string, error) {
	wsMsgData := wsMsg.Data

	obj, ok := f.parseJson(wsMsgData)

	if !ok {
		return f.formatTestMessage(wsMsg.Type, wsMsgData)
	}

	return f.formatJsonMessage(wsMsg.Type, obj)
}

// FormatForFile formats the given WebSocket message for a file.
// It first tries to parse the message data as JSON, and if successful, formats it as JSON.
// If parsing fails, it formats the message data as plain text.
func (f *Formater) FormatForFile(wsMsg ws.Message) (string, error) {
	wsMsgData := wsMsg.Data

	obj, ok := f.parseJson(wsMsgData)

	if !ok {
		return f.text.FormatForFile(wsMsgData)
	}

	return f.json.FormatForFile(obj)
}

// formatTestMessage formats the given WebSocket message data as text based on its type.
func (f *Formater) formatTestMessage(msgType ws.MessageType, data string) (string, error) {
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

// formatJsonMessage formats the given WebSocket message data as JSON based on its type.
func (f *Formater) formatJsonMessage(msgType ws.MessageType, data any) (string, error) {
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

// parseJson parses the given string as JSON and returns the parsed object.
// If the string is not a valid JSON, it returns false as the second value.
func (f *Formater) parseJson(data string) (any, bool) {
	var obj any
	err := json.Unmarshal([]byte(data), &obj)

	if err != nil {
		return obj, false
	}

	return obj, true
}

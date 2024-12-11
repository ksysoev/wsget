package formater

import (
	"github.com/fatih/color"
)

// TextFormat is a struct that holds the color for request and response
type TextFormat struct {
	request  *color.Color
	response *color.Color
}

// NewTextFormat creates a new instance of TextFormat
func NewTextFormat() *TextFormat {
	return &TextFormat{
		request:  color.New(color.FgMagenta),
		response: color.New(color.FgCyan),
	}
}

// FormatRequest formats the request data and returns it as a string
func (tf *TextFormat) FormatRequest(data string) (string, error) {
	output := tf.request.Sprintf("%s", data)
	return output, nil
}

// FormatResponse formats the response data and returns it as a string
func (tf *TextFormat) FormatResponse(data string) (string, error) {
	output := tf.response.Sprintf("%s", data)

	return output, nil
}

// FormatForFile formats the data for file and returns it as a string
func (tf *TextFormat) FormatForFile(data string) (string, error) {
	return data, nil
}

package formater

import (
	"github.com/fatih/color"
)

// TextFormater is a struct that holds the color for request and response
type TextFormater struct {
	request  *color.Color
	response *color.Color
}

// NewTextFormater creates a new instance of TextFormater
func NewTextFormater() *TextFormater {
	return &TextFormater{
		request:  color.New(color.FgMagenta),
		response: color.New(color.FgCyan),
	}
}

// FormatRequest formats the request data and returns it as a string
func (tf *TextFormater) FormatRequest(data string) (string, error) {
	output := tf.request.Sprintf("%s", data)
	return output, nil
}

// FormatResponse formats the response data and returns it as a string
func (tf *TextFormater) FormatResponse(data string) (string, error) {
	output := tf.response.Sprintf("%s", data)

	return output, nil
}

// FormatForFile formats the data for file and returns it as a string
func (tf *TextFormater) FormatForFile(data string) (string, error) {
	return data, nil
}

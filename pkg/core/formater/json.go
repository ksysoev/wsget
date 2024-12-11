package formater

import (
	"encoding/json"

	"github.com/TylerBrock/colorjson"
	"github.com/fatih/color"
)

// JSONFormat is a struct that contains two colorjson formatters for request and response.
type JSONFormat struct {
	request  *colorjson.Formatter
	response *colorjson.Formatter
}

// NewJSONFormat creates a new instance of JSONFormat and returns a pointer to it.
func NewJSONFormat() *JSONFormat {
	request := colorjson.NewFormatter()
	request.Indent = 2
	request.KeyColor = color.New(color.FgMagenta)
	request.StringColor = color.New(color.FgYellow)
	request.BoolColor = color.New(color.FgBlue)
	request.NumberColor = color.New(color.FgGreen)
	request.NullColor = color.New(color.FgRed)

	response := colorjson.NewFormatter()
	response.Indent = 2
	response.KeyColor = color.New(color.FgCyan)
	response.StringColor = color.New(color.FgYellow)
	response.BoolColor = color.New(color.FgBlue)
	response.NumberColor = color.New(color.FgGreen)
	response.NullColor = color.New(color.FgRed)

	return &JSONFormat{
		request:  request,
		response: response,
	}
}

// FormatRequest formats the given data as a JSON string using the request formatter.
func (jf *JSONFormat) FormatRequest(data any) (string, error) {
	output, err := jf.request.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// FormatResponse formats the given data as a JSON string using the response formatter.
func (jf *JSONFormat) FormatResponse(data any) (string, error) {
	output, err := jf.response.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// FormatForFile formats the given data as a JSON string using the default json package.
func (jf *JSONFormat) FormatForFile(data any) (string, error) {
	output, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

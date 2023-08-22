package formater

import (
	"encoding/json"

	"github.com/TylerBrock/colorjson"
	"github.com/fatih/color"
)

// JsonFormater is a struct that contains two colorjson formatters for request and response.
type JsonFormater struct {
	request  *colorjson.Formatter
	response *colorjson.Formatter
}

// NewJsonFormater creates a new instance of JsonFormater and returns a pointer to it.
func NewJsonFormater() *JsonFormater {
	request := colorjson.NewFormatter()
	request.Indent = 2
	request.KeyColor = color.New(color.FgGreen, color.Bold)
	request.StringColor = color.New(color.FgHiGreen)
	request.BoolColor = color.New(color.FgHiMagenta)
	request.NumberColor = color.New(color.FgHiCyan)
	request.NullColor = color.New(color.FgHiRed)

	response := colorjson.NewFormatter()
	response.Indent = 2
	response.KeyColor = color.New(color.FgHiRed, color.Bold)
	response.StringColor = color.New(color.FgHiYellow)
	response.BoolColor = color.New(color.FgHiMagenta)
	response.NumberColor = color.New(color.FgHiCyan)
	response.NullColor = color.New(color.FgHiRed)

	return &JsonFormater{
		request:  request,
		response: response,
	}
}

// FormatRequest formats the given data as a JSON string using the request formatter.
func (jf *JsonFormater) FormatRequest(data any) (string, error) {
	output, err := jf.request.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// FormatResponse formats the given data as a JSON string using the response formatter.
func (jf *JsonFormater) FormatResponse(data any) (string, error) {
	output, err := jf.response.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// FormatForFile formats the given data as a JSON string using the default json package.
func (jf *JsonFormater) FormatForFile(data any) (string, error) {
	output, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

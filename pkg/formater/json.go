package formater

import (
	"encoding/json"

	"github.com/TylerBrock/colorjson"
	"github.com/fatih/color"
)

type JsonFormater struct {
	request  *colorjson.Formatter
	response *colorjson.Formatter
}

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

func (jf *JsonFormater) FormatRequest(data any) (string, error) {
	output, err := jf.request.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (jf *JsonFormater) FormatResponse(data any) (string, error) {
	output, err := jf.response.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (jf *JsonFormater) FormatForFile(data any) (string, error) {
	output, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

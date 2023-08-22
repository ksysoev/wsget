package formater

import (
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestTextFormater_FormatRequest(t *testing.T) {
	tf := NewTextFormater()
	data := "test request data"
	expectedOutput := color.New(color.FgGreen).Sprintf("test request data")

	output, err := tf.FormatRequest(data)

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
}

func TestTextFormater_FormatResponse(t *testing.T) {
	tf := NewTextFormater()
	data := "test response data"
	expectedOutput := color.New(color.FgHiRed).Sprintf("test response data")

	output, err := tf.FormatResponse(data)

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
}

func TestTextFormater_FormatForFile(t *testing.T) {
	tf := NewTextFormater()
	data := "test data"

	expectedOutput := "test data"

	output, err := tf.FormatForFile(data)

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
}

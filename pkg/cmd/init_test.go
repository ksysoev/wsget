package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWsgetInitCommands(t *testing.T) {
	version := "test-version"

	cmd := InitCommands(version)

	assert.NotNil(t, cmd)
	assert.Equal(t, "wsget url [flags]", cmd.Use)
	assert.Equal(t, "A command-line tool for interacting with WebSocket servers", cmd.Short)
	assert.Equal(t, longDescription, cmd.Long)
	assert.Equal(t, version, cmd.Version)
	assert.Equal(t, `wsget wss://ws.postman-echo.com/raw -r "Hello, world!"`, cmd.Example)

	insecureFlag := cmd.Flags().Lookup("insecure")
	assert.NotNil(t, insecureFlag)
	assert.Equal(t, "false", insecureFlag.DefValue)

	requestFlag := cmd.Flags().Lookup("request")
	assert.NotNil(t, requestFlag)
	assert.Equal(t, "", requestFlag.DefValue)

	outputFileFlag := cmd.Flags().Lookup("output")
	assert.NotNil(t, outputFileFlag)
	assert.Equal(t, "", outputFileFlag.DefValue)

	waitResponseFlag := cmd.Flags().Lookup("wait-resp")
	assert.NotNil(t, waitResponseFlag)
	assert.Equal(t, "-1", waitResponseFlag.DefValue)

	headersFlag := cmd.Flags().Lookup("header")
	assert.NotNil(t, headersFlag)
	assert.Equal(t, "[]", headersFlag.DefValue)

	inputFileFlag := cmd.Flags().Lookup("input")
	assert.NotNil(t, inputFileFlag)
	assert.Equal(t, "", inputFileFlag.DefValue)

	verboseFlag := cmd.Flags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue)
}

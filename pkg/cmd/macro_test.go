package cmd

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRunMacroDownloadCommand_NoUrl(t *testing.T) {
	args := &flags{}
	name := "test"
	unnamedArgs := []string{""}

	err := runMacroDownloadCommand(context.Background(), args, &name, unnamedArgs)
	assert.ErrorContains(t, err, "macro URL is required")
}

func TestRunMacroDownloadCommand_NoConfigDir(t *testing.T) {
	args := &flags{}
	name := "test"
	unnamedArgs := []string{"http://localhost:0"}

	err := runMacroDownloadCommand(context.Background(), args, &name, unnamedArgs)
	assert.ErrorContains(t, err, "connect: can't assign requested address")
}

func TestRunMacroDownloadCommand_FailToDownload(t *testing.T) {
	args := &flags{configDir: "testdata"}
	name := "test"
	unnamedArgs := []string{"http://localhost:0"}

	err := runMacroDownloadCommand(context.Background(), args, &name, unnamedArgs)
	assert.ErrorContains(t, err, "connect: can't assign requested address")
}

func TestRunMacroDownloadCommand(t *testing.T) {
	// Act
	runner := createMacroDownloadRunner(&flags{}, nil)
	err := runner(&cobra.Command{}, []string{""})

	// Assert
	assert.ErrorContains(t, err, "macro URL is required")
}

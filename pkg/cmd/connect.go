package cmd

import "github.com/spf13/cobra"

func runConnectCmd(arg *flags) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}

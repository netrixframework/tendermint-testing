package cmd

import "github.com/spf13/cobra"

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "tendermint-testing",
	}
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.AddCommand(PCTStrategy())
	cmd.AddCommand(PCTTestStrategy())
	return cmd
}

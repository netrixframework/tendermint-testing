package cmd

import (
	"github.com/spf13/cobra"
)

func strategyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "strat",
	}
	cmd.AddCommand(timeoutCmd)
	cmd.AddCommand(pctStrategy)
	cmd.AddCommand(pctTestStrategy)
	cmd.AddCommand(testStrat)
	return cmd
}

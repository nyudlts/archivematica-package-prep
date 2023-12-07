package cmd

import (
	"github.com/spf13/cobra"
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

var rootCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {

	},
}

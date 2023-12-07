package cmd

import "github.com/spf13/cobra"

var (
	aipLocation string
	tmpLocation string
)

func init() {
	singleCmd.Flags().StringVar(&aipLocation, "aip-location", "", "")
	singleCmd.Flags().StringVar(&tmpLocation, "tmp-location", "", "")
	rootCmd.AddCommand(singleCmd)
}

var singleCmd = &cobra.Command{
	Use: "single",
	Run: func(cmd *cobra.Command, args []string) {
		if tmpLocation == "" {
			tmpLocation = "/tmp"
		}

		if err := processAIP(aipLocation, tmpLocation); err != nil {
			panic(err)
		}
	},
}

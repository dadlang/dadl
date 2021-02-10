package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Dadl",
	Long:  `All software has versions. This is Dadl's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Dadl 0.0.1")
	},
}

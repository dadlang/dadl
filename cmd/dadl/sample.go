package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sampleCmd)
}

var sampleCmd = &cobra.Command{
	Use:   "sample",
	Short: "Generates sample data for given schema file",
	Long:  `Generates sample data for given schema file`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO sample")
	},
}

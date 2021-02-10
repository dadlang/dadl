package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(describeCmd)
}

var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describes given dadl schema file",
	Long:  `Describes given dadl schema file`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO describe")
	},
}

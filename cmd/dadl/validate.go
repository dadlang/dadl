package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(validateCmd)
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validates given dadl file",
	Long:  `Validates a given dadl file`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO validate")
	},
}

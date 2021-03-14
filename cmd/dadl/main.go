package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dadl",
	Short: "Dadl is a configuration language and a utility tool",
	Long: `A simple but extendable configuration language that allows to describe
complex stuctures using custom DSL and preserve readability at the same time.
Complete documentation is available at http://github.com/dadlang/dadl`,
}

func execute() {
}

func main() {
	log.SetFlags(0)
	log.SetOutput(ioutil.Discard)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

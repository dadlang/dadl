package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dadlang/dadl/pkg/export"
	"github.com/dadlang/dadl/pkg/parser"
	"github.com/spf13/cobra"
)

var (
	format string

	formatChoices = map[string]func(map[string]interface{}) string{"json": export.ToJSON, "yaml": export.ToYAML}
)

func init() {
	exportCmd.Flags().StringVarP(&format, "format", "f", "json", "Format of the exported file {json|yaml}")
	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Exports configuration to given format",
	Long:  `Exports configuration to given format.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("requires a file name")
		}
		if _, ok := formatChoices[format]; !ok {
			return fmt.Errorf("invalid exported file format specified: %s", format)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		exportHandler(args[0])
	},
}

func exportHandler(filePath string) {

	exporter, _ := formatChoices[format]

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	schema, _ := parser.ParseSchema()
	p := parser.NewParser(schema)
	tree, err := p.Parse(file, parser.NewFSResourceProvider(filepath.Dir(filePath)))

	if err != nil {
		println("Err: ", err.Error())
		return
	}

	fmt.Print(exporter(tree))
}

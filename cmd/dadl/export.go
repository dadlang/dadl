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
	format  string
	outFile string

	formatChoices = map[string]func(map[string]interface{}) string{"json": export.ToJSON, "yaml": export.ToYAML}
)

func init() {
	exportCmd.Flags().StringVarP(&format, "format", "f", "json", "Format of the exported file {json|yaml}")
	exportCmd.Flags().StringVarP(&outFile, "out", "o", "", "Save exported data to a file")
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

	p := parser.NewParser()
	tree, err := p.Parse2(file, parser.NewFSResourceProvider(filepath.Dir(filePath)))

	if err != nil {
		println("Err: ", err.Error())
		return
	}

	if outFile != "" {
		saveToFile(outFile, exporter(tree))
	} else {
		fmt.Print(exporter(tree))
	}
}

func saveToFile(file string, data string) {
	f, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString(data)
	if err != nil {
		log.Fatal(err)
	}
}

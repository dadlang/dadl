package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dadlang/dadl/pkg/parser"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
)

func init() {
	rootCmd.AddCommand(printCmd)
}

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Prints data from given file",
	Long:  `Prints data from given file.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires optional path and a file name")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			printHandler(args[0], ".")
		} else {
			printHandler(args[1], args[0])
		}
	},
}

func printHandler(filePath string, treePath string) {

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	p := parser.NewParser()
	tree, err := p.Parse(file, parser.NewFSResourceProvider(filepath.Dir(filePath)))

	if err != nil {
		println(err.Error())
		return
	}

	result, _ := filterTree(tree, treePath)
	printTree(result, treePath)
}

func filterTree(root map[string]interface{}, filterPath string) (interface{}, error) {
	if filterPath == "." {
		return root, nil
	}
	var node interface{} = root
	pathElements := strings.Split(filterPath, ".")
	for _, pathElement := range pathElements {
		node = node.(map[string]interface{})[pathElement]
	}
	return node, nil
}

func printTree(root interface{}, rootName string) {
	var tree treeprint.Tree
	if asMap, ok := root.(map[string]interface{}); ok {
		tree = treeprint.NewWithRoot(rootName)
		buildMapChildren(tree, asMap)
	} else {
		tree = addChild(treeprint.New(), rootName, root)
	}
	fmt.Println(tree.String())
}

func buildMapChildren(tree treeprint.Tree, children map[string]interface{}) {
	keys := make([]string, 0)
	for k := range children {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		addChild(tree, key, children[key])
	}
}

func buildSliceChildren(tree treeprint.Tree, children []interface{}) {
	for key, value := range children {
		addChild(tree, fmt.Sprintf("[%v]", key), value)
	}
}

func addChild(tree treeprint.Tree, key string, value interface{}) treeprint.Tree {
	if asMap, ok := value.(map[string]interface{}); ok {
		branch := tree.AddBranch(key)
		buildMapChildren(branch, asMap)
		return branch
	} else if asSlice, ok := value.([]interface{}); ok {
		branch := tree.AddBranch(key)
		buildSliceChildren(branch, asSlice)
		return branch
	} else if asString, ok := value.(string); ok {
		if strings.Contains(asString, "\n") {
			return tree.AddBranch(key + ":\n" + asString)
		} else {
			return tree.AddBranch(key + ": " + asString)
		}
	} else {
		return tree.AddBranch(fmt.Sprintf("%s: %v", key, value))
	}
}

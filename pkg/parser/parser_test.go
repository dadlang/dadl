package parser

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParserE2E(t *testing.T) {
	parser := NewParser()
	for _, tc := range testCases {
		fullPath := "../../samples/" + tc.testFile
		file, err := os.Open(fullPath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		got, err := parser.Parse(file, NewFSResourceProvider(filepath.Dir(fullPath)))
		if err != nil {
			t.Fatalf("could not parse %v", tc.testFile)
		}
		if !reflect.DeepEqual(got, tc.expected) {
			t.Errorf("[%s] = %+v; want %+v", tc.testFile, got, tc.expected)
		}
	}
}

type testCase struct {
	name     string
	testFile string
	expected Node
}

var testCases = []testCase{
	{
		name:     "simple test",
		testFile: "simple/simple.dad",
		expected: Node{
			"someRoot": Node{
				"firstChild": "some long string value with spaces",
				"secondChild": Node{
					"nestedChild": 7,
				},
			},
		},
	},
	{
		name:     "embedded text test",
		testFile: "embedded_text/embedded_text.dad",
		expected: Node{
			"someJson": `{
	"martin": {
		"name": "Martin D'vloper",
		"job": "Developer"
	}
}`,
			"someYaml": `martin:
	name: Martin D'vloper
	job: Developer`,
			"someDadl": `[martin]
name Martin D'vloper
job Developer`,
			"someBrainfuck": `++++++++++[>+>+++>+++++++>+++++
+++++<<<<-]>>>++.>+.+++++++..++
+.<<++.>----.---.+++.++++++++.`,
		},
	},
}

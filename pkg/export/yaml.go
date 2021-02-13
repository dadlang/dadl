package export

import (
	"log"

	"gopkg.in/yaml.v2"
)

//ToYAML exports tree to YAML format
func ToYAML(tree map[string]interface{}) string {
	result, err := yaml.Marshal(tree)
	if err != nil {
		log.Fatal(err)
	}
	return string(result)
}

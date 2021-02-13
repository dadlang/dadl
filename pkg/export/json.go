package export

import (
	"encoding/json"
	"log"
)

//ToJSON exports tree to JSON format
func ToJSON(tree map[string]interface{}) string {
	result, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	return string(result)
}

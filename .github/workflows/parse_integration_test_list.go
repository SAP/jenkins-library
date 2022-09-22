package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"os"
)

func main() {
	fileName := flag.String("file-name", "../../integration/github_actions_integration_test_list.yml", "Tests to be executed")
	flag.Parse()
	f, _ := os.ReadFile(*fileName)
	var Config struct {
		Version interface{} `json:"version,omitempty" yaml:"version"`
	}
	yaml.Unmarshal(f, &Config)
	output, _ := json.Marshal(Config)
	fmt.Println(string(output))
}

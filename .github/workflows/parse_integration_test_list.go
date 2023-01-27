package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"os"
)

func main() {
	file := flag.String("file", "../../integration/github_actions_integration_test_list.yml", "Tests to be executed")
	flag.Parse()
	f, _ := os.ReadFile(*file)
	var Matrix struct {
		Run interface{} `json:"run,omitempty" yaml:"run"`
	}
	yaml.Unmarshal(f, &Matrix)
	output, _ := json.Marshal(Matrix)
	fmt.Println(string(output))
}

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
		Include interface{} `json:"include,omitempty" yaml:"include"`
	}
	yaml.Unmarshal(f, &Config)
	output, _ := json.Marshal(Config)
	fmt.Println(string(output)[1 : len(output)-1])
}

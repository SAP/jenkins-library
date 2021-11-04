package terraform

import (
	"encoding/json"
)

type TerraformOutput struct {
	Sensitive bool        `json:"sensitive"`
	ObjType   interface{} `json:"type"`
	Value     interface{} `json:"value"`
}

func ReadOutputs(tfOutputJson string) (map[string]interface{}, error) {
	var objmap map[string]TerraformOutput
	err := json.Unmarshal([]byte(tfOutputJson), &objmap)

	if err != nil {
		return nil, err
	}

	retmap := make(map[string]interface{})

	for tfoutvarname, tfoutvar := range objmap {
		retmap[tfoutvarname] = tfoutvar.Value
	}

	return retmap, nil
}

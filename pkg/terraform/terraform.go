package terraform

import (
	"encoding/json"
)

type TerraformOutput struct {
	Sensitive bool `json:"sensitive"`
	ObjType   any  `json:"type"`
	Value     any  `json:"value"`
}

func ReadOutputs(tfOutputJson string) (map[string]any, error) {
	var objmap map[string]TerraformOutput
	err := json.Unmarshal([]byte(tfOutputJson), &objmap)

	if err != nil {
		return nil, err
	}

	retmap := make(map[string]any)

	for tfoutvarname, tfoutvar := range objmap {
		retmap[tfoutvarname] = tfoutvar.Value
	}

	return retmap, nil
}

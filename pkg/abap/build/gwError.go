package build

import (
	"encoding/json"
	"fmt"
	"strings"
)

type GW_error struct {
	Error struct {
		Code    string
		Message struct {
			Lang  string
			Value string
		}
		Innererror struct {
			Application struct {
				Component_id      string
				Service_namespace string
				Service_id        string
				Service_version   string
			}
			Transactionid    string
			Timestamp        string
			Error_Resolution struct {
				SAP_Transaction string
				SAP_Note        string
			}
			Errordetails []struct {
				ContentID   string
				Code        string
				Message     string
				Propertyref string
				Severity    string
				Transition  bool
				Target      string
			}
		}
	}
}

func extractErrorStackFromJsonData(jsonData []byte) string {
	my_error := new(GW_error)
	if err := my_error.FromJson(jsonData); err != nil {
		return string(jsonData)
	}
	return my_error.ExtractStack()
}

func (my_error *GW_error) FromJson(inputJson []byte) error {
	if err := json.Unmarshal(inputJson, my_error); err != nil {
		return err
	}
	return nil
}

func (my_error *GW_error) ExtractStack() string {
	var stack strings.Builder
	var previousMessage string
	for index, detail := range my_error.Error.Innererror.Errordetails {
		if previousMessage == detail.Message {
			continue
		}
		previousMessage = detail.Message
		stack.WriteString(fmt.Sprintf("[%v] %s\n", index, detail.Message))
	}
	return stack.String()
}

package generator

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func readAndAdjustTemplate(docFile io.ReadCloser) string {
	//read template content
	content, err := ioutil.ReadAll(docFile)
	checkError(err)
	contentStr := string(content)

	//replace old placeholder with new ones
	contentStr = strings.ReplaceAll(contentStr, "# ${docGenStepName}", "{{StepName .}}")
	contentStr = strings.ReplaceAll(contentStr, "## ${docGenDescription}", "{{Description .}}")
	contentStr = strings.ReplaceAll(contentStr, "## ${docGenParameters}", "{{Parameters .}}")
	contentStr = strings.ReplaceAll(contentStr, "## ${docGenConfiguration}", "")
	contentStr = strings.ReplaceAll(contentStr, "## ${docJenkinsPluginDependencies}", "")

	return contentStr
}

func checkError(err error) {
	if err != nil {
		fmt.Printf("Error occurred: %v\n", err)
		os.Exit(1)
	}
}

func contains(v []string, s string) bool {
	for _, i := range v {
		if i == s {
			return true
		}
	}
	return false
}

func ifThenElse(condition bool, positive string, negative string) string {
	if condition {
		return positive
	}
	return negative
}

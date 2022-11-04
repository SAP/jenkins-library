package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteTemplate(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T){
		testData := struct{
			Test string
		}{
			Test: "TestValue",
		}
		res, err := ExecuteTemplate("{{.Test}}", testData)
		assert.NoError(t, err)
		assert.Equal(t, testData.Test, res)
	})

	t.Run("failure - parse template", func(t *testing.T){
		_, err := ExecuteTemplate("{{.Test", nil)
		assert.Contains(t, fmt.Sprint(err), "failed to parse template:")
	})

	t.Run("failure - parse template", func(t *testing.T){
		testData := struct{
			Test string
		}{
			Test: "TestValue",
		}
		_, err := ExecuteTemplate("{{range .Test}}{{.NotThere}}{{end}}", testData)
		assert.Contains(t, fmt.Sprint(err), "failed to execute template:")
	})
	
}
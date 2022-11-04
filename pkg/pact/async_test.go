package pact

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"
)

func TestValidateAsynch(t *testing.T) {
	t.Run("success", func(t *testing.T){
		fileUtils := mock.FilesMock{}
		pathToPactFolder := ""
		pathToAsyncFile := ""
		err := ValidateAsynch(pathToPactFolder, pathToAsyncFile, &fileUtils)
		assert.NoError(t, err)
	})

	t.Run("failure - tests failed", func(t *testing.T){
		
	})

	t.Run("failure - read and unmarshal file", func(t *testing.T){
		
	})

	t.Run("failure - generate async API comparison map", func(t *testing.T){
		
	})

	t.Run("failure - generate pact comparison map", func(t *testing.T){
		
	})

	t.Run("failure - contract validation", func(t *testing.T){
		
	})
}
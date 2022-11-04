package util

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"
)

func TestReadAndUnmarshalFile(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T){
		fileMock := mock.FilesMock{}
		testFile := "testPath/test.json"
		fileMock.AddFile(testFile, []byte(`{"testKey":"testVal"}`))
	
		var res map[string]string
	
		err := ReadAndUnmarshalFile(testFile, &res, &fileMock)
		assert.NoError(t, err)
	})

	t.Run("failure - read file", func(t *testing.T){
		fileMock := mock.FilesMock{}
		testFile := "testPath/test.json"
	
		var res map[string]string
	
		err := ReadAndUnmarshalFile(testFile, &res, &fileMock)
		assert.Contains(t, fmt.Sprint(err), "failed to read and open file 'testPath/test.json'")
	})

	t.Run("failure - parse file content", func(t *testing.T){
		fileMock := mock.FilesMock{}
		testFile := "testPath/test.json"
		fileMock.AddFile(testFile, []byte(`{"test"}`))
	
		var res map[string]string
	
		err := ReadAndUnmarshalFile(testFile, &res, &fileMock)
		assert.Contains(t, fmt.Sprint(err), "failed to parse json file 'testPath/test.json'")
	})

}
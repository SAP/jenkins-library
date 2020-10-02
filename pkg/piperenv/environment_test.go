package piperenv

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetResourceParameter(t *testing.T) {
	type args struct {
		path         string
		resourceName string
		paramName    string
		value        interface{}
	}
	tests := []struct {
		name string
		want string
		args args
	}{
		{name: "string", want: "testVal", args: args{path: "", resourceName: "testRes", paramName: "testParam", value: "testVal"}},
		{name: "boolean", want: "{\"value\": true}", args: args{path: "", resourceName: "testRes", paramName: "testParam", value: true}},
		{name: "integer", want: "{\"value\": 1}", args: args{path: "", resourceName: "testRes", paramName: "testParam", value: 1}},
		{name: "float", want: "{\"value\": 0.123}", args: args{path: "", resourceName: "testRes", paramName: "testParam", value: 0.123}},
		{name: "string list", want: "[\"test\",\"abc\"]", args: args{path: "", resourceName: "testRes", paramName: "testParam", value: []string{"test", "abc"}}},
		{name: "boolean list", want: "[true,false]", args: args{path: "", resourceName: "testRes", paramName: "testParam", value: []bool{true, false}}},
		{name: "integer list", want: "[1,2]", args: args{path: "", resourceName: "testRes", paramName: "testParam", value: []int{1, 2}}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// init
			dir, tempDirErr := ioutil.TempDir("", "")
			require.NoError(t, tempDirErr)
			require.DirExists(t, dir, "Failed to create temporary directory")
			// clean up tmp dir
			defer os.RemoveAll(dir)
			targetFile := filepath.Join(dir, testCase.args.resourceName, testCase.args.paramName)
			// test
			err := SetResourceParameter(dir, testCase.args.resourceName, testCase.args.paramName, testCase.args.value)
			// assert
			assert.NoError(t, err)
			assert.FileExists(t, targetFile)
			v, err := ioutil.ReadFile(targetFile)
			require.NoError(t, err)
			assert.Equal(t, testCase.want, string(v))
		})
	}
}

//TODO: add test cases for GetResourceParameter which was previously tested implicitly

func TestSetParameter(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}

	// clean up tmp dir
	defer os.RemoveAll(dir)

	err = SetParameter(dir, "testParam", "testVal")

	assert.NoError(t, err, "Error occurred but none expected")
	assert.Equal(t, "testVal", GetParameter(dir, "testParam"))
}

func TestReadFromDisk(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}

	// clean up tmp dir
	defer os.RemoveAll(dir)

	assert.Equal(t, "", GetParameter(dir, "testParamNotExistingYet"))
}

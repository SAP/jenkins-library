package piperenv

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func Test_writeMapToDisk(t *testing.T) {
	t.Parallel()
	testMap := CPEMap{
		"A/B": "Hallo",
		"sub": map[string]interface{}{
			"A/B": "Test",
		},
		"number": 5,
	}

	tmpDir, err := ioutil.TempDir(os.TempDir(), "test-data-*")
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	err = testMap.WriteToDisk(tmpDir)
	assert.NoError(t, err)

	testData := []struct {
		Path          string
		ExpectedValue string
	}{
		{
			Path:          "A/B",
			ExpectedValue: "Hallo",
		},
		{
			Path:          "sub.json",
			ExpectedValue: "{\"A/B\":\"Test\"}",
		},
		{
			Path:          "number.json",
			ExpectedValue: "5",
		},
	}

	for _, testCase := range testData {
		t.Run(fmt.Sprintf("check path %s", testCase.Path), func(t *testing.T) {
			tPath := path.Join(tmpDir, testCase.Path)
			bytes, err := ioutil.ReadFile(tPath)
			assert.NoError(t, err)
			assert.Equal(t, testCase.ExpectedValue, string(bytes))
		})
	}
}

func TestCPEMap_LoadFromDisk(t *testing.T) {
	t.Parallel()
	tmpDir, err := ioutil.TempDir(os.TempDir(), "test-data-*")
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	err = ioutil.WriteFile(path.Join(tmpDir, "Foo"), []byte("Bar"), 0644)
	assert.NoError(t, err)
	err = ioutil.WriteFile(path.Join(tmpDir, "Hello"), []byte("World"), 0644)
	assert.NoError(t, err)
	subPath := path.Join(tmpDir, "Batman")
	err = os.Mkdir(subPath, 0744)
	assert.NoError(t, err)
	err = ioutil.WriteFile(path.Join(subPath, "Bruce"), []byte("Wayne"), 0644)
	assert.NoError(t, err)
	err = ioutil.WriteFile(path.Join(subPath, "Test.json"), []byte("54"), 0644)
	assert.NoError(t, err)

	cpe := CPEMap{}
	err = cpe.LoadFromDisk(tmpDir)
	assert.NoError(t, err)

	assert.Equal(t, "Bar", cpe["Foo"])
	assert.Equal(t, "World", cpe["Hello"])
	assert.Equal(t, "Wayne", cpe["Batman/Bruce"])
	assert.Equal(t, float64(54), cpe["Batman/Test"])
}

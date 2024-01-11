//go:build unit
// +build unit

package piperenv

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
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

	tmpDir := t.TempDir()
	err := testMap.WriteToDisk(tmpDir)
	require.NoError(t, err)

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
			bytes, err := os.ReadFile(tPath)
			require.NoError(t, err)
			require.Equal(t, testCase.ExpectedValue, string(bytes))
		})
	}
}

func TestCPEMap_LoadFromDisk(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	err := os.WriteFile(path.Join(tmpDir, "Foo"), []byte("Bar"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(path.Join(tmpDir, "Hello"), []byte("World"), 0644)
	require.NoError(t, err)
	subPath := path.Join(tmpDir, "Batman")
	err = os.Mkdir(subPath, 0744)
	require.NoError(t, err)
	err = os.WriteFile(path.Join(subPath, "Bruce"), []byte("Wayne"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(path.Join(subPath, "Robin"), []byte("toBeEmptied"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(path.Join(subPath, "Test.json"), []byte("54"), 0644)
	require.NoError(t, err)

	cpe := CPEMap{}
	err = cpe.LoadFromDisk(tmpDir)
	require.NoError(t, err)

	require.Equal(t, "Bar", cpe["Foo"])
	require.Equal(t, "World", cpe["Hello"])
	require.Equal(t, "", cpe["Batman/Robin"])
	require.Equal(t, "Wayne", cpe["Batman/Bruce"])
	require.Equal(t, json.Number("54"), cpe["Batman/Test"])
}

func TestNumbersArePassedCorrectly(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	const jsonNumber = "5.5000"
	err := os.WriteFile(path.Join(tmpDir, "test.json"), []byte(jsonNumber), 0644)
	require.NoError(t, err)

	cpeMap := CPEMap{}
	err = cpeMap.LoadFromDisk(tmpDir)
	require.NoError(t, err)

	rawJSON, err := json.Marshal(cpeMap["test"])
	require.NoError(t, err)
	require.Equal(t, jsonNumber, string(rawJSON))
}

func TestCommonPipelineEnvDirNotPresent(t *testing.T) {
	cpe := CPEMap{}
	err := cpe.LoadFromDisk("/path/does/not/exist")
	require.NoError(t, err)
	require.Len(t, cpe, 0)
}

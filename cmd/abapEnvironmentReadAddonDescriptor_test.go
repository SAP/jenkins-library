package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadAddonYAML(t *testing.T) {

	t.Run("Test: success case", func(t *testing.T) {

		dir, err := ioutil.TempDir("", "test read addon descriptor")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
			_ = os.RemoveAll(dir)
		}()

		body := `---
addonProduct: /DMO/myAddonProduct
addonVersion: 3.1.4
addonUniqueId: myAddonId
customerID: 1234
repositories:
  - name: /DMO/REPO_A
    tag: v-1.0.1-build-0001
    version: 1.0.1
  - name: /DMO/REPO_B
    tag: rel-2.1.1-build-0001
    version: 2.1.1
`
		file, _ := os.Create("filename.yaml")
		file.Write([]byte(body))

		config := abapEnvironmentReadAddonDescriptorOptions{
			FileName: "filename.yaml",
		}
		cpe := abapEnvironmentReadAddonDescriptorCommonPipelineEnvironment{}
		abapEnvironmentReadAddonDescriptor(config, nil, &cpe)

		assert.Equal(t, `["/DMO/REPO_A","/DMO/REPO_B"]`, cpe.abap.repositoryNames)
		assert.Equal(t, `/DMO/myAddonProduct`, cpe.abap.addonProduct)
		assert.Equal(t, `3.1.4`, cpe.abap.addonVersion)
		assert.Equal(t, `myAddonId`, cpe.abap.addonUniqueID)
		assert.Equal(t, `1234`, cpe.abap.customerID)
		assert.Equal(t, `[{"name":"/DMO/REPO_A","tag":"v-1.0.1-build-0001","version":"1.0.1"},{"name":"/DMO/REPO_B","tag":"rel-2.1.1-build-0001","version":"2.1.1"}]`, cpe.abap.repositories)
	})

	t.Run("Test: file does not exist", func(t *testing.T) {

		config := abapEnvironmentReadAddonDescriptorOptions{
			FileName: "filename.yaml",
		}
		cpe := abapEnvironmentReadAddonDescriptorCommonPipelineEnvironment{}
		err := runAbapEnvironmentReadAddonDescriptor(&config, nil, nil, &cpe)

		assert.EqualError(t, err, fmt.Sprintf("Could not find %v.", "filename.yaml"))

	})
}

//go:build unit
// +build unit

package abaputils

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readFileMock(FileName string) ([]byte, error) {
	return []byte(FileName), nil
}

func TestAddonDescriptorNew(t *testing.T) {
	var addonDescriptor AddonDescriptor
	err := addonDescriptor.initFromYmlFile(TestAddonDescriptorYAML, readFileMock)
	assert.NoError(t, err)
	err = CheckAddonDescriptorForRepositories(addonDescriptor)
	assert.NoError(t, err)

	t.Run("Import addon.yml", func(t *testing.T) {
		assert.Equal(t, "/DMO/myAddonProduct", addonDescriptor.AddonProduct)
		assert.Equal(t, "/DMO/REPO_A", addonDescriptor.Repositories[0].Name)
		assert.Equal(t, "JEK8S273S", addonDescriptor.Repositories[1].CommitID)
		assert.Equal(t, "FR", addonDescriptor.Repositories[1].Languages[2])
		assert.Equal(t, `ISO-DEENFR`, addonDescriptor.Repositories[1].GetAakAasLanguageVector())
		assert.Equal(t, true, addonDescriptor.Repositories[1].UseClassicCTS)
	})

	t.Run("getRepositoriesInBuildScope", func(t *testing.T) {
		assert.Equal(t, 2, len(addonDescriptor.Repositories))
		addonDescriptor.Repositories[1].InBuildScope = true

		repos := addonDescriptor.GetRepositoriesInBuildScope()
		assert.Equal(t, 1, len(repos))
		assert.Equal(t, "/DMO/REPO_B", repos[0].Name)
	})

	t.Run("AsReducedJson", func(t *testing.T) {
		assert.NotContains(t, "commitID", addonDescriptor.AsReducedJson())
	})
}

var TestAddonDescriptorYAML = `---
addonProduct: /DMO/myAddonProduct
addonVersion: 3.1.4
repositories:
   - name: /DMO/REPO_A
     tag: v-1.0.1
     commitID: 89fLKS273S
     branch: release-v.1.0.1
     version: 1.0.1
     languages:
        - DE
        - EN
   - name: /DMO/REPO_B
     tag: rel-2.1.1
     commitID: JEK8S273S
     branch: release-v.2.1.1
     version: 2.1.1
     languages:
        - DE
        - EN
        - FR
     useClassicCTS: true`

func TestReadAddonDescriptor(t *testing.T) {
	t.Run("Test: success case", func(t *testing.T) {

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		file, _ := os.Create("filename.yaml")
		file.Write([]byte(TestAddonDescriptorYAML))

		addonDescriptor, err := ReadAddonDescriptor("filename.yaml")
		assert.NoError(t, err)
		assert.Equal(t, `/DMO/myAddonProduct`, addonDescriptor.AddonProduct)
		assert.Equal(t, `3.1.4`, addonDescriptor.AddonVersionYAML)
		assert.Equal(t, ``, addonDescriptor.AddonSpsLevel)
		assert.Equal(t, `/DMO/REPO_A`, addonDescriptor.Repositories[0].Name)
		assert.Equal(t, `/DMO/REPO_B`, addonDescriptor.Repositories[1].Name)
		assert.Equal(t, `v-1.0.1`, addonDescriptor.Repositories[0].Tag)
		assert.Equal(t, `rel-2.1.1`, addonDescriptor.Repositories[1].Tag)
		assert.Equal(t, `release-v.1.0.1`, addonDescriptor.Repositories[0].Branch)
		assert.Equal(t, `release-v.2.1.1`, addonDescriptor.Repositories[1].Branch)
		assert.Equal(t, `1.0.1`, addonDescriptor.Repositories[0].VersionYAML)
		assert.Equal(t, `2.1.1`, addonDescriptor.Repositories[1].VersionYAML)
		assert.Equal(t, ``, addonDescriptor.Repositories[0].SpLevel)
		assert.Equal(t, ``, addonDescriptor.Repositories[1].SpLevel)

		err = CheckAddonDescriptorForRepositories(addonDescriptor)
		assert.NoError(t, err)
	})
	t.Run("Test: file does not exist", func(t *testing.T) {
		expectedErrorMessage := "AddonDescriptor doesn't contain any repositories"

		addonDescriptor, err := ReadAddonDescriptor("filename.yaml")
		assert.EqualError(t, err, fmt.Sprintf("Could not find %v", "filename.yaml"))
		assert.Equal(t, AddonDescriptor{}, addonDescriptor)

		err = CheckAddonDescriptorForRepositories(addonDescriptor)
		assert.EqualError(t, err, expectedErrorMessage)
	})
	t.Run("Test: empty config - failure case", func(t *testing.T) {
		expectedErrorMessage := "AddonDescriptor doesn't contain any repositories"

		addonDescriptor, err := ReadAddonDescriptor("")

		assert.EqualError(t, err, fmt.Sprintf("Could not find %v", ""))
		assert.Equal(t, AddonDescriptor{}, addonDescriptor)

		err = CheckAddonDescriptorForRepositories(addonDescriptor)
		assert.EqualError(t, err, expectedErrorMessage)
	})
	t.Run("Read empty addon descriptor from wrong config - failure case", func(t *testing.T) {
		expectedErrorMessage := "AddonDescriptor doesn't contain any repositories"
		expectedRepositoryList := AddonDescriptor{Repositories: []Repository{{}, {}}}

		dir := t.TempDir()
		oldCWD, _ := os.Getwd()
		_ = os.Chdir(dir)
		// clean up tmp dir
		defer func() {
			_ = os.Chdir(oldCWD)
		}()

		manifestFileString := `
      repositories:
      - repo: 'testRepo'
      - repo: 'testRepo2'`

		err := os.WriteFile("repositories.yml", []byte(manifestFileString), 0644)
		assert.NoError(t, err)

		addonDescriptor, err := ReadAddonDescriptor("repositories.yml")

		assert.Equal(t, expectedRepositoryList, addonDescriptor)
		assert.NoError(t, err)

		err = CheckAddonDescriptorForRepositories(addonDescriptor)
		assert.EqualError(t, err, expectedErrorMessage)
	})
}

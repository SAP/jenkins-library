package abaputils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func readFileMock(FileName string) ([]byte, error) {
	return []byte(FileName), nil
}

func TestAddonDescriptor(t *testing.T) {
	t.Run("Import addon.yml", func(t *testing.T) {
		var addonDescriptor AddonDescriptor
		err := addonDescriptor.initFromYmlFile(TestAddonDescriptorYAML, readFileMock)
		CheckAddonDescriptorForRepositories(addonDescriptor)

		assert.NoError(t, err)
		assert.Equal(t, "/DMO/myAddonProduct", addonDescriptor.AddonProduct)
		assert.Equal(t, "/DMO/REPO_A", addonDescriptor.Repositories[0].Name)
		assert.Equal(t, "JEK8S273S", addonDescriptor.Repositories[1].CommitID)
		assert.Equal(t, "FR", addonDescriptor.Repositories[1].Languages[2])
		assert.Equal(t, `ISO-DEENFR`, addonDescriptor.Repositories[1].GetAakAasLanguageVector())
	})
}

var TestAddonDescriptorYAML = `---
addonProduct: /DMO/myAddonProduct
addonVersion: 3.1.4
repositories:
   - name: /DMO/REPO_A
     tag: v-1.0.1 // still open
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
        - FR`

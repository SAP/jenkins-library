package cmd

import (
	"encoding/json"
	"testing"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func mockReader(path string) ([]byte, error) {
	var file []byte
	if path == "exists" {
		return file, nil
	}
	return file, errors.New("error reading the file")
}

func TestRegisterPackagesStep(t *testing.T) {
	t.Run("step success", func(t *testing.T) {
		client := &abaputils.ClientMock{
			Body:       responseRegisterPackagesPost,
			Token:      "myToken",
			StatusCode: 200,
		}

		var config abapAddonAssemblyKitRegisterPackagesOptions
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					PackageName:    "SAPK-002AAINDRNMSPC",
					Status:         "P",
					SarXMLFilePath: "exists",
				},
				{
					PackageName: "SAPK-001AAINDRNMSPC",
					Status:      "R",
				},
			},
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)

		var cpe abapAddonAssemblyKitRegisterPackagesCommonPipelineEnvironment
		err := runAbapAddonAssemblyKitRegisterPackages(&config, nil, client, &cpe, mockReader)

		assert.NoError(t, err, "Did not expect error")
		var addonDescriptorFinal abaputils.AddonDescriptor
		json.Unmarshal([]byte(cpe.abap.addonDescriptor), &addonDescriptorFinal)
		assert.Equal(t, "L", addonDescriptorFinal.Repositories[0].Status)
	})

	t.Run("step error - invalid input", func(t *testing.T) {

	})
}

// ********************* Test uploadSarFiles *******************
func TestUploadSarFiles(t *testing.T) {
	t.Run("test uploadSarFiles", func(t *testing.T) {
		client := abaputils.ClientMock{
			Body:       "dummy",
			Token:      "myToken",
			StatusCode: 200,
		}
		repositories, conn := setupRepos("exists", planned, client)
		err := uploadSarFiles(repositories, conn, mockReader)
		assert.NoError(t, err)
	})
}

func TestUploadSarFilesInvalidInput(t *testing.T) {
	t.Run("test uploadSarFiles with missing file path", func(t *testing.T) {
		repositories, conn := setupRepos("", planned, abaputils.ClientMock{})
		err := uploadSarFiles(repositories, conn, mockReader)
		assert.Error(t, err)
	})
}

func TestUploadSarFilesNoFile(t *testing.T) {
	t.Run("test uploadSarFiles with missing file", func(t *testing.T) {
		repositories, conn := setupRepos("does_not_exist", planned, abaputils.ClientMock{})
		err := uploadSarFiles(repositories, conn, mockReader)
		assert.Error(t, err)
	})
}

func TestUploadSarFilesErrorUploading(t *testing.T) {
	t.Run("test uploadSarFiles with error during upload", func(t *testing.T) {
		client := abaputils.ClientMock{
			Body:       "ErrorBody",
			Token:      "myToken",
			StatusCode: 400,
			Error:      errors.New("Failure during upload of SAR file"),
		}
		repositories, conn := setupRepos("exists", planned, client)
		err := uploadSarFiles(repositories, conn, mockReader)
		assert.Error(t, err)
	})
}

// ********************* Test registerPackages *******************
func TestRegisterPackages(t *testing.T) {
	t.Run("test registerPackages", func(t *testing.T) {
		client := abaputils.ClientMock{
			Body:       responseRegisterPackagesPost,
			Token:      "myToken",
			StatusCode: 200,
		}
		repositories, conn := setupRepos("Filepath", planned, client)
		repos, err := registerPackages(repositories, conn)
		assert.NoError(t, err)
		assert.Equal(t, string(locked), repos[0].Status)
	})
}

func TestRegisterPackagesReleased(t *testing.T) {
	t.Run("test registerPackages", func(t *testing.T) {
		repositories, conn := setupRepos("Filepath", released, abaputils.ClientMock{})
		repos, err := registerPackages(repositories, conn)
		assert.NoError(t, err)
		assert.Equal(t, string(released), repos[0].Status)
	})
}

func TestRegisterPackagesError(t *testing.T) {
	t.Run("test registerPackages with error", func(t *testing.T) {
		client := abaputils.ClientMock{
			Body:       "ErrorBody",
			Token:      "myToken",
			StatusCode: 400,
			Error:      errors.New("Failure during registration"),
		}
		repositories, conn := setupRepos("Filepath", planned, client)
		repos, err := registerPackages(repositories, conn)
		assert.Error(t, err)
		assert.Equal(t, string(planned), repos[0].Status)
	})
}

// ********************* Test Setup *******************
func setupRepos(filePath string, status packageStatus, cl abaputils.ClientMock) ([]abaputils.Repository, connector) {
	repositories := []abaputils.Repository{
		{
			Name:           "/DRNMSPC/COMP01",
			VersionYAML:    "1.0.0",
			PackageName:    "SAPK-001AAINDRNMSPC",
			Status:         string(status),
			SarXMLFilePath: filePath,
		},
	}
	conn := new(connector)
	conn.Client = &cl
	conn.Header = make(map[string][]string)
	return repositories, *conn
}

// ********************* Testdata *******************

var responseRegisterPackagesPost = `{
    "d": {
        "__metadata": {
            "id": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/OcsPackageSet('SAPK-001AAINDRNMSPC')",
            "uri": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/OcsPackageSet('SAPK-001AAINDRNMSPC')",
            "type": "SSDA.AAS_ODATA_PACKAGE_SRV.OcsPackage"
        },
        "Name": "SAPK-001AAINDRNMSPC",
        "Type": "AOI",
        "Component": "/DRNMSPC/COMP01",
        "Release": "0001",
        "Level": "0000",
        "Status": "L",
        "Operation": "",
        "Namespace": "/DRNMSPC/",
        "Vendorid": "0000203069"
    }
}`

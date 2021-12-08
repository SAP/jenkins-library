package cmd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCheckCVsStep(t *testing.T) {
	var config abapAddonAssemblyKitCheckCVsOptions
	var cpe abapAddonAssemblyKitCheckCVsCommonPipelineEnvironment
	bundle := aakaas.NewAakBundleMock()
	bundle.SetBody(responseCheckCVs)
	utils := bundle.GetUtils()
	config.Username = "dummyUser"
	config.Password = "dummyPassword"
	t.Run("step success", func(t *testing.T) {
		config.AddonDescriptorFileName = "success"
		err := runAbapAddonAssemblyKitCheckCVs(&config, nil, &utils, &cpe)
		assert.NoError(t, err, "Did not expect error")
		var addonDescriptorFinal abaputils.AddonDescriptor
		json.Unmarshal([]byte(cpe.abap.addonDescriptor), &addonDescriptorFinal)
		assert.Equal(t, "0001", addonDescriptorFinal.Repositories[0].Version)
		assert.Equal(t, "0002", addonDescriptorFinal.Repositories[0].SpLevel)
		assert.Equal(t, "0003", addonDescriptorFinal.Repositories[0].PatchLevel)
		assert.Equal(t, "HUGO1234", addonDescriptorFinal.Repositories[0].CommitID)
	})
	t.Run("step error - in validate(no CommitID)", func(t *testing.T) {
		config.AddonDescriptorFileName = "noCommitID"
		err := runAbapAddonAssemblyKitCheckCVs(&config, nil, &utils, &cpe)
		assert.Error(t, err, "Must end with error")
		assert.Contains(t, err.Error(), "CommitID missing in repo")
	})
	t.Run("step error - in ReadAddonDescriptor", func(t *testing.T) {
		config.AddonDescriptorFileName = "failing"
		err := runAbapAddonAssemblyKitCheckCVs(&config, nil, &utils, &cpe)
		assert.Error(t, err, "Must end with error")
		assert.Contains(t, "error in ReadAddonDescriptor", err.Error())
	})
	t.Run("step error - in validate", func(t *testing.T) {
		config.AddonDescriptorFileName = "success"
		bundle.SetBody("ErrorBody")
		bundle.SetError("error during validation")
		err := runAbapAddonAssemblyKitCheckCVs(&config, nil, &utils, &cpe)
		assert.Error(t, err, "Must end with error")
	})
}

func TestInitCV(t *testing.T) {
	t.Run("test init", func(t *testing.T) {
		conn := new(abapbuild.Connector)
		conn.Client = &abaputils.ClientMock{}
		repo := abaputils.Repository{
			Name:        "/DRNMSPC/COMP01",
			VersionYAML: "1.2.3",
		}
		var c componentVersion
		c.initCV(repo, *conn)
		assert.Equal(t, "/DRNMSPC/COMP01", c.Name)
		assert.Equal(t, "1.2.3", c.VersionYAML)
	})
}

func TestValidateCV(t *testing.T) {
	conn := new(abapbuild.Connector)
	t.Run("test validate - success", func(t *testing.T) {
		conn.Client = &abaputils.ClientMock{
			Body: responseCheckCVs,
		}
		c := componentVersion{
			Connector:   *conn,
			Name:        "/DRNMSPC/COMP01",
			VersionYAML: "1.2.3",
			CommitID:    "HUGO1234",
		}
		conn.Client = &abaputils.ClientMock{
			Body: responseCheckCVs,
		}
		err := c.validate()
		assert.NoError(t, err)
		assert.Equal(t, "0001", c.Version)
		assert.Equal(t, "0002", c.SpLevel)
		assert.Equal(t, "0003", c.PatchLevel)
	})
	t.Run("test validate - with error", func(t *testing.T) {
		conn.Client = &abaputils.ClientMock{
			Body:  "ErrorBody",
			Error: errors.New("Validation failed"),
		}
		c := componentVersion{
			Connector:   *conn,
			Name:        "/DRNMSPC/COMP01",
			VersionYAML: "1.2.3",
			CommitID:    "HUGO1234",
		}
		err := c.validate()
		assert.Error(t, err)
		assert.Equal(t, "", c.Version)
		assert.Equal(t, "", c.SpLevel)
		assert.Equal(t, "", c.PatchLevel)
	})
}

func TestCopyFieldsCV(t *testing.T) {
	t.Run("test copyFieldsToRepo", func(t *testing.T) {
		repo := abaputils.Repository{
			Name:        "/DRNMSPC/COMP01",
			VersionYAML: "1.2.3",
		}
		var c componentVersion
		c.Version = "0001"
		c.SpLevel = "0002"
		c.PatchLevel = "0003"
		c.copyFieldsToRepo(&repo)
		assert.Equal(t, "0001", repo.Version)
		assert.Equal(t, "0002", repo.SpLevel)
		assert.Equal(t, "0003", repo.PatchLevel)
	})
}

func TestCombineYAMLRepositoriesWithCPEProduct(t *testing.T) {
	t.Run("test combineYAMLRepositoriesWithCPEProduct", func(t *testing.T) {
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					Name:        "/DRNMSPC/COMP01",
					VersionYAML: "1.2.3",
				},
				{
					Name:        "/DRNMSPC/COMP02",
					VersionYAML: "3.2.1",
				},
			},
		}
		addonDescriptorFromCPE := abaputils.AddonDescriptor{
			AddonProduct:     "/DRNMSP/PROD",
			AddonVersionYAML: "1.2.3",
		}
		finalAddonDescriptor := combineYAMLRepositoriesWithCPEProduct(addonDescriptor, addonDescriptorFromCPE)
		assert.Equal(t, "/DRNMSP/PROD", finalAddonDescriptor.AddonProduct)
		assert.Equal(t, "1.2.3", finalAddonDescriptor.AddonVersionYAML)
		assert.Equal(t, "/DRNMSPC/COMP01", finalAddonDescriptor.Repositories[0].Name)
		assert.Equal(t, "/DRNMSPC/COMP02", finalAddonDescriptor.Repositories[1].Name)
		assert.Equal(t, "1.2.3", finalAddonDescriptor.Repositories[0].VersionYAML)
		assert.Equal(t, "3.2.1", finalAddonDescriptor.Repositories[1].VersionYAML)
	})
}

var responseCheckCVs = `{
    "d": {
        "__metadata": {
            "id": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/SoftwareComponentVersionSet(Name='%2FDRNMSPC%2FCOMP01',Version='0001')",
            "uri": "https://W7Q.DMZWDF.SAP.CORP:443/odata/aas_ocs_package/SoftwareComponentVersionSet(Name='%2FDRNMSPC%2FCOMP01',Version='0001')",
            "type": "SSDA.AAS_ODATA_PACKAGE_SRV.SoftwareComponentVersion"
        },
        "Name": "/DRNMSPC/COMP01",
        "Version": "0001",
        "SpLevel": "0002",
        "PatchLevel": "0003",
        "Vendor": "",
        "VendorType": ""
    }
}`

//***************************************
type AakBundleMock struct {
	*mock.ExecMockRunner
	*abaputils.ClientMock
	maxRuntime time.Duration
}

func NewAakBundleMock() *AakBundleMock {
	utils := AakBundleMock{
		ExecMockRunner: &mock.ExecMockRunner{},
		ClientMock:     &abaputils.ClientMock{},
		maxRuntime:     1 * time.Second,
	}
	return &utils
}

func (bundle *AakBundleMock) GetUtils() aakaas.AakUtils {
	return bundle
}

func (bundle *AakBundleMock) GetMaxRuntime() time.Duration {
	return bundle.maxRuntime
}

func (bundle *AakBundleMock) SetMaxRuntime(maxRuntime time.Duration) {
	bundle.maxRuntime = maxRuntime
}

func (bundle *AakBundleMock) GetPollingInterval() time.Duration {
	return 1 * time.Microsecond
}

func (bundle *AakBundleMock) SetBodyList(bodyList []string) {
	bundle.ClientMock.Body = ""
	bundle.ClientMock.BodyList = bodyList
}

func (bundle *AakBundleMock) SetBody(body string) {
	bundle.ClientMock.Body = body
	bundle.ClientMock.BodyList = []string{}
}

func (bundle *AakBundleMock) SetError(errorText string) {
	bundle.ClientMock.Error = errors.New(errorText)
}

func (bundle *AakBundleMock) ReadAddonDescriptor(FileName string) (abaputils.AddonDescriptor, error) {
	var addonDescriptor abaputils.AddonDescriptor
	var err error
	switch FileName {
	case "success":
		{
			addonDescriptor = abaputils.AddonDescriptor{
				AddonProduct:     "/DRNMSPC/PRD01",
				AddonVersionYAML: "3.2.1",
				Repositories: []abaputils.Repository{
					{
						Name:        "/DRNMSPC/COMP01",
						VersionYAML: "1.2.3",
						CommitID:    "HUGO1234",
					},
				},
			}
		}
	case "noCommitID":
		{
			addonDescriptor = abaputils.AddonDescriptor{
				AddonProduct:     "/DRNMSPC/PRD01",
				AddonVersionYAML: "3.2.1",
				Repositories: []abaputils.Repository{
					{
						Name:        "/DRNMSPC/COMP01",
						VersionYAML: "1.2.3",
					},
				},
			}
		}
	case "failing":
		{
			err = errors.New("error in ReadAddonDescriptor")
		}
	}
	return addonDescriptor, err
}

//*****************************other client mock *******************************
type AakBundleMockNewMC struct {
	*mock.ExecMockRunner
	*abapbuild.MockClient
	maxRuntime time.Duration
}

func NewAakBundleMockNewMC(mC *abapbuild.MockClient) *AakBundleMockNewMC {
	utils := AakBundleMockNewMC{
		ExecMockRunner: &mock.ExecMockRunner{},
		MockClient:     mC,
		maxRuntime:     1 * time.Second,
	}
	return &utils
}

func (bundle *AakBundleMockNewMC) GetUtils() aakaas.AakUtils {
	return bundle
}

func (bundle *AakBundleMockNewMC) GetMaxRuntime() time.Duration {
	return bundle.maxRuntime
}

func (bundle *AakBundleMockNewMC) GetPollingInterval() time.Duration {
	return 1 * time.Microsecond
}

func (bundle *AakBundleMockNewMC) ReadAddonDescriptor(FileName string) (abaputils.AddonDescriptor, error) {
	var addonDescriptor abaputils.AddonDescriptor
	err := errors.New("don't use this")
	return addonDescriptor, err
}

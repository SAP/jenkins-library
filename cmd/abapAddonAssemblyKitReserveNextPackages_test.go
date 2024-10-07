//go:build unit
// +build unit

package cmd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func arrangeConfig(addonDescriptor abaputils.AddonDescriptor) abapAddonAssemblyKitReserveNextPackagesOptions {
	var config abapAddonAssemblyKitReserveNextPackagesOptions
	config.Username = "dummy"
	config.Password = "dummy"
	adoDesc, _ := json.Marshal(addonDescriptor)
	config.AddonDescriptor = string(adoDesc)
	return config
}

func TestReserveNextPackagesStep(t *testing.T) {
	var cpe abapAddonAssemblyKitReserveNextPackagesCommonPipelineEnvironment
	bundle := aakaas.NewAakBundleMock()
	utils := bundle.GetUtils()
	t.Run("step success tag", func(t *testing.T) {
		config := arrangeConfig(abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					Name:        "/DRNMSPC/COMP01",
					VersionYAML: "1.0.0.",
					Tag:         "myTestTagv1.0",
				},
				{
					Name:        "/DRNMSPC/COMP02",
					VersionYAML: "1.0.0.",
					Tag:         "myTestTagv1.0",
				},
			},
		})
		bodyList := []string{responseReserveNextPackageReleased, responseReserveNextPackagePlanned, responseReserveNextPackagePostReleased, "myToken", responseReserveNextPackagePostPlanned, "myToken"}
		bundle.SetBodyList(bodyList)
		err := runAbapAddonAssemblyKitReserveNextPackages(&config, &utils, &cpe)
		assert.NoError(t, err, "Did not expect error")
		var addonDescriptorFinal abaputils.AddonDescriptor
		err = json.Unmarshal([]byte(cpe.abap.addonDescriptor), &addonDescriptorFinal)
		assert.NoError(t, err)
		assert.Equal(t, "P", addonDescriptorFinal.Repositories[0].Status)
		assert.Equal(t, "R", addonDescriptorFinal.Repositories[1].Status)
	})

	t.Run("step success", func(t *testing.T) {
		config := arrangeConfig(abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					Name:        "/DRNMSPC/COMP01",
					VersionYAML: "1.0.0.",
					CommitID:    "hugo",
				},
				{
					Name:        "/DRNMSPC/COMP02",
					VersionYAML: "1.0.0.",
					CommitID:    "something40charslongxxxxxxxxxxxxxxxxxxxx",
				},
			},
		})
		bodyList := []string{responseReserveNextPackageReleased, responseReserveNextPackagePlanned, responseReserveNextPackagePostReleased, "myToken", responseReserveNextPackagePostPlanned, "myToken"}
		bundle.SetBodyList(bodyList)
		err := runAbapAddonAssemblyKitReserveNextPackages(&config, &utils, &cpe)
		assert.NoError(t, err, "Did not expect error")
		var addonDescriptorFinal abaputils.AddonDescriptor
		err = json.Unmarshal([]byte(cpe.abap.addonDescriptor), &addonDescriptorFinal)
		assert.NoError(t, err)
		assert.Equal(t, "P", addonDescriptorFinal.Repositories[0].Status)
		assert.Equal(t, "R", addonDescriptorFinal.Repositories[1].Status)
	})
	t.Run("step error - invalid input", func(t *testing.T) {
		config := arrangeConfig(abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					Name: "/DRNMSPC/COMP01",
				},
			},
		})
		err := runAbapAddonAssemblyKitReserveNextPackages(&config, &utils, &cpe)
		assert.Error(t, err, "Did expect error")
	})
	t.Run("step error - timeout", func(t *testing.T) {
		config := arrangeConfig(abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					Name:        "/DRNMSPC/COMP01",
					VersionYAML: "1.0.0.",
				},
			},
		})
		bodyList := []string{responseReserveNextPackageCreationTriggered, responseReserveNextPackagePostPlanned, "myToken"}
		bundle.SetBodyList(bodyList)
		bundle.SetMaxRuntime(time.Duration(1 * time.Microsecond))
		bundle.ClientMock.ErrorInsteadOfDump = true
		err := runAbapAddonAssemblyKitReserveNextPackages(&config, &utils, &cpe)
		assert.Error(t, err, "Did expect error")
	})
}

// ********************* Test init *******************
func TestInitPackage(t *testing.T) {
	t.Run("test init", func(t *testing.T) {
		conn := new(abapbuild.Connector)
		conn.Client = &abaputils.ClientMock{}
		repo := abaputils.Repository{
			Name:        "/DRNMSPC/COMP01",
			VersionYAML: "1.0.0",
		}
		var p aakaas.Package
		p.InitPackage(repo, *conn)
		assert.Equal(t, "/DRNMSPC/COMP01", p.ComponentName)
		assert.Equal(t, "1.0.0", p.VersionYAML)
	})
}

// ********************* Test copyFieldsToRepositories *******************
func TestCopyFieldsToRepositoriesPackage(t *testing.T) {
	pckgWR := []aakaas.PackageWithRepository{
		{
			Package: aakaas.Package{
				ComponentName: "/DRNMSPC/COMP01",
				VersionYAML:   "1.0.0",
				PackageName:   "SAPK-001AAINDRNMSPC",
				Type:          "AOI",
				Namespace:     "/DRNMSPC/",
			},
			Repo: abaputils.Repository{
				Name:        "/DRNMSPC/COMP01",
				VersionYAML: "1.0.0",
			},
		},
	}

	t.Run("test copyFieldsToRepositories Planned success w/o predecessorcommitID", func(t *testing.T) {
		pckgWR[0].Package.Status = aakaas.PackageStatusPlanned
		pckgWR[0].Package.PredecessorCommitID = ""
		pckgWR[0].Repo.CommitID = "something40charslongxxxxxxxxxxxxxxxxxxxx"
		pckgWR[0].Package.CommitID = "something40charslongxxxxxxxxxxxxxxxxxxxx"
		repos, err := checkAndCopyFieldsToRepositories(pckgWR)
		assert.Equal(t, "SAPK-001AAINDRNMSPC", repos[0].PackageName)
		assert.Equal(t, "AOI", repos[0].PackageType)
		assert.Equal(t, string(aakaas.PackageStatusPlanned), repos[0].Status)
		assert.Equal(t, "/DRNMSPC/", repos[0].Namespace)
		assert.NoError(t, err)
	})

	t.Run("test copyFieldsToRepositories Planned success with predecessorcommitID", func(t *testing.T) {
		pckgWR[0].Package.Status = aakaas.PackageStatusPlanned
		pckgWR[0].Package.PredecessorCommitID = "something40charslongPREDECESSORyyyyyyyyy"
		pckgWR[0].Repo.CommitID = "something40charslongxxxxxxxxxxxxxxxxxxxx"
		pckgWR[0].Package.CommitID = "something40charslongxxxxxxxxxxxxxxxxxxxx"
		_, err := checkAndCopyFieldsToRepositories(pckgWR)
		assert.NoError(t, err)
	})

	t.Run("test copyFieldsToRepositories Planned success with predecessorcommitI short AAKaaS", func(t *testing.T) {
		pckgWR[0].Package.Status = aakaas.PackageStatusPlanned
		pckgWR[0].Package.PredecessorCommitID = "predecessor"
		pckgWR[0].Repo.CommitID = "something40charslongxxxxxxxxxxxxxxxxxxxx"
		pckgWR[0].Package.CommitID = "something40charslongxxxxxxxxxxxxxxxxxxxx"
		_, err := checkAndCopyFieldsToRepositories(pckgWR)
		assert.NoError(t, err)
	})

	t.Run("test copyFieldsToRepositories Planned success with predecessorcommitID short addon.yml", func(t *testing.T) {
		pckgWR[0].Package.Status = aakaas.PackageStatusPlanned
		pckgWR[0].Package.PredecessorCommitID = "predecessor"
		pckgWR[0].Repo.CommitID = "successor"
		pckgWR[0].Package.CommitID = "successorANDsomemore"
		_, err := checkAndCopyFieldsToRepositories(pckgWR)
		assert.NoError(t, err)
	})

	t.Run("test copyFieldsToRepositories Planned error with predecessorcommitID same as commitID", func(t *testing.T) {
		pckgWR[0].Package.Status = aakaas.PackageStatusPlanned
		pckgWR[0].Repo.CommitID = "something40charslongxxxxxxxxxxxxxxxxxxxx"
		pckgWR[0].Package.PredecessorCommitID = pckgWR[0].Repo.CommitID
		pckgWR[0].Package.CommitID = "something40charslongxxxxxxxxxxxxxxxxxxxx"
		_, err := checkAndCopyFieldsToRepositories(pckgWR)
		assert.Error(t, err)
	})

	t.Run("test copyFieldsToRepositories Released success", func(t *testing.T) {
		pckgWR[0].Package.Status = aakaas.PackageStatusReleased
		pckgWR[0].Package.PredecessorCommitID = "" //released packages do not have this attribute
		pckgWR[0].Repo.CommitID = "something40charslongxxxxxxxxxxxxxxxxxxxx"
		pckgWR[0].Package.CommitID = "something40charslongxxxxxxxxxxxxxxxxxxxx"
		_, err := checkAndCopyFieldsToRepositories(pckgWR)
		assert.NoError(t, err)
	})

	t.Run("test copyFieldsToRepositories Released error, different commitIDs", func(t *testing.T) {
		pckgWR[0].Package.Status = aakaas.PackageStatusReleased
		pckgWR[0].Package.PredecessorCommitID = "" //released packages do not have this attribute
		pckgWR[0].Repo.CommitID = "something40charslongxxxxxxxxxxxxxxxxxxxx"
		pckgWR[0].Package.CommitID = "something40charslongxxxxxxxxxxxxxxxxxxxO"
		_, err := checkAndCopyFieldsToRepositories(pckgWR)
		assert.Error(t, err)
	})
}

// ********************* Test reserveNext *******************
func TestReserveNextPackage(t *testing.T) {
	t.Run("test reserveNext - success", func(t *testing.T) {
		client := abaputils.ClientMock{
			Body: responseReserveNextPackagePostPlanned,
		}
		p := testPackageSetup("/DRNMSPC/COMP01", "1.0.0", client)

		err := p.ReserveNext()
		assert.NoError(t, err)
		assert.Equal(t, "SAPK-001AAINDRNMSPC", p.PackageName)
		assert.Equal(t, "AOI", p.Type)
		assert.Equal(t, aakaas.PackageStatusPlanned, p.Status)
	})
	t.Run("test reserveNext - missing versionYAML", func(t *testing.T) {
		client := abaputils.ClientMock{}
		p := testPackageSetup("/DRNMSPC/COMP01", "", client)
		err := p.ReserveNext()
		assert.Error(t, err)
		assert.Equal(t, "", p.PackageName)
		assert.Equal(t, "", p.Type)
		assert.Equal(t, aakaas.PackageStatus(""), p.Status)
	})
	t.Run("test reserveNext - error from call", func(t *testing.T) {
		client := abaputils.ClientMock{
			Body:  "ErrorBody",
			Error: errors.New("Failure during reserve next"),
		}
		p := testPackageSetup("/DRNMSPC/COMP01", "1.0.0", client)
		err := p.ReserveNext()
		assert.Error(t, err)
		assert.Equal(t, "", p.PackageName)
		assert.Equal(t, "", p.Type)
		assert.Equal(t, aakaas.PackageStatus(""), p.Status)
	})
}

// ********************* Test reservePackages *******************

func TestReservePackages(t *testing.T) {
	t.Run("test reservePackages - success", func(t *testing.T) {
		client := abaputils.ClientMock{
			Body: responseReserveNextPackagePostPlanned,
		}
		repositories, conn := testRepositoriesSetup("/DRNMSPC/COMP01", "1.0.0", client)
		repos, err := reservePackages(repositories, conn)
		assert.NoError(t, err)
		assert.Equal(t, "/DRNMSPC/COMP01", repos[0].Package.ComponentName)
		assert.Equal(t, "1.0.0", repos[0].Package.VersionYAML)
		assert.Equal(t, aakaas.PackageStatusPlanned, repos[0].Package.Status)
	})
	t.Run("test reservePackages - error from call", func(t *testing.T) {
		client := abaputils.ClientMock{
			Body:  "ErrorBody",
			Error: errors.New("Failure during reserve next"),
		}
		repositories, conn := testRepositoriesSetup("/DRNMSPC/COMP01", "1.0.0", client)
		_, err := reservePackages(repositories, conn)
		assert.Error(t, err)
	})
}

// ********************* Test pollReserveNextPackages *******************

func TestPollReserveNextPackages(t *testing.T) {
	bundle := aakaas.NewAakBundleMock()
	utils := bundle.GetUtils()
	t.Run("test pollReserveNextPackages - testing loop", func(t *testing.T) {
		bodyList := []string{responseReserveNextPackagePlanned, responseReserveNextPackageCreationTriggered}
		bundle.SetBodyList(bodyList)
		pckgWR := testPollPackagesSetup(&utils)
		err := pollReserveNextPackages(pckgWR, &utils)
		assert.NoError(t, err)
		assert.Equal(t, aakaas.PackageStatusPlanned, pckgWR[0].Package.Status)
		assert.Equal(t, "/DRNMSPC/", pckgWR[0].Package.Namespace)
	})
	t.Run("test pollReserveNextPackages - status locked", func(t *testing.T) {
		bundle.SetBody(responseReserveNextPackageLocked)
		pckgWR := testPollPackagesSetup(&utils)
		err := pollReserveNextPackages(pckgWR, &utils)
		assert.Error(t, err)
		assert.Equal(t, aakaas.PackageStatusLocked, pckgWR[0].Package.Status)
	})
	t.Run("test pollReserveNextPackages - status released", func(t *testing.T) {
		bundle.SetBody(responseReserveNextPackageReleased)
		pckgWR := testPollPackagesSetup(&utils)
		err := pollReserveNextPackages(pckgWR, &utils)
		assert.NoError(t, err)
		assert.Equal(t, aakaas.PackageStatusReleased, pckgWR[0].Package.Status)
	})
	t.Run("test pollReserveNextPackages - unknow status", func(t *testing.T) {
		bundle.SetBody(responseReserveNextPackageUnknownState)
		pckgWR := testPollPackagesSetup(&utils)
		err := pollReserveNextPackages(pckgWR, &utils)
		assert.Error(t, err)
		assert.Equal(t, aakaas.PackageStatus("X"), pckgWR[0].Package.Status)
	})
	t.Run("test pollReserveNextPackages - timeout", func(t *testing.T) {
		bundle.SetBody(responseReserveNextPackageUnknownState)
		bundle.SetError("Failure during reserve next")
		bundle.SetMaxRuntime(time.Duration(1 * time.Microsecond))
		pckgWR := testPollPackagesSetup(&utils)
		err := pollReserveNextPackages(pckgWR, &utils)
		assert.Error(t, err)
	})
}

// ********************* Setup functions *******************

func testPollPackagesSetup(utils *aakaas.AakUtils) []aakaas.PackageWithRepository {
	conn := new(abapbuild.Connector)
	conn.Client = *utils
	conn.Header = make(map[string][]string)
	pckgWR := []aakaas.PackageWithRepository{
		{
			Package: aakaas.Package{
				Connector:     *conn,
				ComponentName: "/DRNMSPC/COMP01",
				VersionYAML:   "1.0.0",
				PackageName:   "SAPK-001AAINDRNMSPC",
				Type:          "AOI",
			},
			Repo: abaputils.Repository{},
		},
	}
	return pckgWR
}

func testRepositoriesSetup(componentName string, versionYAML string, client abaputils.ClientMock) ([]abaputils.Repository, abapbuild.Connector) {
	conn := new(abapbuild.Connector)
	conn.Client = &client
	conn.Header = make(map[string][]string)
	repositories := []abaputils.Repository{
		{
			Name:        componentName,
			VersionYAML: versionYAML,
		},
	}
	return repositories, *conn
}

func testPackageSetup(componentName string, versionYAML string, client abaputils.ClientMock) aakaas.Package {
	conn := new(abapbuild.Connector)
	conn.Client = &client
	conn.Header = make(map[string][]string)
	p := aakaas.Package{
		Connector:     *conn,
		ComponentName: componentName,
		VersionYAML:   versionYAML,
	}
	return p
}

// ********************* Testdata *******************

var responseReserveNextPackagePostPlanned = `{
    "d": {
        "DeterminePackageForScv": {
            "__metadata": {
                "type": "SSDA.AAS_ODATA_PACKAGE_SRV.PackageExtended"
            },
            "Name": "SAPK-001AAINDRNMSPC",
            "Type": "AOI",
            "ScName": "/DRNMSPC/COMP01",
            "ScVersion": "0001",
            "SpLevel": "0000",
            "PatchLevel": "0000",
            "Predecessor": "",
            "PredecessorCommitId": "",
            "Status": "P"
        }
    }
}`

var responseReserveNextPackagePostReleased = `{
    "d": {
        "DeterminePackageForScv": {
            "__metadata": {
                "type": "SSDA.AAS_ODATA_PACKAGE_SRV.PackageExtended"
            },
            "Name": "SAPK-001AAINDRNMSPC",
            "Type": "AOI",
            "ScName": "/DRNMSPC/COMP02",
            "ScVersion": "0001",
            "SpLevel": "0000",
            "PatchLevel": "0000",
            "Predecessor": "",
            "PredecessorCommitId": "",
			"CommitId": "something40charslongxxxxxxxxxxxxxxxxxxxx",
            "Status": "R"
        }
    }
}`

var responseReserveNextPackageCreationTriggered = `{
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
        "Status": "C",
        "Operation": "",
        "Namespace": "/DRNMSPC/",
        "Vendorid": "0000203069"
    }
}`

var responseReserveNextPackageLocked = `{
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

var responseReserveNextPackagePlanned = `{
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
        "Status": "P",
        "Operation": "",
        "Namespace": "/DRNMSPC/",
        "Vendorid": "0000203069"
    }
}`

var responseReserveNextPackageReleased = `{
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
        "Status": "R",
        "Operation": "",
        "Namespace": "/DRNMSPC/",
        "Vendorid": "0000203069"
    }
}`

var responseReserveNextPackageUnknownState = `{
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
        "Status": "X",
        "Operation": "",
        "Namespace": "/DRNMSPC/",
        "Vendorid": "0000203069"
    }
}`

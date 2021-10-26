package cmd

import (
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"

	"github.com/SAP/jenkins-library/pkg/abaputils"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func mockReader(path string) ([]byte, error) {
	if path == "exists" {
		return []byte("test"), nil
	} else if path == "null" {
		return []byte(""), nil
	}
	return nil, errors.New("error reading the file")
}

func testSetup(client piperhttp.Sender, buildID string) abapbuild.Build {
	conn := new(abapbuild.Connector)
	conn.Client = client
	conn.DownloadClient = &abapbuild.DownloadClientMock{}
	conn.Header = make(map[string][]string)
	b := abapbuild.Build{
		Connector: *conn,
		BuildID:   buildID,
	}
	return b
}

func TestCheckIfFailedAndPrintLogsWithError(t *testing.T) {
	t.Run("checkIfFailedAndPrintLogs with failed build", func(t *testing.T) {
		var repo abaputils.Repository
		b := testSetup(&abapbuild.ClMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		b.RunState = abapbuild.Failed
		var buildsWithRepo []buildWithRepository
		bWR := buildWithRepository{
			build: b,
			repo:  repo,
		}
		buildsWithRepo = append(buildsWithRepo, bWR)
		err := checkIfFailedAndPrintLogs(buildsWithRepo)
		assert.Error(t, err)
	})

	t.Run("checkIfFailedAndPrintLogs", func(t *testing.T) {
		var repo abaputils.Repository
		b := testSetup(&abapbuild.ClMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		b.RunState = abapbuild.Finished
		var buildsWithRepo []buildWithRepository
		bWR := buildWithRepository{
			build: b,
			repo:  repo,
		}
		buildsWithRepo = append(buildsWithRepo, bWR)
		err := checkIfFailedAndPrintLogs(buildsWithRepo)
		assert.NoError(t, err)
	})
}

func TestStartingInvalidInput(t *testing.T) {
	t.Run("Run starting with Invalid Input", func(t *testing.T) {
		client := &abapbuild.ClMock{
			Token: "MyToken",
		}
		conn := new(abapbuild.Connector)
		conn.Client = client
		conn.Header = make(map[string][]string)
		var repos []abaputils.Repository
		repo := abaputils.Repository{
			Name:   "RepoA",
			Status: "P",
		}
		repos = append(repos, repo)
		builds, err := executeBuilds(repos, *conn, time.Duration(0*time.Second), time.Duration(1*time.Millisecond))
		assert.NoError(t, err)
		assert.Equal(t, 1, len(builds))
		assert.Equal(t, abapbuild.Failed, builds[0].build.RunState)
	})
}

func TestStep(t *testing.T) {
	autils := &abaputils.AUtilsMock{
		ReturnedConnectionDetailsHTTP: abaputils.ConnectionDetailsHTTP{
			URL: `/sap/opu/odata/BUILD/CORE_SRV`,
		},
	}
	client := abapbuild.GetBuildMockClient()
	cpe := &abapEnvironmentAssemblePackagesCommonPipelineEnvironment{}

	t.Run("abapEnvironmentAssemblePackages: nothing to do", func(t *testing.T) {

		config := &abapEnvironmentAssemblePackagesOptions{
			AddonDescriptor:             cpeAbapAddonDescriptorPackageLocked,
			MaxRuntimeInMinutes:         1,
			PollIntervalsInMilliseconds: 1,
			PerformAssemblePackages:     true,
		}

		err := runAbapEnvironmentAssemblePackages(config, nil, autils, &client, cpe, mockReader)
		assert.NoError(t, err)
	})
	t.Run("abapEnvironmentAssemblePackages: build", func(t *testing.T) {

		config := &abapEnvironmentAssemblePackagesOptions{
			AddonDescriptor:             cpeAbapAddonDescriptorPackageReserved,
			MaxRuntimeInMinutes:         1,
			PollIntervalsInMilliseconds: 1,
			PerformAssemblePackages:     true,
		}

		err := runAbapEnvironmentAssemblePackages(config, nil, autils, &client, cpe, mockReader)
		assert.NoError(t, err)
		assert.Contains(t, cpe.abap.addonDescriptor, `SAPK-001AAINITAPC1.SAR`)
	})
}

/*
// Wait for Download Method available on github
func OldTestRegisterPackagesStep(t *testing.T) {
	//var config abapAddonAssemblyKitRegisterPackagesOptions
	config := &abapEnvironmentAssemblePackagesOptions{
		PerformRegisterPackages: true,
	}
	//var cpe abapAddonAssemblyKitRegisterPackagesCommonPipelineEnvironment
	cpe := &abapEnvironmentAssemblePackagesCommonPipelineEnvironment{}
	t.Run("step success", func(t *testing.T) {
		client := &abaputils.ClientMock{
			BodyList: []string{responseRegisterPackagesPost, "myToken", "dummyResponseUpload"},
		}
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
		//err := runAbapAddonAssemblyKitRegisterPackages(&config, nil, client, &cpe, mockReader)
		err := runAbapEnvironmentAssemblePackages(config, nil, nil, &client, cpe, mockReader)

		assert.NoError(t, err, "Did not expect error")
		var addonDescriptorFinal abaputils.AddonDescriptor
		json.Unmarshal([]byte(cpe.abap.addonDescriptor), &addonDescriptorFinal)
		assert.Equal(t, "L", addonDescriptorFinal.Repositories[0].Status)
	})
	t.Run("step error - null file", func(t *testing.T) {
		client := &abaputils.ClientMock{
			BodyList: []string{responseRegisterPackagesPost, "myToken", "dummyResponseUpload"},
		}
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					PackageName:    "SAPK-002AAINDRNMSPC",
					Status:         "P",
					SarXMLFilePath: "null",
				},
			},
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)
		//err := runAbapAddonAssemblyKitRegisterPackages(&config, nil, client, &cpe, mockReader)
		err := runAbapEnvironmentAssemblePackages(config, nil, nil, &client, cpe, mockReader)

		assert.Error(t, err, "Did expect error")
	})
	t.Run("step error - uploadSarFiles", func(t *testing.T) {
		client := &abaputils.ClientMock{
			Body:  "ErrorBody",
			Error: errors.New("Failure during upload of SAR file"),
		}

		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					PackageName:    "SAPK-002AAINDRNMSPC",
					Status:         "P",
					SarXMLFilePath: "exists",
				},
			},
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)
		//err := runAbapAddonAssemblyKitRegisterPackages(&config, nil, client, &cpe, mockReader)
		err := runAbapEnvironmentAssemblePackages(config, nil, nil, &client, cpe, mockReader)
		assert.Error(t, err, "Did expect error")
	})
	t.Run("step error - registerPackages - invalid input", func(t *testing.T) {
		client := &abaputils.ClientMock{
			BodyList: []string{"myToken", "dummyResponseUpload"},
		}
		addonDescriptor := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					Status:         "P",
					SarXMLFilePath: "exists",
				},
			},
		}
		adoDesc, _ := json.Marshal(addonDescriptor)
		config.AddonDescriptor = string(adoDesc)
		//err := runAbapAddonAssemblyKitRegisterPackages(&config, nil, client, &cpe, mockReader)
		err := runAbapEnvironmentAssemblePackages(config, nil, nil, &client, cpe, mockReader)
		assert.Error(t, err, "Did expect error")
	})
} */

var cpeAbapAddonDescriptorPackageLocked = `{
	"addonProduct":"/ITAPC1/I_CURRENCZPRODUCT",
	"addonVersion":"1.0.0",
	"addonVersionAAK":"0001",
	"addonUniqueID":"myAddonId",
	"customerID":"$ID",
	"AddonSpsLevel":"0000",
	"AddonPatchLevel":"0000",
	"TargetVectorID":"",
	"repositories":[
		{	"name":"/ITAPC1/I_CURRENCZ",
			"tag":"whatever",
			"branch":"",
			"commitID":"",
			"version":"1.0.0",
			"versionAAK":"0001",
			"PackageName":"SAPK-002AAINITAPC1",
			"PackageType":"AOI",
			"SpLevel":"0000",
			"PatchLevel":"0000",
			"PredecessorCommitID":"",
			"Status":"L",
			"Namespace":"/ITAPC1/",
			"SarXMLFilePath":".pipeline\\commonPipelineEnvironment\\abap\\SAPK-002AAINITAPC1.SAR"
		}
	]
}`

var cpeAbapAddonDescriptorPackageReserved = `{
	"addonProduct":"/ITAPC1/I_CURRENCZPRODUCT",
	"addonVersion":"1.0.0",
	"addonVersionAAK":"0001",
	"addonUniqueID":"myAddonId",
	"customerID":"$ID",
	"AddonSpsLevel":"0000",
	"AddonPatchLevel":"0000",
	"TargetVectorID":"",
	"repositories":[
		{	"name":"/ITAPC1/I_CURRENCZ",
			"tag":"whatever",
			"branch":"",
			"commitID":"",
			"version":"1.0.0",
			"versionAAK":"0001",
			"PackageName":"SAPK-002AAINITAPC1",
			"PackageType":"AOI",
			"SpLevel":"0000",
			"PatchLevel":"0000",
			"PredecessorCommitID":"",
			"Status":"P",
			"Namespace":"/ITAPC1/",
			"SarXMLFilePath":""
		}
	]
}`

// ********************* Test uploadSarFiles *******************
func TestUploadSarFiles(t *testing.T) {
	t.Run("test uploadSarFiles - success", func(t *testing.T) {
		client := abaputils.ClientMock{
			Body: "dummy",
		}
		repositories, conn := setupRepos("exists", aakaas.PackageStatusPlanned, client)
		err := uploadSarFiles(repositories, conn, mockReader)
		assert.NoError(t, err)
	})
	t.Run("test uploadSarFiles - error due to missing file path", func(t *testing.T) {
		repositories, conn := setupRepos("", aakaas.PackageStatusPlanned, abaputils.ClientMock{})
		err := uploadSarFiles(repositories, conn, mockReader)
		assert.Error(t, err)
	})
	t.Run("test uploadSarFiles - error due to missing file", func(t *testing.T) {
		repositories, conn := setupRepos("does_not_exist", aakaas.PackageStatusPlanned, abaputils.ClientMock{})
		err := uploadSarFiles(repositories, conn, mockReader)
		assert.Error(t, err)
	})
	t.Run("test uploadSarFiles - error during upload", func(t *testing.T) {
		client := abaputils.ClientMock{
			Body:  "ErrorBody",
			Error: errors.New("Failure during upload of SAR file"),
		}
		repositories, conn := setupRepos("exists", aakaas.PackageStatusPlanned, client)
		err := uploadSarFiles(repositories, conn, mockReader)
		assert.Error(t, err)
	})
}

// ********************* Test registerPackages *******************
func TestRegisterPackages(t *testing.T) {
	t.Run("test registerPackages - planned", func(t *testing.T) {
		client := abaputils.ClientMock{
			Body: responseRegisterPackagesPost,
		}
		repositories, conn := setupRepos("Filepath", aakaas.PackageStatusPlanned, client)
		repos, err := registerPackages(repositories, conn)
		assert.NoError(t, err)
		assert.Equal(t, string(aakaas.PackageStatusLocked), repos[0].Status)
	})
	t.Run("test registerPackages - released", func(t *testing.T) {
		repositories, conn := setupRepos("Filepath", aakaas.PackageStatusReleased, abaputils.ClientMock{})
		repos, err := registerPackages(repositories, conn)
		assert.NoError(t, err)
		assert.Equal(t, string(aakaas.PackageStatusReleased), repos[0].Status)
	})
	t.Run("test registerPackages - with error", func(t *testing.T) {
		client := abaputils.ClientMock{
			Body:  "ErrorBody",
			Error: errors.New("Failure during registration"),
		}
		repositories, conn := setupRepos("Filepath", aakaas.PackageStatusPlanned, client)
		repos, err := registerPackages(repositories, conn)
		assert.Error(t, err)
		assert.Equal(t, string(aakaas.PackageStatusPlanned), repos[0].Status)
	})
}

// ********************* Test Setup *******************
func setupRepos(filePath string, status aakaas.PackageStatus, cl abaputils.ClientMock) ([]abaputils.Repository, abapbuild.Connector) {
	repositories := []abaputils.Repository{
		{
			Name:           "/DRNMSPC/COMP01",
			VersionYAML:    "1.0.0",
			PackageName:    "SAPK-001AAINDRNMSPC",
			Status:         string(status),
			SarXMLFilePath: filePath,
		},
	}
	conn := new(abapbuild.Connector)
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

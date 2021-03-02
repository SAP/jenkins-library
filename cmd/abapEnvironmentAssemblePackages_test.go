package cmd

import (
	"path/filepath"
	"testing"
	"time"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/stretchr/testify/assert"
)

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
}

func TestCheckIfFailedAndPrintLogs(t *testing.T) {
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

func TestStarting(t *testing.T) {
	t.Run("Run starting", func(t *testing.T) {
		client := &abapbuild.ClMock{
			Token: "MyToken",
		}
		conn := new(abapbuild.Connector)
		conn.Client = client
		conn.Header = make(map[string][]string)
		var repos []abaputils.Repository
		repo := abaputils.Repository{
			Name:        "RepoA",
			Version:     "0001",
			PackageName: "Package",
			PackageType: "AOI",
			SpLevel:     "0000",
			PatchLevel:  "0000",
			Status:      "P",
			Namespace:   "/DEMO/",
		}
		repos = append(repos, repo)
		repo.Status = "R"
		repos = append(repos, repo)

		builds, buildsAlreadyReleased, err := starting(repos, *conn, time.Duration(0*time.Second))
		assert.NoError(t, err)
		assert.Equal(t, 1, len(builds))
		assert.Equal(t, 1, len(buildsAlreadyReleased))
		assert.Equal(t, abapbuild.Accepted, builds[0].build.RunState)
		assert.Equal(t, abapbuild.RunState(""), buildsAlreadyReleased[0].build.RunState)
	})
}

func TestStartingInvalidInput(t *testing.T) {
	t.Run("Run starting", func(t *testing.T) {
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
		_, _, err := starting(repos, *conn, time.Duration(0*time.Second))
		assert.Error(t, err)
	})
}

func TestPolling(t *testing.T) {
	t.Run("Run polling", func(t *testing.T) {
		var repo abaputils.Repository
		b := testSetup(&abapbuild.ClMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		var buildsWithRepo []buildWithRepository
		bWR := buildWithRepository{
			build: b,
			repo:  repo,
		}
		buildsWithRepo = append(buildsWithRepo, bWR)
		timeout := time.Duration(600 * time.Second)
		pollInterval := time.Duration(1 * time.Second)
		err := polling(buildsWithRepo, timeout, pollInterval)
		assert.NoError(t, err)
		assert.Equal(t, abapbuild.Finished, buildsWithRepo[0].build.RunState)
	})
}

func TestDownloadSARXML(t *testing.T) {
	t.Run("Run downloadSARXML", func(t *testing.T) {
		var repo abaputils.Repository
		b := testSetup(&abapbuild.ClMock{}, "ABIFNLDCSQPOVMXK4DNPBDRW2M")
		var buildsWithRepo []buildWithRepository
		bWR := buildWithRepository{
			build: b,
			repo:  repo,
		}
		buildsWithRepo = append(buildsWithRepo, bWR)
		repos, err := downloadSARXML(buildsWithRepo)
		assert.NoError(t, err)
		downloadPath := filepath.Join(GeneralConfig.EnvRootPath, "commonPipelineEnvironment", "abap", "SAPK-001AAINITAPC1.SAR")
		assert.Equal(t, downloadPath, repos[0].SarXMLFilePath)
	})
}

func TestStep(t *testing.T) {
	t.Run("abapEnvironmentAssemblePackages: nothing to do", func(t *testing.T) {

		config := &abapEnvironmentAssemblePackagesOptions{
			AddonDescriptor:             cpeAbapAddonDescriptorPackageLocked,
			MaxRuntimeInMinutes:         1,
			PollIntervalsInMilliseconds: 1,
		}

		autils := &abaputils.AUtilsMock{}
		client := abapbuild.GetBuildMockClient()
		cpe := &abapEnvironmentAssemblePackagesCommonPipelineEnvironment{}

		err := runAbapEnvironmentAssemblePackages(config, nil, autils, &client, cpe)
		assert.NoError(t, err)
	})
	t.Run("abapEnvironmentAssemblePackages: build", func(t *testing.T) {

		config := &abapEnvironmentAssemblePackagesOptions{
			AddonDescriptor:             cpeAbapAddonDescriptorPackageReserved,
			MaxRuntimeInMinutes:         1,
			PollIntervalsInMilliseconds: 1,
		}

		autils := &abaputils.AUtilsMock{
			ReturnedConnectionDetailsHTTP: abaputils.ConnectionDetailsHTTP{
				URL: `/sap/opu/odata/BUILD/CORE_SRV`,
			},
		}

		client := abapbuild.GetBuildMockClient()

		cpe := &abapEnvironmentAssemblePackagesCommonPipelineEnvironment{}

		err := runAbapEnvironmentAssemblePackages(config, nil, autils, &client, cpe)
		assert.NoError(t, err)
		assert.Contains(t, cpe.abap.addonDescriptor, `SAPK-001AAINITAPC1.SAR`)
	})
}

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

//go:build unit

package cmd

import (
	"testing"
	"time"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
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
		aD := abaputils.AddonDescriptor{
			Repositories: []abaputils.Repository{
				{
					Name:   "RepoA",
					Status: "P",
				},
			},
		}
		builds, err := executeBuilds(&aD, *conn, time.Duration(0*time.Second), time.Duration(1*time.Millisecond), "")
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
		}

		err := runAbapEnvironmentAssemblePackages(config, autils, &mock.FilesMock{}, &client, cpe)
		assert.NoError(t, err)
		assert.NotContains(t, cpe.abap.addonDescriptor, `"InBuildScope"`)
	})
	t.Run("abapEnvironmentAssemblePackages: build", func(t *testing.T) {
		config := &abapEnvironmentAssemblePackagesOptions{
			AddonDescriptor:             cpeAbapAddonDescriptorPackageReserved,
			MaxRuntimeInMinutes:         1,
			PollIntervalsInMilliseconds: 1,
		}

		err := runAbapEnvironmentAssemblePackages(config, autils, &mock.FilesMock{}, &client, cpe)
		assert.NoError(t, err)
		assert.Contains(t, cpe.abap.addonDescriptor, `SAPK-001AAINITAPC1.SAR`)
		assert.Contains(t, cpe.abap.addonDescriptor, `"InBuildScope":true`)
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

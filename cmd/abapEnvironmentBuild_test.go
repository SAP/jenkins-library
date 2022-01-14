package cmd

import (
	"testing"
	"time"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/stretchr/testify/assert"
)

type abapEnvironmentBuildMockUtils struct {
	*mock.ExecMockRunner
	*abapbuild.MockClient
}

func newAbapEnvironmentBuildTestsUtils() abapEnvironmentBuildUtils {
	mC := abapbuild.GetBuildMockClient()
	utils := abapEnvironmentBuildMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		MockClient:     &mC,
	}
	return &utils
}
func (mB abapEnvironmentBuildMockUtils) PersistReportsAndLinks(stepName, workspace string, reports, links []piperutils.Path) {
}
func (mB abapEnvironmentBuildMockUtils) GetAbapCommunicationArrangementInfo(options abaputils.AbapEnvironmentOptions, oDataURL string) (abaputils.ConnectionDetailsHTTP, error) {
	var cd abaputils.ConnectionDetailsHTTP
	cd.URL = "/sap/opu/odata/BUILD/CORE_SRV"
	return cd, nil
}
func (mB abapEnvironmentBuildMockUtils) GetPollIntervall() time.Duration {
	return 1 * time.Microsecond
}

func (mB abapEnvironmentBuildMockUtils) getMaxRuntime() time.Duration {
	return 1 * time.Second
}
func (mB abapEnvironmentBuildMockUtils) getPollingIntervall() time.Duration {
	return 1 * time.Microsecond
}

func TestRunAbapEnvironmentBuild(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		cpe := abapEnvironmentBuildCommonPipelineEnvironment{}
		config := abapEnvironmentBuildOptions{}
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		config.DownloadAllResultFiles = true
		config.PublishAllDownloadedResultFiles = true
		utils := newAbapEnvironmentBuildTestsUtils()
		// test
		err := runAbapEnvironmentBuild(&config, nil, &utils, &cpe)
		// assert
		finalValues := `[{"value_id":"PHASE","value":"AUNIT"},{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"AunitValue1"},{"value_id":"MyId2","value":"AunitValue2"},{"value_id":"BUILD_FRAMEWORK_MODE","value":"P"}]`
		assert.NoError(t, err)
		assert.Equal(t, finalValues, cpe.abap.buildValues)
	})

	t.Run("happy path, download only one", func(t *testing.T) {
		t.Parallel()
		// init
		cpe := abapEnvironmentBuildCommonPipelineEnvironment{}
		config := abapEnvironmentBuildOptions{}
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		config.DownloadResultFilenames = []string{"SAR_XML"}
		config.PublishResultFilenames = []string{"SAR_XML"}
		utils := newAbapEnvironmentBuildTestsUtils()
		// test
		err := runAbapEnvironmentBuild(&config, nil, &utils, &cpe)
		// assert
		assert.NoError(t, err)
	})

	t.Run("happy path, use AddonDescriptor", func(t *testing.T) {
		t.Parallel()
		//TODO alles ändern
		// init
		cpe := abapEnvironmentBuildCommonPipelineEnvironment{}
		config := abapEnvironmentBuildOptions{}
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		config.DownloadAllResultFiles = true
		config.PublishAllDownloadedResultFiles = true
		utils := newAbapEnvironmentBuildTestsUtils()
		// test
		err := runAbapEnvironmentBuild(&config, nil, &utils, &cpe)
		// assert
		finalValues := `[{"value_id":"PHASE","value":"AUNIT"},{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"AunitValue1"},{"value_id":"MyId2","value":"AunitValue2"},{"value_id":"BUILD_FRAMEWORK_MODE","value":"P"}]`
		assert.NoError(t, err)
		assert.Equal(t, finalValues, cpe.abap.buildValues)
	})

	t.Run("error path, try to publish file, which was not downloaded", func(t *testing.T) {
		t.Parallel()
		// init
		cpe := abapEnvironmentBuildCommonPipelineEnvironment{}
		config := abapEnvironmentBuildOptions{}
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		config.DownloadResultFilenames = []string{"DELIVERY_LOGS.ZIP"}
		config.PublishResultFilenames = []string{"SAR_XML"}
		utils := newAbapEnvironmentBuildTestsUtils()
		// test
		err := runAbapEnvironmentBuild(&config, nil, &utils, &cpe)
		// assert
		assert.Error(t, err)
	})

	t.Run("error path, try to download file which does not exist", func(t *testing.T) {
		t.Parallel()
		// init
		cpe := abapEnvironmentBuildCommonPipelineEnvironment{}
		config := abapEnvironmentBuildOptions{}
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		config.DownloadResultFilenames = []string{"DOES_NOT_EXIST"}
		config.PublishAllDownloadedResultFiles = true
		utils := newAbapEnvironmentBuildTestsUtils()
		// test
		err := runAbapEnvironmentBuild(&config, nil, &utils, &cpe)
		// assert
		assert.Error(t, err)
	})
}

func TestGenerateValues(t *testing.T) {
	t.Parallel()
	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentBuildOptions{}
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		config.CpeValues = `[{"value_id":"PHASE","value":"AUNIT"},{"value_id":"PACKAGES","value":"CPE_PACKAGE"},{"value_id":"MyId2","value":"Value2"}]`
		// test
		values, err := generateValues(&config, []abapbuild.Value{})
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 3, len(values.Values))
		assert.Equal(t, "/BUILD/AUNIT_DUMMY_TESTS", values.Values[0].Value)
		assert.Equal(t, "Value1", values.Values[1].Value)
		assert.Equal(t, "Value2", values.Values[2].Value)
	})
	t.Run("happy path, use addonDescriptor", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentBuildOptions{}
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		config.CpeValues = `[{"value_id":"PHASE","value":"AUNIT"},{"value_id":"PACKAGES","value":"CPE_PACKAGE"},{"value_id": "Status","value":"R"},{"value_id":"MyId2","value":"Value2"}]`
		config.AddonDescriptor = addonDescriptor
		config.UseAddonDescriptor = true
		// test
		valuesAddonDescriptor, err := evaluateAddonDescriptor(&config)
		values0, err := generateValues(&config, valuesAddonDescriptor[0].Values)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 19, len(values0.Values))
		assert.Equal(t, "/BUILD/AUNIT_DUMMY_TESTS", values0.Values[0].Value)
		assert.Equal(t, "Value1", values0.Values[1].Value)
		assert.Equal(t, "/ITAPC1/I_CURRENCY", values0.Values[2].Value)
		assert.Equal(t, "Value2", values0.Values[18].Value)
	})
	t.Run("error path - duplicate in config", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentBuildOptions{}
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"},{"value_id":"MyId1","value":"Value1"}]`
		// test
		values, err := generateValues(&config, []abapbuild.Value{})
		// assert
		assert.Error(t, err)
		assert.Equal(t, 0, len(values.Values))
	})
	t.Run("error path - bad formating in config.Values", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentBuildOptions{}
		config.Values = `[{"task_id":"PACKAGES","task":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		// test
		_, err := generateValues(&config, []abapbuild.Value{})
		// assert
		assert.Error(t, err)
	})
}

func TestEvaluateAddonDescriptor(t *testing.T) {
	//global init
	config := abapEnvironmentBuildOptions{}
	config.AddonDescriptor = addonDescriptor
	config.UseAddonDescriptor = true
	t.Run("Find one", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = `[{"field":"Name","operator":"==","value":"/ITAPC1/I_CURRENCY"},{"field":"Status","operator":"!=","value":"R"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 1, len(values))
	})
	t.Run("Find both", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = `[{"field":"PackageType","operator":"==","value":"AOI"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 2, len(values))
	})
	t.Run("Find none", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = `[{"field":"PackageType","operator":"!=","value":"AOI"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 0, len(values))
	})
	t.Run("No condition", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = ``
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 2, len(values))
	})
	t.Run("Wrong fieldname", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = `[{"fieldxxx":"Name","operator":"==","value":"/ITAPC1/I_CURRENCY"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.Error(t, err)
		assert.Equal(t, 0, len(values))
	})
	t.Run("Wrong value fieldname", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = `[{"field":"Name","operator":"==","valuexxxx":"/ITAPC1/I_CURRENCY"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.Error(t, err)
		assert.Equal(t, 0, len(values))
	})
	t.Run("Wrong operator", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = `[{"field":"Name","operator":"()","value":"/ITAPC1/I_CURRENCY"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.Error(t, err)
		assert.Equal(t, 0, len(values))
	})
}

var addonDescriptor = `{
	"addonProduct":"/ITAPC1/I_CURRENCZPRODUCT",
	"addonVersion":"1.0.0",
	"addonVersionAAK":"0001",
	"AddonSpsLevel":"0000",
	"AddonPatchLevel":"0000",
	"TargetVectorID":"",
	"repositories":[
	  {
		"name":"/ITAPC1/I_CURRENCY",
		"tag":"",
		"branch":"v1.0.0",
		"commitID":"1cb96a82",
		"version":"1.0.0",
		"versionAAK":"0001",
		"PackageName":"SAPK-004AAINITAPC1",
		"PackageType":"AOI",
		"SpLevel":"0000",
		"PatchLevel":"0000",
		"PredecessorCommitID":"",
		"Status":"P",
		"Namespace":"/ITAPC1/",
		"SarXMLFilePath":"",
		"languages":null,
		"InBuildScope":false
		},
	  {
		"name":"/ITAPC1/I_FLIGHT",
		"tag":"",
		"branch":"v2.0.0",
		"commitID":"5b87c03",
		"version":"2.0.0",
		"versionAAK":"0001",
		"PackageName":"SAPK-005AAINITAPC2",
		"PackageType":"AOI",
		"SpLevel":"0000",
		"PatchLevel":"0000",
		"PredecessorCommitID":"",
		"Status":"R",
		"Namespace":"/ITAPC1/",
		"SarXMLFilePath":"",
		"languages":null,
		"InBuildScope":false
		}
	  ]
	}`

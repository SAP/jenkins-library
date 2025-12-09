//go:build unit
// +build unit

package cmd

import (
	"encoding/json"
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
	*mock.FilesMock
}

func newAbapEnvironmentBuildTestsUtils() abapEnvironmentBuildUtils {
	mC := abapbuild.GetBuildMockClientToRun2Times()
	utils := abapEnvironmentBuildMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		MockClient:     &mC,
		FilesMock:      &mock.FilesMock{},
	}
	return &utils
}

func newAbapEnvironmentBuildTestsUtilsWithClient() abapEnvironmentBuildUtils {
	mC := abapbuild.GetBuildMockClientWithClient()
	utils := abapEnvironmentBuildMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		MockClient:     &mC,
		FilesMock:      &mock.FilesMock{},
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

func (mB abapEnvironmentBuildMockUtils) publish() {
}

func (mB abapEnvironmentBuildMockUtils) GetPollIntervall() time.Duration {
	return 1 * time.Microsecond
}

func (mB abapEnvironmentBuildMockUtils) getMaxRuntime() time.Duration {
	return 1 * time.Second
}
func (mB abapEnvironmentBuildMockUtils) getPollingInterval() time.Duration {
	return 1 * time.Microsecond
}

func TestRunAbapEnvironmentBuild(t *testing.T) {
	t.Parallel()
	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		cpe := abapEnvironmentBuildCommonPipelineEnvironment{}
		config := abapEnvironmentBuildOptions{}
		config.AddonDescriptor = addonDescriptor
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		config.DownloadAllResultFiles = true
		config.PublishAllDownloadedResultFiles = true
		utils := newAbapEnvironmentBuildTestsUtils()
		// test
		err := runAbapEnvironmentBuild(&config, utils, &cpe)
		// assert
		finalValues := `[{"value_id":"PHASE","value":"AUNIT"},{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"AunitValue1"},{"value_id":"MyId2","value":"AunitValue2"},{"value_id":"BUILD_FRAMEWORK_MODE","value":"P"}]`
		assert.NoError(t, err)
		assert.Equal(t, finalValues, cpe.abap.buildValues)
	})

	t.Run("happy path, use client", func(t *testing.T) {
		t.Parallel()
		// init
		cpe := abapEnvironmentBuildCommonPipelineEnvironment{}
		config := abapEnvironmentBuildOptions{}
		config.AddonDescriptor = addonDescriptor
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		config.AbapSourceClient = "001"
		utils := newAbapEnvironmentBuildTestsUtilsWithClient()
		// test
		err := runAbapEnvironmentBuild(&config, utils, &cpe)
		// assert
		finalValues := `[{"value_id":"PHASE","value":"AUNIT"},{"value_id":"SUN","value":"SUMMER"}]`
		assert.NoError(t, err)
		assert.Equal(t, finalValues, cpe.abap.buildValues)
	})

	t.Run("happy path, download only one", func(t *testing.T) {
		t.Parallel()
		// init
		cpe := abapEnvironmentBuildCommonPipelineEnvironment{}
		config := abapEnvironmentBuildOptions{}
		config.AddonDescriptor = addonDescriptor
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		config.DownloadResultFilenames = []string{"SAR_XML"}
		config.PublishResultFilenames = []string{"SAR_XML"}
		utils := newAbapEnvironmentBuildTestsUtils()
		// test
		err := runAbapEnvironmentBuild(&config, utils, &cpe)
		// assert
		assert.NoError(t, err)
	})

	t.Run("happy path, use AddonDescriptor", func(t *testing.T) {
		t.Parallel()
		// init
		expectedValueList := []abapbuild.Value{}
		recordedValueList := []abapbuild.Value{}
		cpe := abapEnvironmentBuildCommonPipelineEnvironment{}
		config := abapEnvironmentBuildOptions{}
		config.AddonDescriptor = addonDescriptor
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"}]`
		//config.ConditionOnAddonDescriptor = `[{"field":"PackageType","operator":"!=","value":"AOI"}]`
		//The mock client returns MyId1 & and MyId2 therefore we rename it to these values so that we can remove it from the output
		config.UseFieldsOfAddonDescriptor = `[{"use":"Name","renameTo":"MyId1"},{"use":"Status","renameTo":"MyId2"}]`
		utils := newAbapEnvironmentBuildTestsUtils()
		// test
		err := runAbapEnvironmentBuild(&config, utils, &cpe)
		// assert
		finalValues := `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"}]`
		err = json.Unmarshal([]byte(finalValues), &expectedValueList)
		assert.NoError(t, err)
		err = json.Unmarshal([]byte(cpe.abap.buildValues), &recordedValueList)
		assert.NoError(t, err)
		assert.NoError(t, err)
		assert.ElementsMatch(t, expectedValueList, recordedValueList)
	})

	t.Run("error path, try to publish file, which was not downloaded", func(t *testing.T) {
		t.Parallel()
		// init
		cpe := abapEnvironmentBuildCommonPipelineEnvironment{}
		config := abapEnvironmentBuildOptions{}
		config.AddonDescriptor = addonDescriptor
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		config.DownloadResultFilenames = []string{"DELIVERY_LOGS.ZIP"}
		config.PublishResultFilenames = []string{"SAR_XML"}
		utils := newAbapEnvironmentBuildTestsUtils()
		// test
		err := runAbapEnvironmentBuild(&config, utils, &cpe)
		// assert
		assert.Error(t, err)
	})

	t.Run("error path, try to download file which does not exist", func(t *testing.T) {
		t.Parallel()
		// init
		cpe := abapEnvironmentBuildCommonPipelineEnvironment{}
		config := abapEnvironmentBuildOptions{}
		config.AddonDescriptor = addonDescriptor
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		config.DownloadResultFilenames = []string{"DOES_NOT_EXIST"}
		config.PublishAllDownloadedResultFiles = true
		utils := newAbapEnvironmentBuildTestsUtils()
		// test
		err := runAbapEnvironmentBuild(&config, utils, &cpe)
		// assert
		assert.Error(t, err)
	})
}

func TestGenerateValues(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// init
		config := abapEnvironmentBuildOptions{}
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		config.CpeValues = `[{"value_id":"PHASE","value":"AUNIT"},{"value_id":"PACKAGES","value":"CPE_PACKAGE"},{"value_id":"MyId2","value":"Value2"}]`
		// test
		values, err := generateValuesOnlyFromConfig(&config)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 3, len(values))

		//values generated by map -> therefore we can not just iterate over the generated values and check them, as they are in random order
		checkMap := map[string]string{
			"PACKAGES": "/BUILD/AUNIT_DUMMY_TESTS",
			"MyId1":    "Value1",
			"MyId2":    "Value2",
		}
		for _, value := range values {
			checkValue, present := checkMap[value.ValueID]
			assert.True(t, present)
			assert.Equal(t, checkValue, value.Value)
		}
	})
	t.Run("happy path, use addonDescriptor", func(t *testing.T) {
		// init
		config := abapEnvironmentBuildOptions{}
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		config.CpeValues = `[{"value_id":"PHASE","value":"AUNIT"},{"value_id":"PACKAGES","value":"CPE_PACKAGE"},{"value_id": "Status","value":"R"},{"value_id":"MyId2","value":"Value2"}]`
		config.AddonDescriptor = addonDescriptor
		config.UseFieldsOfAddonDescriptor = `[{"use":"Name","renameTo":"SWC"},{"use":"Status","renameTo":"Status"}]`
		// test
		valuesAddonDescriptor, err := evaluateAddonDescriptor(&config)
		values0, err := generateValuesWithAddonDescriptor(&config, valuesAddonDescriptor[0])
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 5, len(values0))
		//values generated by map -> therefore we can not just iterate over the generated values and check them, as they are in random order
		checkMap := map[string]string{
			"PACKAGES": "/BUILD/AUNIT_DUMMY_TESTS",
			"MyId1":    "Value1",
			"MyId2":    "Value2",
			"SWC":      "/ITAPC1/I_CURRENCY",
			"Status":   "P",
		}
		for _, value := range values0 {
			checkValue, present := checkMap[value.ValueID]
			assert.True(t, present)
			assert.Equal(t, checkValue, value.Value)
		}
	})
	t.Run("error path, use addonDescriptor already in config", func(t *testing.T) {
		// init
		config := abapEnvironmentBuildOptions{}
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"},{"value_id":"Branch","value":"configBranch"}]`
		config.CpeValues = `[{"value_id":"PHASE","value":"AUNIT"},{"value_id":"PACKAGES","value":"CPE_PACKAGE"},{"value_id": "Status","value":"R"},{"value_id":"MyId2","value":"Value2"}]`
		config.AddonDescriptor = addonDescriptor
		config.UseFieldsOfAddonDescriptor = `[{"use":"Name","renameTo":"SWC"},{"use":"Status","renameTo":"Status"},{"use":"Branch","renameTo":"Branch"}]`
		// test
		valuesAddonDescriptor, err := evaluateAddonDescriptor(&config)
		values0, err := generateValuesWithAddonDescriptor(&config, valuesAddonDescriptor[0])
		// assert
		assert.Error(t, err)
		assert.Equal(t, 0, len(values0))
	})
	t.Run("error path - duplicate in config", func(t *testing.T) {
		// init
		config := abapEnvironmentBuildOptions{}
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"},{"value_id":"MyId1","value":"Value1"}]`
		// test
		values, err := generateValuesOnlyFromConfig(&config)
		// assert
		assert.Error(t, err)
		assert.Equal(t, 0, len(values))
	})
	t.Run("error path - bad formating in config.Values", func(t *testing.T) {
		// init
		config := abapEnvironmentBuildOptions{}
		config.Values = `[{"task_id":"PACKAGES","task":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"}]`
		// test
		_, err := generateValuesOnlyFromConfig(&config)
		// assert
		assert.Error(t, err)
	})
}

func TestEvaluateAddonDescriptor(t *testing.T) {
	//global init
	config := abapEnvironmentBuildOptions{}
	config.AddonDescriptor = addonDescriptor
	t.Run("Find one", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = `[{"field":"Name","operator":"==","value":"/ITAPC1/I_CURRENCY"},{"field":"Status","operator":"!=","value":"R"}]`
		config.UseFieldsOfAddonDescriptor = `[{"use":"Name","renameTo":"SWC"},{"use":"Status","renameTo":"Status"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 1, len(values))
		assert.Equal(t, 2, len(values[0]))
	})
	t.Run("Find both", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = `[{"field":"PackageType","operator":"==","value":"AOI"}]`
		config.UseFieldsOfAddonDescriptor = `[{"use":"Name","renameTo":"SWC"},{"use":"Status","renameTo":"Status"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 2, len(values))
		assert.Equal(t, 2, len(values[0]))
		assert.Equal(t, 2, len(values[1]))
	})
	t.Run("Find none", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = `[{"field":"PackageType","operator":"!=","value":"AOI"}]`
		config.UseFieldsOfAddonDescriptor = `[{"use":"Name","renameTo":"SWC"},{"use":"Status","renameTo":"Status"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 0, len(values))
	})
	t.Run("No condition", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = ``
		config.UseFieldsOfAddonDescriptor = `[{"use":"Name","renameTo":"SWC"},{"use":"Status","renameTo":"Status"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 2, len(values))
		assert.Equal(t, 2, len(values[0]))
		assert.Equal(t, 2, len(values[1]))
	})
	t.Run("No UseFields", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = ``
		config.UseFieldsOfAddonDescriptor = ``
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 0, len(values))
	})
	t.Run("Wrong fieldname", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = `[{"fieldxxx":"Name","operator":"==","value":"/ITAPC1/I_CURRENCY"}]`
		config.UseFieldsOfAddonDescriptor = `[{"use":"Name","renameTo":"SWC"},{"use":"Status","renameTo":"Status"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.Error(t, err)
		assert.Equal(t, 0, len(values))
	})
	t.Run("Wrong value fieldname", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = `[{"field":"Name","operator":"==","valuexxxx":"/ITAPC1/I_CURRENCY"}]`
		config.UseFieldsOfAddonDescriptor = `[{"use":"Name","renameTo":"SWC"},{"use":"Status","renameTo":"Status"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.Error(t, err)
		assert.Equal(t, 0, len(values))
	})
	t.Run("Field which does not exist", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = `[{"field":"DoesNotExist","operator":"==","valuexxxx":"/ITAPC1/I_CURRENCY"}]`
		config.UseFieldsOfAddonDescriptor = `[{"use":"Name","renameTo":"SWC"},{"use":"Status","renameTo":"Status"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.Error(t, err)
		assert.Equal(t, 0, len(values))
	})
	t.Run("Wrong operator", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = `[{"field":"Name","operator":"()","value":"/ITAPC1/I_CURRENCY"}]`
		config.UseFieldsOfAddonDescriptor = `[{"use":"Name","renameTo":"SWC"},{"use":"Status","renameTo":"Status"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.Error(t, err)
		assert.Equal(t, 0, len(values))
	})
	t.Run("Test UseFieldsOfAddonDescriptor: Bad formatting use", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = ``
		config.UseFieldsOfAddonDescriptor = `[{"usexxx":"Name","renameTo":"SWC"},{"use":"Status","renameTo":"Status"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.Error(t, err)
		assert.Equal(t, 0, len(values))
	})
	t.Run("Test UseFieldsOfAddonDescriptor: Bad formatting as", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = ``
		config.UseFieldsOfAddonDescriptor = `[{"use":"Name","renameTo":"SWC"},{"use":"Status","asxxx":"Status"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.Error(t, err)
		assert.Equal(t, 0, len(values))
	})
	t.Run("Test UseFieldsOfAddonDescriptor: use which does not exist", func(t *testing.T) {
		// init
		config.ConditionOnAddonDescriptor = ``
		config.UseFieldsOfAddonDescriptor = `[{"use":"DoesNotExist","renameTo":"SWC"},{"use":"Status","renameTo":"Status"}]`
		// test
		values, err := evaluateAddonDescriptor(&config)
		// assert
		assert.Error(t, err)
		assert.Equal(t, 0, len(values))
	})
}

func TestValues2String(t *testing.T) {
	t.Run("dito", func(t *testing.T) {
		var myValues []abapbuild.Value
		myValues = append(myValues, abapbuild.Value{ValueID: "Name", Value: "Hugo"})
		myValues = append(myValues, abapbuild.Value{ValueID: "Age", Value: "43"})
		myValues = append(myValues, abapbuild.Value{ValueID: "Hight", Value: "17cm"})
		myString := values2string(myValues)
		assert.Equal(t, "Name = Hugo; Age = 43; Hight = 17cm", myString)
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

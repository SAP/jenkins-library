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
		values, err := generateValues(&config)
		// assert
		assert.NoError(t, err)
		assert.Equal(t, 3, len(values.Values))
		assert.Equal(t, "/BUILD/AUNIT_DUMMY_TESTS", values.Values[0].Value)
		assert.Equal(t, "Value1", values.Values[1].Value)
		assert.Equal(t, "Value2", values.Values[2].Value)
	})
	t.Run("error path - duplicate in config", func(t *testing.T) {
		t.Parallel()
		// init
		config := abapEnvironmentBuildOptions{}
		config.Values = `[{"value_id":"PACKAGES","value":"/BUILD/AUNIT_DUMMY_TESTS"},{"value_id":"MyId1","value":"Value1"},{"value_id":"MyId1","value":"Value1"}]`
		// test
		values, err := generateValues(&config)
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
		_, err := generateValues(&config)
		// assert
		assert.Error(t, err)
	})
}

//go:build unit

package versioning

import (
	"fmt"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
)

func TestHelmChartInit(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		chartMetadata := chart.Metadata{Version: "1.2.3"}
		content, err := yaml.Marshal(chartMetadata)
		assert.NoError(t, err)

		fileUtils := newVersioningMockUtils()
		fileUtils.AddFile("testchart/Chart.yaml", content)

		helmChart := HelmChart{
			utils: fileUtils,
		}

		err = helmChart.init()

		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", helmChart.metadata.Version)
	})

	t.Run("success case - with chart path", func(t *testing.T) {
		chartMetadata := chart.Metadata{Version: "1.2.3"}
		content, err := yaml.Marshal(chartMetadata)
		assert.NoError(t, err)

		fileUtils := newVersioningMockUtils()
		fileUtils.AddFile("chart1/Chart.yaml", []byte(""))
		fileUtils.AddFile("chart2/Chart.yaml", content)
		fileUtils.AddFile("chart3/Chart.yaml", []byte(""))

		helmChart := HelmChart{
			path:  "chart2/Chart.yaml",
			utils: fileUtils,
		}

		err = helmChart.init()

		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", helmChart.metadata.Version)
	})

	t.Run("error case - init failed with missing utils", func(t *testing.T) {
		helmChart := HelmChart{
			path: "chart2/Chart.yaml",
		}

		err := helmChart.init()
		assert.EqualError(t, err, "no file utils provided")
	})

	t.Run("error case - init failed with missing chart", func(t *testing.T) {
		fileUtils := newVersioningMockUtils()

		helmChart := HelmChart{
			utils: fileUtils,
		}

		err := helmChart.init()
		assert.EqualError(t, err, "failed to find a helm chart file")
	})

	t.Run("error case - failed reading file", func(t *testing.T) {
		fileUtils := newVersioningMockUtils()
		fileUtils.FileReadErrors = map[string]error{"testchart/Chart.yaml": fmt.Errorf("read error")}

		helmChart := HelmChart{
			utils: fileUtils,
			path:  "testchart/Chart.yaml",
		}

		err := helmChart.init()
		assert.EqualError(t, err, "failed to read file 'testchart/Chart.yaml': read error")
	})

	t.Run("error case - chart invalid", func(t *testing.T) {
		fileUtils := newVersioningMockUtils()
		fileUtils.AddFile("testchart/Chart.yaml", []byte("{"))

		helmChart := HelmChart{
			utils: fileUtils,
			path:  "testchart/Chart.yaml",
		}

		err := helmChart.init()
		assert.Contains(t, fmt.Sprint(err), "helm chart content invalid 'testchart/Chart.yaml'")
	})
}

func TestHelmChartVersioningScheme(t *testing.T) {
	helmChart := HelmChart{}
	assert.Equal(t, "semver2", helmChart.VersioningScheme())
}

func TestHelmChartGetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		chartMetadata := chart.Metadata{Version: "1.2.3"}
		content, err := yaml.Marshal(chartMetadata)
		assert.NoError(t, err)

		fileUtils := newVersioningMockUtils()
		fileUtils.AddFile("testchart/Chart.yaml", content)

		helmChart := HelmChart{
			utils: fileUtils,
		}

		version, err := helmChart.GetVersion()
		assert.NoError(t, err)
		assert.Equal(t, "1.2.3", version)
	})

	t.Run("error case - init failed", func(t *testing.T) {
		fileUtils := newVersioningMockUtils()

		helmChart := HelmChart{
			utils: fileUtils,
		}

		_, err := helmChart.GetVersion()
		assert.Contains(t, fmt.Sprint(err), "failed to init helm chart versioning:")
	})
}

func TestHelmChartSetVersion(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		fileUtils := newVersioningMockUtils()

		helmChart := HelmChart{
			utils:    fileUtils,
			path:     "testchart/Chart.yaml",
			metadata: chart.Metadata{Version: "1.2.3"},
		}

		err := helmChart.SetVersion("1.2.4")
		assert.NoError(t, err)
		assert.Equal(t, "1.2.4", helmChart.metadata.Version)

		fileContent, err := fileUtils.FileRead("testchart/Chart.yaml")
		assert.NoError(t, err)
		assert.Contains(t, string(fileContent), "version: 1.2.4")
	})

	t.Run("success case - update app version", func(t *testing.T) {
		fileUtils := newVersioningMockUtils()

		helmChart := HelmChart{
			utils:            fileUtils,
			path:             "testchart/Chart.yaml",
			metadata:         chart.Metadata{Version: "1.2.3"},
			updateAppVersion: true,
		}

		err := helmChart.SetVersion("1.2.4")
		assert.NoError(t, err)
		assert.Equal(t, "1.2.4", helmChart.metadata.AppVersion)
	})

	t.Run("success case - update app version: '+' is being replaced with '_' (semver2)", func(t *testing.T) {
		fileUtils := newVersioningMockUtils()

		helmChart := HelmChart{
			utils:            fileUtils,
			path:             "testchart/Chart.yaml",
			metadata:         chart.Metadata{Version: "1.2.3"},
			updateAppVersion: true,
		}

		// '+' is being replaced with '_' since k8s does not allow a plus sign in labels
		err := helmChart.SetVersion("1.2.4-2022+123")
		assert.NoError(t, err)
		assert.Equal(t, "1.2.4-2022_123", helmChart.metadata.AppVersion)
	})

	t.Run("error case - init failed with missing chart", func(t *testing.T) {
		fileUtils := newVersioningMockUtils()

		helmChart := HelmChart{
			utils: fileUtils,
		}

		err := helmChart.SetVersion("1.2.4")
		assert.Contains(t, fmt.Sprint(err), "failed to init helm chart versioning:")
	})

	t.Run("error case - failed to write chart", func(t *testing.T) {
		fileUtils := newVersioningMockUtils()
		fileUtils.FileWriteError = fmt.Errorf("write error")

		helmChart := HelmChart{
			path:     "testchart/Chart.yaml",
			utils:    fileUtils,
			metadata: chart.Metadata{Version: "1.2.3"},
		}

		err := helmChart.SetVersion("1.2.4")
		assert.EqualError(t, err, "failed to write file 'testchart/Chart.yaml': write error")
	})
}

func TestHelmChartGetCoordinates(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		fileUtils := newVersioningMockUtils()
		helmChart := HelmChart{
			utils:    fileUtils,
			path:     "testchart/Chart.yaml",
			metadata: chart.Metadata{Version: "1.2.3", Name: "myChart", Home: "myHome"},
		}

		coordinates, err := helmChart.GetCoordinates()
		assert.NoError(t, err)
		assert.Equal(t, Coordinates{GroupID: "myHome", ArtifactID: "myChart", Version: "1.2.3"}, coordinates)
	})

	t.Run("error case - version retrieval failed", func(t *testing.T) {
		fileUtils := newVersioningMockUtils()

		helmChart := HelmChart{
			utils: fileUtils,
		}

		_, err := helmChart.GetCoordinates()
		assert.Contains(t, fmt.Sprint(err), "failed to init helm chart versioning:")
	})
}

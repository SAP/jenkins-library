//go:build unit

package kubernetes

import (
	"errors"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestRunUtils(t *testing.T) {
	t.Run("Get container info", func(t *testing.T) {
		testTable := []struct {
			chartYamlFile          string
			dataChartYaml          string
			expectedChartName      string
			expectedPackageVersion string
			expectedError          error
			setFileReadError       bool
		}{
			{
				chartYamlFile:          "path/to/Chart.yaml",
				dataChartYaml:          "name: nginx-testChart\nversion: 1.3.5",
				expectedChartName:      "nginx-testChart",
				expectedPackageVersion: "1.3.5",
				expectedError:          nil,
				setFileReadError:       false,
			},
			{
				chartYamlFile:          "path/to/Chart.yaml",
				dataChartYaml:          "name: nginx-testChart\nversion: 1.3.5",
				expectedChartName:      "nginx-testChart",
				expectedPackageVersion: "1.3.5",
				expectedError:          errors.New("file couldn't read"),
				setFileReadError:       true,
			},
			{
				chartYamlFile:          "path/to/Chart.yaml",
				dataChartYaml:          "version: 1.3.5",
				expectedChartName:      "nginx-testChart",
				expectedPackageVersion: "1.3.5",
				expectedError:          errors.New("name not found in chart yaml file (or wrong type)"),
				setFileReadError:       false,
			},
			{
				chartYamlFile:          "path/to/Chart.yaml",
				dataChartYaml:          "name: nginx-testChart",
				expectedChartName:      "nginx-testChart",
				expectedPackageVersion: "1.3.5",
				expectedError:          errors.New("version not found in chart yaml file (or wrong type)"),
				setFileReadError:       false,
			},
			{
				chartYamlFile:          "path/to/Chart.yaml",
				dataChartYaml:          "name=nginx-testChart",
				expectedChartName:      "nginx-testChart",
				expectedPackageVersion: "1.3.5",
				expectedError:          errors.New("failed unmarshal"),
				setFileReadError:       false,
			},
		}

		for _, testCase := range testTable {
			utils := helmMockUtilsBundle{
				ExecMockRunner: &mock.ExecMockRunner{},
				FilesMock:      &mock.FilesMock{},
				HttpClientMock: &mock.HttpClientMock{
					FileUploads: map[string]string{},
				},
			}
			utils.AddFile(testCase.chartYamlFile, []byte(testCase.dataChartYaml))
			if testCase.setFileReadError {
				utils.FileReadErrors = map[string]error{testCase.chartYamlFile: testCase.expectedError}
			}
			nameChart, packageVersion, err := GetChartInfo(testCase.chartYamlFile, utils)
			if testCase.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedChartName, nameChart)
				assert.Equal(t, testCase.expectedPackageVersion, packageVersion)
			}

		}
	})
}
